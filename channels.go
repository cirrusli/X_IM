package X_IM

import "time"

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
