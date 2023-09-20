package websocket

import (
	x "X_IM"
	"bufio"
	"github.com/gobwas/ws"
	"net"
)

const (
	DefaultReaderSize = 4096
	DefaultWriterSize = 1024
)

type Frame struct {
	raw ws.Frame
}

func (f *Frame) SetOpCode(opCode x.OpCode) {
	f.raw.Header.OpCode = ws.OpCode(opCode)
}
func (f *Frame) GetOpCode() x.OpCode {
	return x.OpCode(f.raw.Header.OpCode)
}

// SetPayload
// 由于在Frame中没有Server与Client端的区别，基于服务端逻辑优先的考虑
// 我们在GetPayload()中对Payload做了Mask解码简化Server读取逻辑
// 但是在SetPayload中是没有使用Mask来编码
// 因此客户端发送消息时不能直接使用websocket.Conn这个对象
func (f *Frame) SetPayload(payload []byte) {
	f.raw.Payload = payload
}
func (f *Frame) GetPayload() []byte {
	if f.raw.Header.Masked {
		ws.Cipher(f.raw.Payload, f.raw.Header.Mask, 0)
	}
	f.raw.Header.Masked = false
	return f.raw.Payload
}

type WsConn struct {
	net.Conn
	rd *bufio.Reader
	wr *bufio.Writer
}

func (c *WsConn) Flush() error {
	//TODO implement me
	panic("implement me")
}

func NewConn(conn net.Conn) *WsConn {
	return &WsConn{
		Conn: conn,
		rd:   bufio.NewReaderSize(conn, DefaultReaderSize),
		wr:   bufio.NewWriterSize(conn, DefaultWriterSize),
	}
}
func NewConnWithRW(conn net.Conn, rd *bufio.Reader, wr *bufio.Writer) *WsConn {
	return &WsConn{
		Conn: conn,
		rd:   rd,
		wr:   wr,
	}
}
func (c *WsConn) ReadFrame() (x.Frame, error) {
	f, err := ws.ReadFrame(c.Conn)
	if err != nil {
		return nil, err
	}
	return &Frame{raw: f}, nil
}
func (c *WsConn) WriteFrame(op x.OpCode, payload []byte) error {
	//在websocket协议中第一个bit位就是fin，表示当前帧是否为连续帧中的最后一帧
	//由于我们的数据包大小不会超过一个websocket协议单个帧最大值
	//因此这里fin直接为true，也就是不会把包拆分成多个。
	return ws.WriteFrame(c.Conn, ws.NewFrame(ws.OpCode(op), true, payload))
}
