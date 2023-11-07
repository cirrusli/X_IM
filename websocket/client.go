package websocket

import (
	x "X_IM"
	"X_IM/pkg/logger"
	"errors"
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

func NewClient(id, name string, opts ClientOptions) x.Client {
	return NewClientWithProps(id, name, make(map[string]string), opts)
}

func NewClientWithProps(id, name string, meta map[string]string, opts ClientOptions) x.Client {
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
		Meta:    meta,
	}
	logger.Infoln("in websocket/client.go:NewClientWithProps():succeed.")
	return cli
}

// Connect to logic
func (c *Client) Connect(addr string) error {
	logger.Infoln("in websocket/client.go:Connect():arrived here.")

	_, err := url.Parse(addr)
	if err != nil {
		return err
	}
	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		return fmt.Errorf("client has connected")
	}
	// 拨号与握手
	conn, err := c.Dialer.DialAndHandshake(x.DialerContext{
		ID:      c.id,
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

func (c *Client) heartbeatLoop(conn net.Conn) error {
	tick := time.NewTicker(c.options.Heartbeat)
	for range tick.C {
		// 发送一个ping的心跳包给服务端
		if err := c.ping(conn); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) ping(conn net.Conn) error {
	c.Lock()
	defer c.Unlock()
	//写消息之前通过conn.SetWriteDeadline重置写超时
	//如果连接异常在发送端可以感知到
	err := conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}
	logger.Tracef("%s send ping to logic", c.id)
	return wsutil.WriteClientMessage(conn, ws.OpPing, nil)
}

// Send data to connection
func (c *Client) Send(payload []byte) error {

	if atomic.LoadInt32(&c.state) == 0 {
		return fmt.Errorf("client has not connected")
	}
	c.Lock()
	defer c.Unlock()
	err := c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}
	//内部会判断isClientSide，会使用mask
	return wsutil.WriteClientMessage(c.conn, ws.OpBinary, payload)
}

// not thread safe!
func (c *Client) Read() (x.Frame, error) {
	if c.conn == nil {
		return nil, errors.New("connection is nil")
	}
	if c.options.Heartbeat > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.options.ReadWait))
	}
	frame, err := ws.ReadFrame(c.conn)
	if err != nil {
		return nil, err
	}
	if frame.Header.OpCode == ws.OpClose {
		return nil, errors.New("remote side closed the channel")
	}
	return &Frame{raw: frame}, nil
}

func (c *Client) Close() {
	c.once.Do(func() {
		if c.conn == nil {
			return
		}
		// graceful close connection
		_ = wsutil.WriteClientMessage(c.conn, ws.OpClose, nil)

		_ = c.conn.Close()
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
	})
}

func (c *Client) SetDialer(dialer x.Dialer) {
	c.Dialer = dialer
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
