package container

import (
	x "X_IM"
	"X_IM/pkg/wire/pkt"
	"hash/crc32"
)

// Selector is used to select a Service
// 选择器的作用是将请求或流量分发给适当的容器或实例，以实现负载均衡、故障转移等
// 除了hash selector，还有最少连接数、随机、轮询（顺序分发）、加权等selectors
// 如果需要添加其他selector，再考虑接口文件解耦
type Selector interface {
	Lookup(*pkt.Header, []x.Service) string
}

// HashCode generated a hash code
// 把ChannelId通过crc32算法得到一个数字，取模之后就落到了数组srvs的一个索引上
// 只要srvs的数量不发生变化，同一个用户的消息始终会落到同一台逻辑服务中。
func HashCode(key string) int {
	hash32 := crc32.NewIEEE()
	_, _ = hash32.Write([]byte(key))
	return int(hash32.Sum32())
}

type HashSelector struct {
}

// Lookup a logic
func (s *HashSelector) Lookup(header *pkt.Header, srvs []x.Service) string {
	ll := len(srvs)
	code := HashCode(header.ChannelID)
	return srvs[code%ll].ServiceID()
}
