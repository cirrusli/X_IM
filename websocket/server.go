package websocket

import (
	x "X_IM"
	"X_IM/naming"
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
