package pkt

import (
	"X_IM/wire/endian"
	"io"
)

const (
	CodePing = uint16(1)
	CodePong = uint16(2)
)

type BasicPkt struct {
	Code   uint16
	Length uint16
	Body   []byte
}

// Decode 解码，从r中读取数据到p中
func (p *BasicPkt) Decode(r io.Reader) error {
	var err error
	if p.Code, err = endian.ReadUint16(r); err != nil {
		return err
	}
	if p.Length, err = endian.ReadUint16(r); err != nil {
		return err
	}
	if p.Length > 0 {
		if p.Body, err = endian.ReadFixedBytes(int(p.Length), r); err != nil {
			return err
		}
	}
	return nil
}

// Encode 编码，将p中的数据写入w中
func (p *BasicPkt) Encode(w io.Writer) error {
	if err := endian.WriteUint16(w, p.Code); err != nil {
		return err
	}
	if err := endian.WriteUint16(w, p.Length); err != nil {
		return err
	}
	if p.Length > 0 {
		if _, err := w.Write(p.Body); err != nil {
			return err
		}
	}
	return nil
}
