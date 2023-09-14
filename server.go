package main

import (
	"context"
	"net"
	"time"
)

const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
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

// Acceptor Start()监听到连接后，调用此方法让上层业务处理握手
type Acceptor interface {
	Accept(Conn, time.Duration) (string, error)
}

type SetStateListener interface {
	Disconnect(string) error
}

type MessageListener interface {
	Receive(Agent, []byte)
}

type Agent interface {
	ID() string
	Push([]byte) error
}
type Frame interface {
	SetOpCode(OpCode)
	GetOpCode() OpCode
	SetPayload([]byte)
	GetPayload() []byte
}

// Conn Connection
type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(OpCode, []byte) error
	Flush() error
}
