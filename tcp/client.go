package tcp

import (
	x "X_IM"
	"X_IM/logger"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type ClientOptions struct {
	Heartbeat time.Duration //登录超时
	ReadWait  time.Duration //读超时
	WriteWait time.Duration //写超时
}

type Client struct {
	sync.Mutex
	x.Dialer
	once    sync.Once
	id      string
	name    string
	conn    x.Conn
	state   int32
	options ClientOptions
	Meta    map[string]string
}

func NewClient(id, name string, opts ClientOptions) x.Client {
	return NewClientWithProps(id, name, make(map[string]string), opts)
}

// NewClientWithProps with properties
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
	logger.Infoln("in tcp/client.go:NewClientWithProps():succeed.")
	return cli
}

func (c *Client) Connect(addr string) error {
	logger.Infoln("in tcp/client.go:Connect():arrived here.")

	_, err := url.Parse(addr)
	if err != nil {
		return err
	}
	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		return fmt.Errorf("client has connected")
	}

	rawConn, err := c.Dialer.DialAndHandshake(x.DialerContext{
		ID:      c.id,
		Name:    c.name,
		Address: addr,
		Timeout: x.DefaultLoginWait,
	})
	if err != nil {
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
		return err
	}
	if rawConn == nil {
		return fmt.Errorf("conn is nil")
	}
	//封装tcp读写，与websocket统一
	c.conn = NewConn(rawConn)

	if c.options.Heartbeat > 0 {
		go func() {
			err := c.heartbeatLoop()
			if err != nil {
				logger.WithField("module", "tcp.client").Warn("heartbeat loop stopped", err)
			}
		}()
	}
	return nil
}

func (c *Client) heartbeatLoop() error {
	tick := time.NewTicker(c.options.Heartbeat)
	for range tick.C {
		if err := c.ping(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) ping() error {
	logger.WithField("module", "tcp.client").Tracef("%s send ping to server", c.id)

	err := c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}
	return c.conn.WriteFrame(x.OpPing, nil)
}

func (c *Client) Send(payload []byte) error {
	if atomic.LoadInt32(&c.state) == 0 {
		return fmt.Errorf("connection is nil")
	}
	c.Lock()
	defer c.Unlock()
	err := c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}
	return c.conn.WriteFrame(x.OpBinary, payload)
}

func (c *Client) Read() (x.Frame, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("connection is nil")
	}
	if c.options.Heartbeat > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.options.ReadWait))
	}
	frame, err := c.conn.ReadFrame()
	if err != nil {
		return nil, err
	}
	if frame.GetOpCode() == x.OpClose {
		return nil, errors.New("remote side close the channel")
	}
	return frame, nil
}

func (c *Client) Close() {
	c.once.Do(func() {
		if c.conn == nil {
			return
		}
		// graceful close connection
		_ = WriteFrame(c.conn, x.OpClose, nil)

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
func (c *Client) GetMeta() map[string]string { return c.Meta }
