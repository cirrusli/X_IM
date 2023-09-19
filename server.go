package X_IM

import (
	"context"
	"net"
	"time"
)

const (
	DefaultReadWait  = time.Minute * 3
	DefaultWriteWait = time.Second * 10
	DefaultLoginWait = time.Second * 10
	DefaultHeartbeat = time.Second * 55
)

// Service 定义了基础服务的抽象接口
type Service interface {
	ServiceID() string
	ServiceName() string
	GetMeta() map[string]string
}

// ServiceRegistration 定义服务注册的抽象接口
type ServiceRegistration interface {
	Service
	PublicAddress() string
	PublicPort() int
	DialURL() string
	GetTags() []string
	GetProtocol() string
	GetNamespace() string
	String() string
}

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

// Agent 消息发送方
type Agent interface {
	ID() string
	// Push 线程安全的发送数据的方法
	Push([]byte) error
}

type OpCode byte

const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

// Frame 解决tcp流式传输导致的封包与拆包
type Frame interface {
	SetOpCode(OpCode)
	GetOpCode() OpCode
	SetPayload([]byte)
	GetPayload() []byte
}

// Conn 封装tcp读写操作
type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(OpCode, []byte) error
	Flush() error
}

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
type Dialer interface {
	DialAndHandshake(DialerContext) (net.Conn, error)
}
type DialerContext struct {
	Id      string
	Name    string
	Address string
	Timeout time.Duration
}
