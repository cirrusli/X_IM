package handler

import (
	x "X_IM"
	"X_IM/services/logic/client"
	"X_IM/wire/pkt"
	"X_IM/wire/rpc"
	"errors"
	"time"
)

var ErrNoDestination = errors.New("no destination")

type ChatHandler struct {
	msgService   client.Message
	groupService client.Group
}

func NewChatHandler(msg client.Message, group client.Group) *ChatHandler {
	return &ChatHandler{
		msgService:   msg,
		groupService: group,
	}
}

// DoSingleTalk 单聊
func (h *ChatHandler) DoSingleTalk(ctx x.Context) {
	if ctx.Header().Dest == "" {
		_ = ctx.RespWithError(pkt.Status_NoDestination, ErrNoDestination)
		return
	}
	// 1. 解包
	var req pkt.MessageReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	// 2. 获取接收方的位置信息
	receiver := ctx.Header().GetDest()
	loc, err := ctx.GetLocation(receiver, "")
	if err != nil && errors.Is(err, x.ErrSessionNil) {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	// 3. 保存离线消息
	sendTime := time.Now().UnixNano()
	resp, err := h.msgService.InsertUser(ctx.Session().GetApp(), &rpc.InsertMessageReq{
		Sender:   ctx.Session().GetAccount(),
		Dest:     receiver,
		SendTime: sendTime,
		Message: &rpc.Message{
			Type:  req.GetType(),
			Body:  req.GetBody(),
			Extra: req.GetExtra(),
		},
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	msgID := resp.MessageID

	//4.if receiver is online,send the message to receiver
	if loc != nil {
		if err = ctx.Dispatch(&pkt.MessagePush{
			MessageID: msgID,
			Type:      req.GetType(),
			Body:      req.GetBody(),
			Extra:     req.GetExtra(),
			Sender:    ctx.Session().GetAccount(),
			SendTime:  sendTime,
		}, loc); err != nil {
			_ = ctx.RespWithError(pkt.Status_SystemException, err)
			return
		}
	}
	// 5. 返回一条resp消息
	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageResp{
		MessageID: msgID,
		SendTime:  sendTime,
	})
}

// DoGroupTalk 群聊
func (h *ChatHandler) DoGroupTalk(ctx x.Context) {
	if ctx.Header().GetDest() == "" {
		_ = ctx.RespWithError(pkt.Status_NoDestination, ErrNoDestination)
		return
	}
	// 1. 解包
	var req pkt.MessageReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	// 群聊里dest就不再是user account，而是群ID
	group := ctx.Header().GetDest()
	sendTime := time.Now().UnixNano()

	// 2. 保存离线消息
	resp, err := h.msgService.InsertGroup(ctx.Session().GetApp(), &rpc.InsertMessageReq{
		Sender:   ctx.Session().GetAccount(),
		Dest:     group,
		SendTime: sendTime,
		Message: &rpc.Message{
			Type:  req.GetType(),
			Body:  req.GetBody(),
			Extra: req.GetExtra(),
		},
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	// 3. 读取群成员列表
	membersResp, err := h.groupService.Members(ctx.Session().GetApp(), &rpc.GroupMembersReq{
		GroupID: group,
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	var members = make([]string, len(membersResp.Users))
	for i, user := range membersResp.Users {
		members[i] = user.Account
	}
	// 4. 批量寻址（群成员）
	locs, err := ctx.GetLocations(members...)
	if err != nil && !errors.Is(err, x.ErrSessionNil) {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	// 5. 批量推送消息给成员
	if len(locs) > 0 {
		if err = ctx.Dispatch(&pkt.MessagePush{
			MessageID: resp.MessageID,
			Type:      req.GetType(),
			Body:      req.GetBody(),
			Extra:     req.GetExtra(),
			Sender:    ctx.Session().GetAccount(),
			SendTime:  sendTime,
		}, locs...); err != nil {
			_ = ctx.RespWithError(pkt.Status_SystemException, err)
			return
		}
	}
	// 6. 返回一条resp消息
	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageResp{
		MessageID: resp.MessageID,
		SendTime:  sendTime,
	})
}

func (h *ChatHandler) DoTalkAck(ctx x.Context) {
	var req pkt.MessageAckReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	//
	err := h.msgService.SetACK(ctx.Session().GetApp(), &rpc.AckMessageReq{
		Account:   ctx.Session().GetAccount(),
		MessageID: req.GetMessageID(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	_ = ctx.Resp(pkt.Status_Success, nil)
}
