package endian

import "io"

func WriteBytes(w io.Writer, buf []byte) error {
	buflen := len(buf)
	if err := WriteUint32(w, uint32(buflen)); err != nil {
		return err
	}
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}
