package serv

import (
	x "X_IM"
	"X_IM/container"
	"X_IM/logger"
	"X_IM/wire/common"
	"X_IM/wire/pkt"
	"X_IM/wire/token"
	"bytes"
	"fmt"
	"regexp"
	"time"
)

var log = logger.WithFields(logger.Fields{
	"service": "gateway",
	"pkg":     "serv",
})

type Handler struct {
	ServiceID string
}

func (h *Handler) Accept(conn x.Conn, timeout time.Duration) (string, error) {
	log := logger.WithFields(logger.Fields{
		"ServiceID": h.ServiceID,
		"module":    "Handler",
		"handler":   "Accept",
	})
	log.Infoln("Accept")
	// 1. 读取客户端发送的握手数据（登录包）
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(frame.GetPayload())
	req, err := pkt.MustReadLogicPkt(buf)
	if err != nil {
		return "", err
	}
	//2. 必须为登录包
	if req.Command != common.CommandLoginSignIn {
		res := pkt.NewFrom(&req.Header)
		res.Status = pkt.Status_InvalidCommand
		_ = conn.WriteFrame(x.OpBinary, pkt.Marshal(res))

		return "", fmt.Errorf("is an \"InvalidCommand\" command")
	}
	//3.unmarshal Body
	var login pkt.LoginReq
	err = req.ReadBody(&login)
	if err != nil {
		return "", err
	}
	//4.use DefaultSecret to parse token
	tk, err := token.Parse(token.DefaultSecret, login.Token)
	if err != nil {
		//5.if token is invalid ,return Unauthenticated to SDK
		res := pkt.NewFrom(&req.Header)
		res.Status = pkt.Status_Unauthenticated
		_ = conn.WriteFrame(x.OpBinary, pkt.Marshal(res))

		return "", err
	}
	//6.生成全局唯一的ChannelID
	id := generateChannelID(h.ServiceID, tk.Account)

	req.ChannelID = id
	req.WriteBody(&pkt.Session{
		Account:   tk.Account,
		ChannelID: id,
		GateID:    id,
		App:       tk.App,
		RemoteIP:  getIP(conn.RemoteAddr().String()),
	})
	//7.将login转发给服务
	//todo:
	// 登陆服务与聊天服务在一个进程内部，主要是方便测试。如果改成login
	// 就要使用两个配置分别启动一个login和chat服务，因此这里就把它们合在一起了
	err = container.Forward(common.SNLogin, req)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (h *Handler) Receive(agent x.Agent, payload []byte) {
	buf := bytes.NewBuffer(payload)
	packet, err := pkt.Read(buf)
	if err != nil {
		log.Errorln(err)
		return
	}
	//处理心跳包
	if basicPkt, ok := packet.(*pkt.BasicPkt); ok {
		if basicPkt.Code == pkt.CodePing {
			_ = agent.Push(pkt.Marshal(&pkt.BasicPkt{Code: pkt.CodePong}))
		}
		return
	}
	//转发给逻辑服务处理
	if logicPkt, ok := packet.(*pkt.LogicPkt); ok {
		logicPkt.ChannelID = agent.ID()
		err = container.Forward(logicPkt.ServiceName(), logicPkt)
		if err != nil {
			logger.WithFields(logger.Fields{
				"module": "Handler",
				"id":     agent.ID(),
				"cmd":    logicPkt.Command,
				"dest":   logicPkt.Dest,
			}).Error(err)
		}
	}
}

// Disconnect 登出的逻辑SDK不需要发送协议包，正常断开连接或者心跳超时等情况时网关就会发出连接断开通知
func (h *Handler) Disconnect(id string) error {
	log.Infof("disconnect %s", id)

	logout := pkt.New(common.CommandLoginSignOut, pkt.WithChannel(id))
	err := container.Forward(common.SNLogin, logout)
	if err != nil {
		logger.WithFields(logger.Fields{
			"module": "handler",
			"id":     id,
		}).Error(err)
	}
	return nil
}

var ipExp = regexp.MustCompile(string("\\:[0-9]+$"))

func getIP(remoteAddr string) string {
	if remoteAddr == "" {
		return ""
	}
	return ipExp.ReplaceAllString(remoteAddr, "")
}

func generateChannelID(serviceID, account string) string {
	return fmt.Sprintf("%s_%s_%d", serviceID, account, common.Seq.Next())
}
