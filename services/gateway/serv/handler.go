package serv

import (
	x "X_IM"
	"X_IM/container"
	"X_IM/pkg/logger"
	"X_IM/pkg/token"
	"X_IM/wire/common"
	"X_IM/wire/pkt"
	"bytes"
	"fmt"
	"regexp"
	"time"
)

const (
	MetaKeyApp     = "app"
	MetaKeyAccount = "account"
)

var log = logger.WithFields(logger.Fields{
	"service": "gateway",
	"pkg":     "serv",
})

type Handler struct {
	ServiceID string
	AppSecret string
}

// Accept this connection
func (h *Handler) Accept(conn x.Conn, timeout time.Duration) (string, x.Meta, error) {
	// 1. 读取登录包
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", nil, err
	}

	buf := bytes.NewBuffer(frame.GetPayload())
	req, err := pkt.MustReadLogicPkt(buf)
	if err != nil {
		log.Error(err)
		return "", nil, err
	}
	// 2. 必须是登录包
	if req.Command != common.CommandLoginSignIn {
		resp := pkt.NewFrom(&req.Header)
		resp.Status = pkt.Status_InvalidCommand
		_ = conn.WriteFrame(x.OpBinary, pkt.Marshal(resp))

		return "", nil, fmt.Errorf("is an \"InvalidCommand\" command")
	}
	//3.unmarshal Body
	var login pkt.LoginReq
	err = req.ReadBody(&login)
	if err != nil {
		return "", nil, err
	}
	secret := h.AppSecret
	if secret == "" {
		secret = token.DefaultSecret
	}
	// 4. 使用默认的DefaultSecret 解析token
	tk, err := token.Parse(secret, login.Token)
	if err != nil {
		//5.if token is invalid ,return Unauthenticated to SDK
		resp := pkt.NewFrom(&req.Header)
		resp.Status = pkt.Status_Unauthenticated
		_ = conn.WriteFrame(x.OpBinary, pkt.Marshal(resp))

		return "", nil, err
	}
	//6.生成全局唯一的ChannelID
	id := generateChannelID(h.ServiceID, tk.Account)
	log.Infof("accept %v channel:%s", tk, id)

	req.ChannelID = id
	req.WriteBody(&pkt.Session{
		Account:   tk.Account,
		ChannelID: id,
		GateID:    h.ServiceID,
		App:       tk.App,
		RemoteIP:  getIP(conn.RemoteAddr().String()),
	})
	req.AddStringMeta(MetaKeyApp, tk.App)
	req.AddStringMeta(MetaKeyAccount, tk.Account)

	// 7. 把login.转发给Login服务
	err = container.Forward(common.SNLogin, req)
	if err != nil {
		log.Errorf("container.Forward :%v", err)
		return "", nil, err
	}
	return id, x.Meta{
		MetaKeyApp:     tk.App,
		MetaKeyAccount: tk.Account,
	}, nil
}

// Receive default listener
func (h *Handler) Receive(ag x.Agent, payload []byte) {
	buf := bytes.NewBuffer(payload)
	packet, err := pkt.Read(buf)
	if err != nil {
		log.Errorln(err)
		return
	}
	//处理心跳包
	if basicPkt, ok := packet.(*pkt.BasicPkt); ok {
		if basicPkt.Code == pkt.CodePing {
			_ = ag.Push(pkt.Marshal(&pkt.BasicPkt{Code: pkt.CodePong}))
		}
		return
	}
	//转发给逻辑服务处理
	if logicPkt, ok := packet.(*pkt.LogicPkt); ok {
		logicPkt.ChannelID = ag.ID()

		messageInTotal.WithLabelValues(h.ServiceID, common.SNTGateway, logicPkt.Command).Inc()
		messageInFlowBytes.WithLabelValues(h.ServiceID, common.SNTGateway, logicPkt.Command).Add(float64(len(payload)))

		// 把meta注入到header中
		if ag.GetMeta() != nil {
			logicPkt.AddStringMeta(MetaKeyApp, ag.GetMeta()[MetaKeyApp])
			logicPkt.AddStringMeta(MetaKeyAccount, ag.GetMeta()[MetaKeyAccount])
		}

		err = container.Forward(logicPkt.ServiceName(), logicPkt)
		if err != nil {
			logger.WithFields(logger.Fields{
				"module": "handler",
				"id":     ag.ID(),
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
