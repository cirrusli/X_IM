package ut

import (
	x "X_IM"
	mymock "X_IM/examples/mock"
	"X_IM/pkg/logger"
	"X_IM/pkg/token"
	"X_IM/pkg/wire/common"
	"X_IM/pkg/wire/pkt"
	"X_IM/services/gateway/serv"
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net"
	"testing"
	"time"
)

var (
	loginReq      = &pkt.LoginReq{}
	logicReqBytes []byte
)

const tokenSecret = "test_secret"

func init() {
	// 生成一个JWT token
	mytoken := &token.Token{
		Account: "test1",
		App:     "x_im",
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
	logicReqBytes = pkt.Marshal(logicPkt)
	logger.Infoln("init(): after pkt.Marshal,with magic code: ", logicReqBytes)
}
func TestHandlerAccept(t *testing.T) {
	// 创建一个模拟的Conn对象
	conn := new(mymock.Conn)

	// 设置期望的行为
	conn.On("SetReadDeadline", mock.Anything).Return(nil)
	conn.On("ReadFrame").Return(x.OpBinary, logicReqBytes, nil) // 替换为实际的登录包数据
	conn.On("WriteFrame", x.OpBinary, mock.Anything).Return(nil)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080})

	// 创建一个Handler对象
	handler := &serv.Handler{
		ServiceID: "test_gateway01",
		AppSecret: tokenSecret,
	}

	// 调用Accept方法
	id, meta, err := handler.Accept(conn, time.Second)

	// 验证结果
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.Equal(t, "test_app", meta[serv.MetaKeyApp])
	assert.Equal(t, "test_account", meta[serv.MetaKeyAccount])

	// 验证是否所有的期望都已满足
	conn.AssertExpectations(t)
}
