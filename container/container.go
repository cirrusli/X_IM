package container

import (
	x "X_IM"
	"X_IM/logger"
	"X_IM/naming"
	"X_IM/tcp"
	"X_IM/wire/common"
	"X_IM/wire/pkt"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	stateUninitialized = iota
	stateInitialized
	stateStarted
	stateClosed
)
const (
	StateYoung = "young"
	StateAdult = "adult"
)

const (
	KeyServiceState = "service_state"
)

type Container struct {
	//读写分离
	sync.RWMutex
	Naming     naming.Naming
	Srv        x.Server
	state      uint32
	srvClients map[string]ClientMap
	selector   Selector
	dialer     x.Dialer
	deps       map[string]struct{}
}

var log = logger.WithField("module", "container")

// 一个服务只允许一个容器
// 使用单例模式初始化Container对象
var c = &Container{
	state:    0,
	selector: &HashSelector{},
	deps:     make(map[string]struct{}),
}

func Default() *Container {
	return c
}

// Init examples:
// Gateway: _ = container.Init(srv, wire.SNChat, wire.SNLogin)
// Chat: _ = container.Init(srv),no other deps
func Init(srv x.Server, deps ...string) error {
	if !atomic.CompareAndSwapUint32(&c.state, stateUninitialized, stateInitialized) {
		return errors.New("already initialized")
	}
	c.Srv = srv
	for _, dep := range deps {
		if _, ok := c.deps[dep]; ok {
			continue
		}
		c.deps[dep] = struct{}{}
	}
	log.WithField("func", "Init").Infof("srv %s:%s - deps %v", srv.ServiceID(), srv.ServiceName(), c.deps)
	c.srvClients = make(map[string]ClientMap, len(deps))
	return nil
}

func Start() error {
	if c.Naming == nil {
		return errors.New("naming is nil")
		//todo why use :return fmt.Errorf("naming is nil")
	}
	if !atomic.CompareAndSwapUint32(&c.state, stateInitialized, stateStarted) {
		return errors.New("already started")
	}
	//1.start server
	go func(srv x.Server) {
		err := srv.Start()
		if err != nil {
			log.Errorln(err)
		}
	}(c.Srv)

	//2.connect to deps
	for service := range c.deps {
		go func(service string) {
			err := connect2Service(service)
			if err != nil {
				log.Errorln(err)
			}
		}(service)
	}

	//3.register to naming
	if c.Srv.PublicAddress() != "" && c.Srv.PublicPort() != 0 {
		err := c.Naming.Register(c.Srv)
		if err != nil {
			log.Errorln(err)
		}
	}

	//wait the quit signal from system
	c := make(chan os.Signal, 1)
	//todo :why not syscall.Signal? add os.Interrupt to adapt Windows?
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	log.Infoln("shutdown signal:", <-c)
	//4.quit
	return shutdown()
}

func Push(server string, p *pkt.LogicPkt) error {
	p.AddStringMeta(common.MetaDestServer, server)
	return c.Srv.Push(server, pkt.Marshal(p))
}
func pushMessage(packet *pkt.LogicPkt) error {
	server, _ := packet.GetMeta(common.MetaDestServer)
	if server != c.Srv.ServiceID() {
		return fmt.Errorf("dest server is not correct,%s != %s", server, c.Srv.ServiceID())
	}
	channels, ok := packet.GetMeta(common.MetaDestChannels)
	if !ok {
		return fmt.Errorf("dest channels is nil")
	}

	channelIDs := strings.Split(channels.(string), ",")
	packet.DelMeta(common.MetaDestServer)
	packet.DelMeta(common.MetaDestChannels)
	payload := pkt.Marshal(packet)
	log.Debugf("Push to %v %v", channelIDs, packet)

	for _, channel := range channelIDs {
		err := c.Srv.Push(channel, payload)
		if err != nil {
			log.Debug(err)
		}
	}

	return nil
}

// Forward message to service
func Forward(serviceName string, packet *pkt.LogicPkt) error {
	if packet == nil {
		return errors.New("packet is nil")
	}
	if packet.Command == "" {
		return errors.New("command is empty in packet")
	}
	if packet.ChannelID == "" {
		return errors.New("ChannelID is empty in packet")
	}
	return ForwardWithSelector(serviceName, packet, c.selector)
}

// ForwardWithSelector 可以动态指定Selector
func ForwardWithSelector(serviceName string, packet *pkt.LogicPkt, selector Selector) error {
	cli, err := lookup(serviceName, &packet.Header, selector)
	if err != nil {
		return err
	}
	// add a tag in packet
	packet.AddStringMeta(common.MetaDestServer, c.Srv.ServiceID())
	log.Debugf("forward message to %v with %s", cli.ServiceID(), &packet.Header)
	return cli.Send(pkt.Marshal(packet))
}

func shutdown() error {
	if !atomic.CompareAndSwapUint32(&c.state, stateStarted, stateClosed) {
		return errors.New("already closed")
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()
	//1.gracefully shutdown server
	err := c.Srv.Shutdown(ctx)
	if err != nil {
		log.Errorln(err)
	}
	//2.deregister from naming
	err = c.Naming.Deregister(c.Srv.ServiceID())
	if err != nil {
		log.Warnln(err)
	}
	//3. unsubscribe deps events from naming
	for dep := range c.deps {
		_ = c.Naming.Unsubscribe(dep)
	}

	log.Infoln("shutdown")
	return nil
}

func lookup(serviceName string, header *pkt.Header, selector Selector) (x.Client, error) {
	//来自于 connect2Service
	clients, ok := c.srvClients[serviceName]
	if !ok {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}
	//只获取状态为StateAdult的服务，新上线的服务需要delay
	srvs := clients.Services(KeyServiceState, StateAdult)
	if len(srvs) == 0 {
		return nil, fmt.Errorf("no services found for %s ", serviceName)
	}
	id := selector.Lookup(header, srvs)
	if cli, ok := clients.Get(id); ok {
		return cli, nil
	}
	return nil, fmt.Errorf("no clients found")
}

// SetDialer set tcp dialer
func SetDialer(dialer x.Dialer) {
	c.dialer = dialer
}

// EnableMonitor start
func EnableMonitor(listen string) error {
	return nil
}

// SetSelector 上层业务注册一个自定义的服务路由器
func SetSelector(selector Selector) {
	c.selector = selector
}

func SetServiceNaming(nm naming.Naming) {
	c.Naming = nm
}

func connect2Service(serviceName string) error {

	return nil
}

func buildClient(clients ClientMap, service x.ServiceRegistration) (x.Client, error) {
	c.Lock()
	defer c.Unlock()
	var (
		id   = service.ServiceID()
		name = service.ServiceName()
		meta = service.GetMeta()
	)
	//1.check if client's connection exists
	if _, ok := clients.Get(id); ok {
		return nil, nil
	}
	//2.服务之间只允许使用TCP
	if service.GetProtocol() != string(common.ProtocolTCP) {
		return nil, fmt.Errorf("unexpected service protocol:%s", service.GetProtocol())
	}
	//3.build client and connect to service
	cli := tcp.NewClientWithProps(id, name, meta, tcp.ClientOptions{
		Heartbeat: x.DefaultHeartbeat,
		ReadWait:  x.DefaultReadWait,
		WriteWait: x.DefaultWriteWait,
	})
	if c.dialer == nil {
		return nil, fmt.Errorf("dialer is nil")
	}
	cli.SetDialer(c.dialer)
	err := cli.Connect(service.DialURL())
	if err != nil {
		return nil, err
	}
	//4.read messages
	go func(cli x.Client) {
		err := readLoop(cli)
		if err != nil {
			log.Debug(err)
		}
		clients.Remove(id)
		cli.Close()
	}(cli)
	//5.add to clients
	clients.Add(cli)
	return cli, nil
}

// 由于是内部服务间消息转发，不需要基础协议中的心跳（有注册中心）
func readLoop(cli x.Client) error {
	log := logger.WithFields(logger.Fields{
		"module": "container",
		"func":   "readLoop",
	})
	log.Infof("readLoop started of %s %s", cli.ServiceID(), cli.ServiceName())
	for {
		frame, err := cli.Read()
		if err != nil {
			return err
		}
		if frame.GetOpCode() != x.OpBinary {
			continue
		}
		buf := bytes.NewBuffer(frame.GetPayload())

		packet, err := pkt.MustReadLogicPkt(buf)
		if err != nil {
			log.Info(err)
			continue
		}
		err = pushMessage(packet)
		if err != nil {
			log.Info(err)
		}
	}
}
