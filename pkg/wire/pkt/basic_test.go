package pkt

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicPktEncode(t *testing.T) {
	pkt := &BasicPkt{
		Code:   CodePing,
		Length: 5,
		Body:   []byte("hello"),
	}

	buf := new(bytes.Buffer)
	err := pkt.Encode(buf)

	assert.Nil(t, err)
	assert.Equal(t, []byte{0, 1, 0, 5, 'h', 'e', 'l', 'l', 'o'}, buf.Bytes())
}

func TestBasicPktDecode(t *testing.T) {
	pkt := &BasicPkt{}

	buf := bytes.NewBuffer([]byte{0, 1, 0, 5, 'h', 'e', 'l', 'l', 'o'})
	err := pkt.Decode(buf)

	assert.Nil(t, err)
	assert.Equal(t, CodePing, pkt.Code)
	assert.Equal(t, uint16(5), pkt.Length)
	assert.Equal(t, []byte("hello"), pkt.Body)
}
