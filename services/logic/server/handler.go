package server

import (
	x "X_IM"
	"X_IM/container"
	"X_IM/pkg/logger"
	"X_IM/wire/common"
	"X_IM/wire/pkt"
	"bytes"
	"errors"
	"google.golang.org/protobuf/proto"
	"strings"
	"time"
)

var log = logger.WithFields(logger.Fields{
	"service": common.SNChat,
	"pkg":     "server",
})

type Handler struct {
	r *x.Router
	//Redis中的会话管理
	cache      x.SessionStorage
	dispatcher *SvrDispatcher
}

func NewServHandler(r *x.Router, cache x.SessionStorage) *Handler {
	return &Handler{
		r:          r,
		dispatcher: &SvrDispatcher{},
		cache:      cache,
	}
}

func (h *Handler) Accept(conn x.Conn, timeout time.Duration) (string, x.Meta, error) {
	log.Infoln("accept")

	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", nil, err
	}

	var req pkt.InnerHandshakeReq
	_ = proto.Unmarshal(frame.GetPayload(), &req)
	log.Infoln("accept -- ", req.ServiceID)

	return req.ServiceID, nil, nil
}

func (h *Handler) Receive(agent x.Agent, payload []byte) {
	buf := bytes.NewBuffer(payload)
	packet, err := pkt.MustReadLogicPkt(buf)
	if err != nil {
		log.Error(err)
		return
	}

	var session *pkt.Session
	//登录包
	if packet.Command == common.CommandLoginSignIn {
		server, _ := packet.GetMeta(common.MetaDestServer)
		//登录后生成pkt.Session
		session = &pkt.Session{
			ChannelID: packet.ChannelID,
			GateID:    server.(string),
			Tags:      []string{"AutoGenerated"},
		}
	} else {
		//todo: to be optimized

		session, err = h.cache.Get(packet.ChannelID)
		//session不存在，需要重新连接并登录
		if errors.Is(err, x.ErrSessionNil) {
			_ = RespErr(agent, packet, pkt.Status_SessionNotFound)
			return
		} else if err != nil {
			_ = RespErr(agent, packet, pkt.Status_SystemException)
			return
		}
	}
	log.Debugf("recv a message from %s  %s", session, &packet.Header)
	err = h.r.Serve(packet, h.dispatcher, h.cache, session)
	if err != nil {
		log.Warn(err)
	}
}

func RespErr(agent x.Agent, p *pkt.LogicPkt, status pkt.Status) error {
	packet := pkt.NewFrom(&p.Header)
	packet.Status = status
	packet.Flag = pkt.Flag_Response

	packet.AddStringMeta(common.MetaDestChannels, p.Header.ChannelID)
	return container.Push(agent.ID(), packet)
}

type SvrDispatcher struct {
}

// Push 把多个channels数组合并成一个string
// 设置到消息包LogicPkt的Meta附加信息中，再传输给网关。
func (d *SvrDispatcher) Push(gateway string, channels []string, p *pkt.LogicPkt) error {
	p.AddStringMeta(common.MetaDestChannels, strings.Join(channels, ","))
	return container.Push(gateway, p)
}

// Disconnect default listener
func (h *Handler) Disconnect(id string) error {
	logger.Warnf("in services/logic/server/handler.go:Disconnect(): close event of %s", id)
	return nil
}
