package websocket

import (
	x "X_IM"
	"X_IM/logger"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// ClientOptions ClientOptions
type ClientOptions struct {
	Heartbeat time.Duration //登录超时
	ReadWait  time.Duration //读超时
	WriteWait time.Duration //写超时
}

// Client is a websocket implement of the terminal
type Client struct {
	sync.Mutex
	x.Dialer
	once    sync.Once
	id      string
	name    string
	conn    net.Conn
	state   int32
	options ClientOptions
	Meta    map[string]string
}

// NewClient NewClient
func NewClient(id, name string, opts ClientOptions) x.Client {
	if opts.WriteWait == 0 {
		opts.WriteWait = x.DefaultWriteWait
	}
	if opts.ReadWait == 0 {
		opts.ReadWait = x.DefaultReadWait
	}
	cli := &Client{
		id:      id,
		name:    name,
		options: opts,
	}
	return cli
}

// Connect to server
func (c *Client) Connect(addr string) error {
	_, err := url.Parse(addr)
	if err != nil {
		return err
	}
	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		return fmt.Errorf("client has connected")
	}
	// 拨号与握手
	conn, err := c.Dialer.DialAndHandshake(x.DialerContext{
		Id:      c.id,
		Name:    c.name,
		Address: addr,
		Timeout: x.DefaultLoginWait,
	})
	if err != nil {
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
		return err
	}
	if conn == nil {
		return fmt.Errorf("conn is nil")
	}
	c.conn = conn

	if c.options.Heartbeat > 0 {
		go func() {
			err := c.heartbeatLoop(conn)
			if err != nil {
				logger.Error("heartbeatLoop stopped ", err)
			}
		}()
	}
	return nil
}
func (c *Client) SetDialer(dialer x.Dialer) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Send(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Read() (x.Frame, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Close() {
	//TODO implement me
	panic("implement me")
}
func (c *Client) heartbeatLoop(conn net.Conn) error {
	tick := time.NewTicker(c.options.Heartbeat)
	for range tick.C {
		if err := c.ping(conn); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) ping(conn net.Conn) error {
	c.Lock()
	defer c.Unlock()
	err := conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}
	logger.Tracef("%s send ping to server", c.id)
	return wsutil.WriteClientMessage(conn, ws.OpPing, nil)
}
func (c *Client) ServiceID() string {
	return c.id
}

func (c *Client) ServiceName() string {
	return c.name
}

func (c *Client) GetMeta() map[string]string {
	return c.Meta
}
