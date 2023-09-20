package X_IM

import (
	"X_IM/logger"
	"bufio"
	"context"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/segmentio/ksuid"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Upgrader interface {
	Name() string
	Upgrade(rawConn net.Conn, rd *bufio.Reader, wr *bufio.Writer) (Conn, error)
}
type ServerOptions struct {
	LoginWait       time.Duration //登录超时
	ReadWait        time.Duration //读超时
	WriteWait       time.Duration //写超时
	MessageGPool    int
	ConnectionGPool int
}

type ServerOption func(*ServerOptions)

// DefaultServer is a websocket implement of the DefaultServer
type DefaultServer struct {
	Upgrader
	listen string
	ServiceRegistration
	ChannelMap
	Acceptor
	MessageListener
	StateListener
	once    sync.Once
	options *ServerOptions
	quit    int32
}

func NewServer(listen string, service ServiceRegistration, upgrader Upgrader, options ...ServerOption) Server {
	defaultOpts := &ServerOptions{
		LoginWait:       DefaultLoginWait,
		ReadWait:        DefaultReadWait,
		WriteWait:       DefaultWriteWait,
		MessageGPool:    DefaultMessageReadPool,
		ConnectionGPool: DefaultConnectionPool,
	}
	for _, option := range options {
		option(defaultOpts)
	}
	return &DefaultServer{
		listen:              listen,
		ServiceRegistration: service,
		options:             defaultOpts,
		Upgrader:            upgrader,
		quit:                0,
	}
}

type defaultAcceptor struct {
}

func (a *defaultAcceptor) Accept(conn Conn, timeout time.Duration) (string, Meta, error) {
	return ksuid.New().String(), Meta{}, nil
}
func (s *DefaultServer) Start() error {
	log := logger.WithFields(logger.Fields{
		"module": s.Name(),
		"listen": s.listen,
		"id":     s.ServiceID(),
		"func":   "Start",
	})

	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}
	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}
	if s.ChannelMap == nil {
		s.ChannelMap = NewChannels(100)
	}
	lst, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}
	//采用协程池来增加复用
	mgpool, NewPoolErr := ants.NewPool(s.options.MessageGPool, ants.WithPreAlloc(true))
	if NewPoolErr != nil {
		//TODO if panic?
		log.Warn("NewPoolErr:", NewPoolErr)
	}
	defer func() {
		mgpool.Release()
	}()
	log.Info("Started")

	for {
		rawConn, err := lst.Accept()
		if err != nil {
			if rawConn != nil {
				_ = rawConn.Close()
			}
			log.Warn(err)
			continue
		}
		//处理连接中的消息
		go s.connHandler(rawConn, mgpool)
		//原子操作来检查Server是否需要退出，结束循环
		if atomic.LoadInt32(&s.quit) == 1 {
			break
		}
	}
	log.Info("Quited")
	return nil
}
func (s *DefaultServer) connHandler(conn net.Conn, mgpool *ants.Pool) {
	//TODO implement me
	return
}
func (s *DefaultServer) SetAcceptor(acceptor Acceptor) {
	//TODO implement me
	panic("implement me")
}

func (s *DefaultServer) SetMessageListener(listener MessageListener) {
	//TODO implement me
	panic("implement me")
}

func (s *DefaultServer) SetStateListener(listener StateListener) {
	//TODO implement me
	panic("implement me")
}

func (s *DefaultServer) SetReadWait(duration time.Duration) {
	//TODO implement me
	panic("implement me")
}

func (s *DefaultServer) SetChannelMap(channelMap ChannelMap) {
	//TODO implement me
	panic("implement me")
}

func (s *DefaultServer) Push(id string, data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (s *DefaultServer) Shutdown(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}
