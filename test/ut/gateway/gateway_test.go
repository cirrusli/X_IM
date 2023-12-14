package gateway

import (
	"X_IM/internal/gateway/serv"
	"X_IM/pkg/logger"
	"X_IM/pkg/token"
	"X_IM/pkg/wire/common"
	"X_IM/pkg/wire/pkt"
	"X_IM/pkg/x"
	mymock "X_IM/test/mock"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

var (
	loginReq      = &pkt.LoginReq{}
	loginReqBytes []byte
)

const tokenSecret = "test_secret"

type Handler struct {
	ServiceID string
	AppSecret string
	Container *mymock.Container
}

func init() {
	// 生成一个JWT token
	mytoken := &token.Token{
		Account: "test_account",
		App:     "test_app",
		Exp:     time.Now().Add(time.Hour * 24 * 7).Unix(),
	}
	secret := tokenSecret

	tk, err := token.Generate(secret, mytoken)

	// 创建一个LoginReq对象
	loginReq = &pkt.LoginReq{
		Token: tk,
	}

	// 创建一个LogicPkt对象
	logicPkt := pkt.New(common.CommandLoginSignIn).WriteBody(loginReq)

	// 序列化LogicPkt对象
	buf := new(bytes.Buffer)
	err = logicPkt.Encode(buf)
	if err != nil {
		logger.Errorln("Error writing packet:", err)
		return
	}

	// 打印序列化后的数据
	logger.Infoln("init(): after Encode: Serialized login packet: ", buf.Bytes())

	// 将LogicPkt对象序列化为字节切片
	loginReqBytes = pkt.Marshal(logicPkt)
	logger.Infoln("init(): after pkt.Marshal,with magic code: ", loginReqBytes)
}

func TestHandlerAccept(t *testing.T) {
	// 创建一个模拟的Conn对象
	conn := new(mymock.Conn)
	// 设置期望的行为
	conn.On("SetReadDeadline", mock.Anything).Return(nil)
	conn.On("ReadFrame").Return(x.OpBinary, loginReqBytes, nil)
	//conn.On("WriteFrame", x.OpBinary, mock.Anything).Return(nil)
	//conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080})

	// 创建一个模拟的Container对象
	container := new(mymock.Container)
	container.On("Forward", common.SNLogin, mock.Anything).Return(nil)

	// 创建一个Handler对象
	handler := &Handler{
		ServiceID: "test_gateway01",
		AppSecret: tokenSecret,
		Container: container,
	}

	// 调用Accept方法
	id, meta, err := handler.Accept(conn, time.Second)
	t.Log("id is ", id, "\nmeta is ", meta)
	// 验证结果
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.Equal(t, "test_app", meta[serv.MetaKeyApp])
	assert.Equal(t, "test_account", meta[serv.MetaKeyAccount])

	// 验证是否所有的期望都已满足
	conn.AssertExpectations(t)
	container.AssertExpectations(t)
}

// 重写services/gateway/serv/handler.go中的Accept方法
// 将container.Forward()及接收方login server进行mock
func (h *Handler) Accept(conn x.Conn, timeout time.Duration) (string, x.Meta, error) {
	// 设置读取超时
	_ = conn.SetReadDeadline(time.Now().Add(timeout))

	// 读取登录包
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", nil, err
	}

	// 解析登录包
	buf := bytes.NewBuffer(frame.GetPayload())
	req, err := pkt.MustReadLogicPkt(buf)
	if err != nil {
		return "", nil, err
	}

	// 必须是登录包
	if req.Command != common.CommandLoginSignIn {
		resp := pkt.NewFrom(&req.Header)
		resp.Status = pkt.Status_InvalidCommand
		_ = conn.WriteFrame(x.OpBinary, pkt.Marshal(resp))

		return "", nil, fmt.Errorf("is an \"InvalidCommand\" command")
	}

	// 解析Body
	var login pkt.LoginReq
	err = req.ReadBody(&login)
	if err != nil {
		return "", nil, err
	}

	// 验证token
	secret := h.AppSecret
	if secret == "" {
		secret = token.DefaultSecret
	}
	tk, err := token.Parse(secret, login.Token)
	if err != nil {
		resp := pkt.NewFrom(&req.Header)
		resp.Status = pkt.Status_Unauthenticated
		_ = conn.WriteFrame(x.OpBinary, pkt.Marshal(resp))

		return "", nil, err
	}

	// 生成全局唯一的ChannelID
	id := fmt.Sprintf("%s_%s_%d", h.ServiceID, tk.Account, common.Seq.Next())
	logger.Infoln("channel id is:", id)
	// 设置请求的ChannelID和meta
	req.ChannelID = id
	req.WriteBody(&pkt.Session{
		Account:   tk.Account,
		ChannelID: id,
		GateID:    h.ServiceID,
		App:       tk.App,
		RemoteIP:  "127.0.0.1",
	})
	req.AddStringMeta(serv.MetaKeyApp, tk.App)
	req.AddStringMeta(serv.MetaKeyAccount, tk.Account)

	// 把login.转发给Login服务
	err = h.Container.Forward(common.SNLogin, req)
	if err != nil {
		return "", nil, err
	}

	return id, x.Meta{
		serv.MetaKeyApp:     tk.App,
		serv.MetaKeyAccount: tk.Account,
	}, nil
}
