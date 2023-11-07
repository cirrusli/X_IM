package dialer

import (
	x "X_IM"
	"X_IM/pkg/logger"
	"X_IM/wire/common"
	"X_IM/wire/pkt"
	"X_IM/wire/token"
	"bytes"
	"context"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	"time"
)

type ClientDialer struct {
	AppSecret string
}

func (d *ClientDialer) DialAndHandshake(ctx x.DialerContext) (net.Conn, error) {
	logger.Infoln("in dialer/client.go:DialAndHandshake():arrived here.")
	// 1. 拨号
	conn, _, _, err := ws.Dial(context.TODO(), ctx.Address)
	if err != nil {
		return nil, err
	}
	if d.AppSecret == "" {
		d.AppSecret = token.DefaultSecret
	}
	// 2. 直接使用封装的JWT包生成一个token
	tk, err := token.Generate(d.AppSecret, &token.Token{
		Account: ctx.ID,
		App:     "x_im",
		Exp:     time.Now().AddDate(0, 0, 1).Unix(),
	})
	if err != nil {
		return nil, err
	}
	// 3. 发送一条CommandLoginSignIn消息
	loginReq := pkt.New(common.CommandLoginSignIn).WriteBody(&pkt.LoginReq{
		Token: tk,
	})
	err = wsutil.WriteClientBinary(conn, pkt.Marshal(loginReq))
	if err != nil {
		return nil, err
	}

	// wait resp
	_ = conn.SetReadDeadline(time.Now().Add(ctx.Timeout))
	logger.Infoln("in dialer/client.go:DialAndHandshake():doing login_test.go:will read frame.")
	frame, err := ws.ReadFrame(conn)
	if err != nil {
		return nil, err
	}
	ack, err := pkt.MustReadLogicPkt(bytes.NewBuffer(frame.Payload))
	if err != nil {
		return nil, err
	}
	// 4. 判断是否登录成功
	if ack.Status != pkt.Status_Success {
		return nil, fmt.Errorf("login failed: %v", &ack.Header)
	}
	var resp = new(pkt.LoginResp)
	_ = ack.ReadBody(resp)

	logger.Debug("logined ", resp.GetChannelID())
	return conn, nil
}
