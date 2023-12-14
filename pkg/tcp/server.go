package tcp

import (
	x2 "X_IM/pkg/x"
	"bufio"
	"net"
)

type Upgrader struct {
}

// NewServer 返回 default server 对象
func NewServer(listen string, service x2.ServiceRegistration, options ...x2.ServerOption) x2.Server {
	return x2.NewServer(listen, service, new(Upgrader), options...)
}

func (u *Upgrader) Name() string {
	return "TCP.Server"
}

func (u *Upgrader) Upgrade(rawConn net.Conn, rd *bufio.Reader, wr *bufio.Writer) (x2.Conn, error) {
	conn := NewConnWithRW(rawConn, rd, wr)
	return conn, nil
}
