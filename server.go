package X_IM

import (
	"context"
	"net"
	"time"
)

type OpCode byte

const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

const (
	DefaultReadWait  = time.Minute * 3
	DefaultWriteWait = time.Second * 10
	DefaultLoginWait = time.Second * 10
	DefaultHeartbeat = time.Second * 55
)

// 定义读取消息的默认goroutine池大小
const (
	DefaultMessageReadPool = 5000
	DefaultConnectionPool  = 5000
)

// Server 定义了一个tcp/websocket不同协议通用的服务端的接口
type Server interface {
	ServiceRegistration
	SetAcceptor(Acceptor)
	// SetMessageListener 设置上行消息监听器
	SetMessageListener(MessageListener)
	// SetStateListener 设置连接状态监听服务
	SetStateListener(StateListener)
	// SetReadWait 设置读超时
	SetReadWait(time.Duration)
	// SetChannelMap 设置Channel管理服务
	SetChannelMap(ChannelMap)

	Start() error
	Push(string, []byte) error
	Shutdown(ctx context.Context) error
}

type Channel interface {
	Conn
	Agent
	// Close 关闭连接，重写net.Conn的Close方法
	Close() error
	// ReadLoop 阻塞的方法，将读消息和心跳封装
	ReadLoop(lst MessageListener) error
	SetWriteWait(time.Duration)
	SetReadWait(time.Duration)
}

// Agent 消息发送方
type Agent interface {
	ID() string
	// Push 线程安全的发送数据的方法，channel实现了消息的异步批量发送
	Push([]byte) error
}

// Frame 解决tcp流式传输导致的封包与拆包，将tcp/websocket的数据接口统一
type Frame interface {
	SetOpCode(OpCode)
	GetOpCode() OpCode
	SetPayload([]byte)
	GetPayload() []byte
}

// Conn 封装tcp读写操作，保证与websocket的帧读写一起抽象
type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(OpCode, []byte) error
	Flush() error
}

// Client 定义了一个tcp/websocket不同协议通用的客户端的接口
type Client interface {
	ID() string
	Name() string
	Connect(string) error
	// SetDialer 由Connect调用，完成连接的握手和建立
	SetDialer(Dialer)
	Send([]byte) error
	// Read 读取一帧数据，底层复用了X.Conn，因此返回Frame
	Read() (Frame, error)
	Close()
}

// Acceptor Start()监听到连接后，调用此方法让业务层处理握手
type Acceptor interface {
	// Accept 返回error则断开连接
	Accept(Conn, time.Duration) (string, error)
}

// StateListener 报告断开连接的事件
type StateListener interface {
	Disconnect(string) error
}

type MessageListener interface {
	Receive(Agent, []byte)
}

type Dialer interface {
	DialAndHandshake(DialerContext) (net.Conn, error)
}

// DialerContext 定义了连接的上下文信息
type DialerContext struct {
	ID      string
	Name    string
	Address string
	Timeout time.Duration
}
type Meta map[string]string

// Service 定义基础服务的抽象接口
type Service interface {
	ServiceID() string
	ServiceName() string
	// GetMeta 获取服务的元数据
	GetMeta() map[string]string
}

// ServiceRegistration 定义服务注册的抽象接口
type ServiceRegistration interface {
	Service
	// PublicAddress 获取服务的地址
	PublicAddress() string
	// PublicPort 获取服务的端口
	PublicPort() int
	// DialURL 获取服务的连接地址
	DialURL() string
	// GetTags 获取服务的标签
	GetTags() []string
	// GetProtocol 获取服务的协议
	GetProtocol() string
	// GetNamespace 获取服务的命名空间
	GetNamespace() string
	String() string
}
