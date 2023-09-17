package X_IM

import (
	"context"
	"net"
	"time"
)

type Server interface {
	SetAcceptor(Acceptor)
	SetMessageListener(MessageListener)
	SetStateListener(StateListener)
	SetReadWait(time.Duration)
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
type Channel interface {
	Conn
	Agent
	// Close 关闭连接
	Close() error
	ReadLoop(lst MessageListener) error
	// SetWriteWait 设置写超时
	SetWriteWait(time.Duration)
	SetReadWait(time.Duration)
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
