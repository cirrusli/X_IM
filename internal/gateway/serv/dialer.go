package serv

import (
	"X_IM/pkg/tcp"
	"X_IM/pkg/wire/pkt"
	"X_IM/pkg/x"
	"google.golang.org/protobuf/proto"
	"net"
)

type TCPDialer struct {
	ServiceID string
}

func NewDialer(serviceID string) x.Dialer {
	return &TCPDialer{ServiceID: serviceID}
}

// DialAndHandshake ServiceID重复服务端会关闭连接，容器会把新创建的这个Client删除
func (d *TCPDialer) DialAndHandshake(ctx x.DialerContext) (net.Conn, error) {
	//1. as a client 连接到服务
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}
	req := &pkt.InnerHandshakeReq{
		ServiceID: d.ServiceID,
	}
	log.Infof("in DialAndHandshake(): send req: %+v", req)
	//2.将自己的ServiceID发送给连接方
	bts, _ := proto.Marshal(req)
	err = tcp.WriteFrame(conn, x.OpBinary, bts)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
