package X_IM

//重构tcp/websocket的server逻辑，抽离统一为default server

import (
	"X_IM/pkg/logger"
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/gobwas/pool/pbufio"
	"github.com/gobwas/ws"
	"github.com/panjf2000/ants/v2"
	"github.com/segmentio/ksuid"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Upgrader 抽象握手升级逻辑
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

func WithMessageGPool(val int) ServerOption {
	return func(opts *ServerOptions) {
		opts.MessageGPool = val
	}
}

func WithConnectionGPool(val int) ServerOption {
	return func(opts *ServerOptions) {
		opts.ConnectionGPool = val
	}
}

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

// NewServer 设置server的参数，返回一个server对象
// 随后由container启动server
func NewServer(listen string, service ServiceRegistration, upgrader Upgrader, options ...ServerOption) *DefaultServer {
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

// Start server
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
	// 采用协程池来增加复用
	mgpool, _ := ants.NewPool(s.options.MessageGPool, ants.WithPreAlloc(true))
	defer func() {
		mgpool.Release()
	}()
	log.Infoln("default server started")

	for {
		rawConn, err := lst.Accept()
		if err != nil {
			if rawConn != nil {
				_ = rawConn.Close()
			}
			log.Warn(err)
			continue
		}

		go s.connHandler(rawConn, mgpool)

		if atomic.LoadInt32(&s.quit) == 1 {
			break
		}
	}
	log.Info("quit")
	return nil
}

// use pbufio's pool to reduce memory allocation
func (s *DefaultServer) connHandler(rawConn net.Conn, gpool *ants.Pool) {
	rd := pbufio.GetReader(rawConn, ws.DefaultServerReadBufferSize)
	wr := pbufio.GetWriter(rawConn, ws.DefaultServerWriteBufferSize)
	//回收缓冲池，防止被多个线程并发使用，导致数据混乱
	//通过defer在get方法结束时回收
	defer func() {
		pbufio.PutReader(rd)
		pbufio.PutWriter(wr)
	}()
	conn, err := s.Upgrade(rawConn, rd, wr)
	if err != nil {
		logger.Errorf("Upgrade error: %v", err)
		_ = rawConn.Close()
		return
	}
	id, meta, err := s.Accept(conn, s.options.LoginWait)
	if err != nil {
		_ = conn.WriteFrame(OpClose, []byte(err.Error()))
		_ = conn.Close()
		return
	}

	if _, ok := s.Get(id); ok {
		_ = conn.WriteFrame(OpClose, []byte("channelId is repeated"))
		_ = conn.Close()
		return
	}
	if meta == nil {
		meta = Meta{}
	}
	channel := NewChannel(id, meta, conn, gpool)
	channel.SetReadWait(s.options.ReadWait)
	channel.SetWriteWait(s.options.WriteWait)
	s.Add(channel)

	//gaugeWithLabel := channelTotalGauge.WithLabelValues(s.ServiceID(), s.ServiceName())
	//gaugeWithLabel.Inc()
	//defer gaugeWithLabel.Dec()

	logger.Infof("accept channel - ID: %s RemoteAddr: %s", channel.ID(), channel.RemoteAddr())
	err = channel.ReadLoop(s.MessageListener)
	if err != nil {
		logger.Info(err)
	}
	s.Remove(channel.ID())
	_ = s.Disconnect(channel.ID())
	_ = channel.Close()
}

func (s *DefaultServer) Shutdown(ctx context.Context) error {
	log := logger.WithFields(logger.Fields{
		"module": s.Name(),
		"id":     s.ServiceID(),
	})
	s.once.Do(func() {
		defer func() {
			log.Infoln("shutdown")
		}()
		if atomic.CompareAndSwapInt32(&s.quit, 0, 1) {
			return
		}

		// close channels
		channels := s.ChannelMap.All()
		for _, ch := range channels {
			_ = ch.Close()

			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
	})
	return nil
}

// Push string channelID
// []byte data
func (s *DefaultServer) Push(id string, data []byte) error {
	ch, ok := s.ChannelMap.Get(id)
	if !ok {
		return errors.New("channel no found")
	}
	return ch.Push(data)
}

// SetAcceptor SetAcceptor
func (s *DefaultServer) SetAcceptor(acceptor Acceptor) {
	s.Acceptor = acceptor
}

// SetMessageListener SetMessageListener
func (s *DefaultServer) SetMessageListener(listener MessageListener) {
	s.MessageListener = listener
}

// SetStateListener SetStateListener
func (s *DefaultServer) SetStateListener(listener StateListener) {
	s.StateListener = listener
}

func (s *DefaultServer) SetChannelMap(channels ChannelMap) {
	s.ChannelMap = channels
}

// SetReadWait set read wait duration
func (s *DefaultServer) SetReadWait(ReadWait time.Duration) {
	s.options.ReadWait = ReadWait
}

type defaultAcceptor struct {
}

// Accept defaultAcceptor
func (a *defaultAcceptor) Accept(conn Conn, timeout time.Duration) (string, Meta, error) {
	return ksuid.New().String(), Meta{}, nil
}
