package X_IM

import (
	"sync"
)

type ChannelMap interface {
	Add(channel Channel)
	Remove(id string)
	Get(id string) (Channel, bool)
	All() []Channel
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
