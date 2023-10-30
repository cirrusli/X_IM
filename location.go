package X_IM

import (
	"X_IM/wire/endian"
	"bytes"
	"errors"
)

// Location 通过网关ID和ChannelID定位用户
// 发送消息时寻址只需要ChannelID与GateID两个信息即可定位到网关与网关中的channel，
// 在高频的消息发送时，特别是存在群的寻址扩散情况下，就要尽量提高寻址的速度。
// 在不优化外部逻辑的情况下，可以从两个方面做优化：
// 减少寻址内容空间占用。
// 提高寻址内容序列化速度与内存分配。
// 因此从Session中分出Location
type Location struct {
	ChannelID string
	GateID    string
}

// Bytes 使用自定义的序列化方法
func (loc *Location) Bytes() []byte {
	if loc == nil {
		return []byte{}
	}
	buf := new(bytes.Buffer)
	_ = endian.WriteShortBytes(buf, []byte(loc.ChannelID))
	_ = endian.WriteShortBytes(buf, []byte(loc.GateID))
	return buf.Bytes()
}

// Unmarshal 使用自定义的反序列化方法
func (loc *Location) Unmarshal(data []byte) (err error) {
	if len(data) == 0 {
		return errors.New("data is empty")
	}
	buf := bytes.NewBuffer(data)
	loc.ChannelID, err = endian.ReadShortString(buf)
	if err != nil {
		return
	}
	loc.GateID, err = endian.ReadShortString(buf)
	if err != nil {
		return
	}
	return
}
