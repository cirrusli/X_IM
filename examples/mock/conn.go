package mock

import (
	x "X_IM"
	"github.com/stretchr/testify/mock"
	"net"
	"time"
)

type Conn struct {
	mock.Mock
}
type Frame struct {
	OpCode  x.OpCode
	Payload []byte
}

func (f *Frame) SetOpCode(code x.OpCode) {
	f.OpCode = code
}

func (f *Frame) SetPayload(bytes []byte) {
	f.Payload = bytes
}

func (f *Frame) GetOpCode() x.OpCode {
	return f.OpCode
}

func (f *Frame) GetPayload() []byte {
	return f.Payload
}
func (m *Conn) Flush() error {
	args := m.Called()
	return args.Error(0)
}

func (m *Conn) Read(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *Conn) Write(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *Conn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *Conn) LocalAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *Conn) RemoteAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *Conn) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *Conn) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *Conn) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}
func (m *Conn) ReadFrame() (x.Frame, error) {
	args := m.Called()
	return &Frame{
		OpCode:  args.Get(0).(x.OpCode),
		Payload: args.Get(1).([]byte),
	}, args.Error(2)
}

func (m *Conn) WriteFrame(op x.OpCode, payload []byte) error {
	args := m.Called(op, payload)
	return args.Error(0)
}
