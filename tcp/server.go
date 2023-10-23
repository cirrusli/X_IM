package tcp

import (
	x "X_IM"
	"X_IM/logger"
	"X_IM/naming"
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/ksuid"
	"net"
	"sync"
	"time"
)

type ServerOptions struct {
	loginWait time.Duration //登录超时
	readWait  time.Duration //读超时
	writeWait time.Duration //写超时
}

type Server struct {
	listen string
	naming.ServiceRegistration
	x.ChannelMap
	x.Acceptor
	x.MessageListener
	x.StateListener
	once    sync.Once
	options ServerOptions
	//管理退出事件
	quit *x.Event
}

type defaultAcceptor struct{}

// Accept 生成一个新的channelID
func (a *defaultAcceptor) Accept(conn x.Conn, timeout time.Duration) (string, error) {
	return ksuid.New().String(), nil
}

func NewServer(listen string, service naming.ServiceRegistration) x.Server {
	return &Server{
		listen:              listen,
		ServiceRegistration: service,
		ChannelMap:          x.NewChannels(100),
		options: ServerOptions{
			loginWait: x.DefaultLoginWait,
			readWait:  x.DefaultReadWait,
			writeWait: x.DefaultWriteWait,
		},
		quit: x.NewEvent(),
	}
}

func (s *Server) Start() error {
	log := logger.WithFields(logger.Fields{
		"module": "tcp.server",
		"listen": s.listen,
		"id":     s.ServiceID(),
	})

	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}
	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}
	lst, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}
	log.Info("started")

	for {
		rawConn, err := lst.Accept()
		if err != nil {
			_ = rawConn.Close()
			log.Warn(err)
			continue
		}
		//此处与websocket/server的参数不同
		//因为accept()后的逻辑影响新连接
		//所以需要使协程尽快调度出去
		go func(rawConn net.Conn) {
			conn := NewConn(rawConn)
			defer func(conn *WrappedConn) {
				err := conn.Close()
				if err != nil {
					log.Warn(err)
				}
			}(conn)

			id, err := s.Accept(conn, s.options.loginWait)
			if err != nil {
				_ = conn.WriteFrame(x.OpClose, []byte(err.Error()))
				//_ = conn.Close()
				return
			}
			if _, ok := s.Get(id); ok {
				log.Warnf("ChannelID:%s already exists", id)
				//通知客户端ChannelID已经存在
				_ = conn.WriteFrame(x.OpClose, []byte("ChannelID is duplicate"))
				//_ = conn.Close()
				return
			}

			channel := x.NewChannel(id, conn)
			channel.SetReadWait(s.options.readWait)
			channel.SetWriteWait(s.options.writeWait)

			s.Add(channel)

			log.Info("accept ", channel)
			err = channel.ReadLoop(s.MessageListener)
			if err != nil {
				log.Info(err)
			}
			//ReadLoop返回err说明连接断开，从ChannelMap中移除
			s.Remove(channel.ID())
			//回调业务层
			_ = s.Disconnect(channel.ID())
			//channel.Close()
		}(rawConn)

		select {
		case <-s.quit.Done():
			return fmt.Errorf("listen quit")
		default:
		}
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	log := logger.WithFields(logger.Fields{
		"module": "tcp.server",
		"id":     s.ServiceID(),
	})
	s.once.Do(func() {
		defer func() {
			log.Infoln("shutdown")
		}()
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
	//todo <-s.quit.Done()？
	return nil
}

func (s *Server) SetAcceptor(acceptor x.Acceptor) {
	s.Acceptor = acceptor
}

func (s *Server) SetMessageListener(listener x.MessageListener) {
	s.MessageListener = listener
}

func (s *Server) SetStateListener(listener x.StateListener) {
	s.StateListener = listener
}

func (s *Server) SetReadWait(readWait time.Duration) {
	s.options.readWait = readWait
}

func (s *Server) SetChannelMap(channels x.ChannelMap) {
	s.ChannelMap = channels
}

func (s *Server) Push(id string, data []byte) error {
	ch, ok := s.ChannelMap.Get(id)
	if !ok {
		return errors.New("channel no found")
	}
	return ch.Push(data)
}
