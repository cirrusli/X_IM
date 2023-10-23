package websocket

import (
	x "X_IM"
	"X_IM/logger"
	"X_IM/naming"
	"context"
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/segmentio/ksuid"
	"net/http"
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
		options: ServerOptions{
			loginWait: x.DefaultLoginWait,
			readWait:  x.DefaultReadWait,
			writeWait: x.DefaultWriteWait,
		},
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	log := logger.WithFields(logger.Fields{
		"module": "ws.server",
		"listen": s.listen,
		"id":     s.ServiceID(),
	})

	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}
	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}
	if s.ChannelMap == nil {
		//创建默认的连接管理器
		s.ChannelMap = x.NewChannels(100)
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rawConn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			respError(w, http.StatusBadRequest, err.Error())
			return
		}
		//wrap rawConn to Conn
		conn := NewConn(rawConn)
		//accept 回调给上层业务完成权限认证等操作
		id, err := s.Acceptor.Accept(conn, s.options.loginWait)
		if err != nil {
			_ = conn.WriteFrame(x.OpClose, []byte(err.Error()))
			_ = conn.Close()
			return
		}
		//创建一个新的Channel
		ch := x.NewChannel(id, conn)
		ch.SetReadWait(s.options.readWait)
		ch.SetWriteWait(s.options.writeWait)
		//将Channel添加到ChannelMap中
		s.Add(ch)
		//循环读取消息
		go func(ch x.Channel) {
			defer func(ch x.Channel) {
				err := ch.Close()
				if err != nil {
					log.Info(err)
				}
			}(ch)

			err := ch.ReadLoop(s.MessageListener)
			if err != nil {
				log.Info(err)
			}
			//从ChannelMap中移除并断开连接
			s.Remove(ch.ID())
			err = s.Disconnect(ch.ID())
			if err != nil {
				log.Warn(err)
			}
			//ch.Close()
		}(ch)
	})
	log.Infoln("started")
	return http.ListenAndServe(s.listen, mux)
}

func respError(w http.ResponseWriter, code int, body string) {
	w.WriteHeader(code)
	if body != "" {
		_, _ = w.Write([]byte(body))
	}
	logger.Warnf("response with code:%d %s", code, body)
}

func (s *Server) Shutdown(ctx context.Context) error {
	log := logger.WithFields(logger.Fields{
		"module": "ws.server",
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
		return errors.New("channel not found")
	}
	return ch.Push(data)
}
