package tcp

import (
	x "X_IM"
	"bufio"
	"net"
)

type Upgrader struct {
}

// NewServer 返回 default server 对象
func NewServer(listen string, service x.ServiceRegistration, options ...x.ServerOption) x.Server {
	return x.NewServer(listen, service, new(Upgrader), options...)
}

func (u *Upgrader) Name() string {
	return "TCP.Server"
}

func (u *Upgrader) Upgrade(rawConn net.Conn, rd *bufio.Reader, wr *bufio.Writer) (x.Conn, error) {
	conn := NewConnWithRW(rawConn, rd, wr)
	return conn, nil
}
