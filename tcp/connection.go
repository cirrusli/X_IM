package tcp

import (
	x "X_IM"
	"X_IM/wire/endian"
	"io"
)

func WriteFrame(w io.Writer, code x.OpCode, payload []byte) error {
	if err := endian.WriteUint8(w, uint8(code)); err != nil {
		return err
	}
	if err := endian.WriteBytes(w, payload); err != nil {
		return err
	}
	return nil
}
