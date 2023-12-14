package websocket

import (
	x2 "X_IM/pkg/x"
	"bufio"
	"github.com/gobwas/ws"
	"net"
)

// Upgrader Server is a websocket implement of the Server
type Upgrader struct {
}

// NewServer NewServer
func NewServer(listen string, service x2.ServiceRegistration, options ...x2.ServerOption) x2.Server {
	return x2.NewServer(listen, service, new(Upgrader), options...)
}

func (u *Upgrader) Name() string {
	return "WebSocket.Server"
}

func (u *Upgrader) Upgrade(rawConn net.Conn, rd *bufio.Reader, wr *bufio.Writer) (x2.Conn, error) {
	_, err := ws.Upgrade(rawConn)
	if err != nil {
		return nil, err
	}
	conn := NewConnWithRW(rawConn, rd, wr)
	return conn, nil
}
