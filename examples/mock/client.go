package mock

import (
	x "X_IM"
	"X_IM/logger"
	"X_IM/tcp"
	"X_IM/websocket"
	"context"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// ClientDemo Client demo
type ClientDemo struct {
}

func (c *ClientDemo) Start(userID, protocol, addr string) {
	logger.Infoln("in mock/client.go:Start():arrived here.")

	var cli x.Client

	// step1: 初始化客户端
	if protocol == "ws" {
		cli = websocket.NewClient(userID, "client", websocket.ClientOptions{})
		// set dialer
		cli.SetDialer(&WebsocketDialer{})
	} else if protocol == "tcp" {
		cli = tcp.NewClient("test1", "client", tcp.ClientOptions{})
		cli.SetDialer(&TCPDialer{})
	}

	// step2: 建立连接
	err := cli.Connect(addr)
	if err != nil {
		logger.Error(err)
	}
	count := 2
	logger.Infoln("in mock/client.go:Start():sending message twice.")
	go func() {
		// step3: 发送count次消息然后退出
		for i := 0; i < count; i++ {
			err := cli.Send([]byte("hello"))
			if err != nil {
				logger.Error(err)
				return
			}
			time.Sleep(time.Millisecond * 10)
		}
	}()

	// step4: 接收消息
	recv := 0
	for {
		frame, err := cli.Read()
		if err != nil {
			logger.Info(err)
			break
		}
		if frame.GetOpCode() != x.OpBinary {
			continue
		}
		recv++
		logger.Infof("cli.ServiceID:%s receive message [%s]", cli.ServiceID(), frame.GetPayload())
		if recv == count { // 接收完消息
			break
		}
	}
	cli.Close()
}

type WebsocketDialer struct {
}

func (d *WebsocketDialer) DialAndHandshake(ctx x.DialerContext) (net.Conn, error) {
	logger.Infoln("in mock/client.go:DialAndHandshake():websocket dialer")

	// 1 调用ws.Dial拨号
	conn, _, _, err := ws.Dial(context.TODO(), ctx.Address)
	if err != nil {
		return nil, err
	}
	// 2. 发送用户认证信息，示例就是userid
	err = wsutil.WriteClientBinary(conn, []byte(ctx.ID))
	if err != nil {
		return nil, err
	}
	// 3. return conn
	return conn, nil
}

type TCPDialer struct {
}

func (d *TCPDialer) DialAndHandshake(ctx x.DialerContext) (net.Conn, error) {
	logger.Infoln("in examples/mock/client.go:DialAndHandshake(): TCPDialer dialing: ", ctx.Address)
	// 1 调用net.Dial拨号
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}
	// 2. 发送用户认证信息，示例就是userid
	err = tcp.WriteFrame(conn, x.OpBinary, []byte(ctx.ID))
	if err != nil {
		return nil, err
	}
	// 3. return conn
	return conn, nil
}
