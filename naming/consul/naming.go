package consul

import (
	x "X_IM"
	"X_IM/naming"
	"X_IM/pkg/logger"
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"sync"
	"time"
)

const (
	KeyProtocol  = "protocol"
	KeyHealthURL = "health_url"
)

// Watch 服务变化监听
type Watch struct {
	Service   string
	Callback  func([]x.ServiceRegistration)
	WaitIndex uint64
	Quit      chan struct{}
}

// Naming consul服务发现
type Naming struct {
	sync.RWMutex
	cli     *api.Client
	watches map[string]*Watch
}

// NewNaming 创建一个consul服务发现实例
func NewNaming(consulUrl string) (naming.Naming, error) {
	conf := api.DefaultConfig()
	conf.Address = consulUrl
	cli, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}
	nm := &Naming{
		cli:     cli,
		watches: make(map[string]*Watch, 1),
	}
	return nm, nil
}

// Find 服务发现，支持通过tag查询
func (n *Naming) Find(serviceName string, tags ...string) ([]x.ServiceRegistration, error) {
	services, _, err := n.load(serviceName, 0, tags...)
	if err != nil {
		return nil, err
	}
	return services, nil
}

// load waitIndex表示阻塞查询，Watch时被使用，0表示不阻塞
func (n *Naming) load(serviceName string, waitIndex uint64, tags ...string) ([]x.ServiceRegistration, *api.QueryMeta, error) {
	opts := &api.QueryOptions{
		UseCache: true,
		// 如果UseCache开启，那么MaxAge表示从缓存中获取的最大数据的时间
		// 如果超过这个时间，那么就会从consul中获取最新的数据
		MaxAge:    time.Minute,
		WaitIndex: waitIndex,
	}
	catalogServices, meta, err := n.cli.Catalog().ServiceMultipleTags(serviceName, tags, opts)
	if err != nil {
		return nil, meta, err
	}

	services := make([]x.ServiceRegistration, 0, len(catalogServices))
	for _, s := range catalogServices {
		if s.Checks.AggregatedStatus() != api.HealthPassing {
			logger.Infof("load service: id:%s name:%s %s:%d status:%s",
				s.ServiceID, s.ServiceName, s.ServiceAddress, s.ServicePort, s.Checks.AggregatedStatus())
			continue
		}
		services = append(services, &naming.DefaultService{
			ID:       s.ServiceID,
			Name:     s.ServiceName,
			Address:  s.ServiceAddress,
			Port:     s.ServicePort,
			Protocol: s.ServiceMeta[KeyProtocol],
			Tags:     s.ServiceTags,
			Meta:     s.ServiceMeta,
		})
	}
	logger.Infof("load service: %v, meta:%v", services, meta)
	return services, meta, nil
}
func (n *Naming) Register(s x.ServiceRegistration) error {
	reg := &api.AgentServiceRegistration{
		ID:      s.ServiceID(),
		Name:    s.ServiceName(),
		Address: s.PublicAddress(),
		Port:    s.PublicPort(),
		Tags:    s.GetTags(),
		Meta:    s.GetMeta(),
	}
	if reg.Meta == nil {
		reg.Meta = make(map[string]string)
	}
	reg.Meta[KeyProtocol] = s.GetProtocol()

	// consul健康检查
	healthURL := s.GetMeta()[KeyHealthURL]
	if healthURL != "" {
		check := new(api.AgentServiceCheck)
		check.CheckID = fmt.Sprintf("%s_normal", s.ServiceID())
		check.HTTP = healthURL
		check.Timeout = "1s" // http timeout
		check.Interval = "10s"
		// 服务故障20s后由Agent将其下线
		check.DeregisterCriticalServiceAfter = "20s"
		reg.Check = check
	}

	err := n.cli.Agent().ServiceRegister(reg)
	return err
}

func (n *Naming) Deregister(serviceID string) error {
	return n.cli.Agent().ServiceDeregister(serviceID)
}

func (n *Naming) Subscribe(serviceName string, callback func([]x.ServiceRegistration)) error {
	n.Lock()
	defer n.Unlock()
	if _, ok := n.watches[serviceName]; ok {
		return errors.New("service name has already been registered")
	}
	w := &Watch{
		Service:  serviceName,
		Callback: callback,
		Quit:     make(chan struct{}, 1),
	}
	n.watches[serviceName] = w
	// 等n.load返回后才结束
	go n.watch(w)
	return nil
}

// watch 监听服务变化
func (n *Naming) watch(w *Watch) {
	stopped := false

	var doWatch = func(service string, callback func([]x.ServiceRegistration)) {
		// load 阻塞式调用，直到有服务变化或者超时
		services, meta, err := n.load(service, w.WaitIndex)
		if err != nil {
			logger.Warn(err)
			return
		}
		select {
		// 通过 Unsubscribe 中的 close 关闭 channel
		// 此时接收数据不会阻塞，而是立即返回 channel 类型的零值nil
		case <-w.Quit:
			stopped = true
			logger.Infof("watch %s stopped", w.Service)
			return
		default:
		}

		w.WaitIndex = meta.LastIndex
		if callback != nil {
			callback(services)
		}
	}

	// 首次执行时不会callback，只是用来初始化WaitIndex
	// 所以这次执行n.load不会阻塞，而是返回meta.LastIndex
	doWatch(w.Service, nil)
	for !stopped {
		doWatch(w.Service, w.Callback)
	}
}

func (n *Naming) Unsubscribe(serviceName string) error {
	n.Lock()
	defer n.Unlock()
	w, ok := n.watches[serviceName]

	delete(n.watches, serviceName)
	if ok {
		close(w.Quit)
	}
	return nil
}
