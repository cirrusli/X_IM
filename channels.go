package X_IM

import (
	"sync"
	"time"
)

type ChannelMap interface {
	Add(channel Channel)
	Remove(id string)
	Get(id string) (Channel, bool)
	All() []Channel
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
type ChannelsImpl struct {
	channels *sync.Map
}

func NewChannels(num int) ChannelMap {
	return &ChannelsImpl{
		channels: new(sync.Map),
	}
}
func (ch *ChannelsImpl) Add(channel Channel) {
	//todo
}
func (ch *ChannelsImpl) Remove(id string) {
	//todo
}
func (ch *ChannelsImpl) Get(id string) (Channel, bool) {
	//todo
	return nil, false
}
func (ch *ChannelsImpl) All() []Channel {
	//todo
	return nil
}
