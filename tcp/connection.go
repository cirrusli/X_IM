package tcp

import (
	x "X_IM"
	"X_IM/wire/endian"
	"bufio"
	"io"
	"net"
)

// WrappedConn is the tcp connection
type WrappedConn struct {
	net.Conn
	rd *bufio.Reader
	wr *bufio.Writer
}
type Frame struct {
	OpCode  x.OpCode
	Payload []byte
}

// NewConn 用来包装TCP的底层连接
func NewConn(conn net.Conn) x.Conn {
	return &WrappedConn{
		Conn: conn,
		rd:   bufio.NewReaderSize(conn, 4096),
		wr:   bufio.NewWriterSize(conn, 1024),
	}
}

func NewConnWithRW(conn net.Conn, rd *bufio.Reader, wr *bufio.Writer) *WrappedConn {
	return &WrappedConn{
		Conn: conn,
		rd:   rd,
		wr:   wr,
	}
}

func (c *WrappedConn) ReadFrame() (x.Frame, error) {
	opcode, err := endian.ReadUint8(c.Conn)
	if err != nil {
		return nil, err
	}
	payload, err := endian.ReadBytes(c.Conn)
	if err != nil {
		return nil, err
	}
	return &Frame{
		OpCode:  x.OpCode(opcode),
		Payload: payload,
	}, nil
}

func (c *WrappedConn) WriteFrame(code x.OpCode, payload []byte) error {
	return WriteFrame(c.Conn, code, payload)
}
func (c *WrappedConn) Flush() error {
	return c.wr.Flush()
}

func WriteFrame(w io.Writer, code x.OpCode, payload []byte) error {
	if err := endian.WriteUint8(w, uint8(code)); err != nil {
		return err
	}
	if err := endian.WriteBytes(w, payload); err != nil {
		return err
	}
	return nil
}

func (f Frame) SetOpCode(code x.OpCode) {
	f.OpCode = code
}

func (f Frame) GetOpCode() x.OpCode {
	return f.OpCode
}

func (f Frame) SetPayload(payload []byte) {
	f.Payload = payload
}

func (f Frame) GetPayload() []byte {
	return f.Payload
}
