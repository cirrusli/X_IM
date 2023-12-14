package x

import (
	"X_IM/pkg/wire/pkt"
)

// Dispatcher 将消息分派到网关的组件
// 向gateway中的channels两个连接推送一条LogicPkt消息。这个能力是容器提供的
// 但是在这里为了组件的职责划分更合理，以及组件之间的解耦，不会直接依赖container
type Dispatcher interface {
	Push(gateway string, channels []string, p *pkt.LogicPkt) error
}
