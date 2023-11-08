package websocket

import (
	x "X_IM"
	"bufio"
	"github.com/gobwas/ws"
	"net"
)

// Upgrader Server is a websocket implement of the Server
type Upgrader struct {
}

// NewServer NewServer
func NewServer(listen string, service x.ServiceRegistration, options ...x.ServerOption) x.Server {
	return x.NewServer(listen, service, new(Upgrader), options...)
}

func (u *Upgrader) Name() string {
	return "WebSocket.Server"
}

func (u *Upgrader) Upgrade(rawConn net.Conn, rd *bufio.Reader, wr *bufio.Writer) (x.Conn, error) {
	_, err := ws.Upgrade(rawConn)
	if err != nil {
		return nil, err
	}
	conn := NewConnWithRW(rawConn, rd, wr)
	return conn, nil
}
