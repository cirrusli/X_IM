package container

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus 指标（metric）来统计消息的出流量字节数
var messageOutFlowBytes = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "x_im",
	Name:      "message_out_flow_bytes",
	Help:      "网关下发的消息字节数",
}, []string{"command"})
