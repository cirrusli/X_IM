package handler

import (
	x "X_IM"
	"X_IM/services/logic/client"
	"X_IM/wire/pkt"
	"X_IM/wire/rpc"
	"errors"
)

type OfflineHandler struct {
	msgService client.Message
}

func NewOfflineHandler(message client.Message) *OfflineHandler {
	return &OfflineHandler{
		msgService: message,
	}
}

// DoSyncIndex 同步离线消息索引
func (h *OfflineHandler) DoSyncIndex(ctx x.Context) {
	var req pkt.MessageIndexReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	resp, err := h.msgService.GetMessageIndex(ctx.Session().GetApp(), &rpc.GetOfflineMessageIndexReq{
		Account:   ctx.Session().GetAccount(),
		MessageID: req.GetMessageID(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	var list = make([]*pkt.MessageIndex, len(resp.List))
	for i, val := range resp.List {
		list[i] = &pkt.MessageIndex{
			MessageID: val.MessageID,
			Direction: val.Direction,
			SendTime:  val.SendTime,
			AccountB:  val.AccountB,
			Group:     val.Group,
		}
	}
	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageIndexResp{
		Indexes: list,
	})
}

// DoSyncContent 同步离线消息内容
func (h *OfflineHandler) DoSyncContent(ctx x.Context) {
	var req pkt.MessageContentReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	if len(req.MessageIDs) == 0 {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, errors.New("empty MessageIds"))
		return
	}
	resp, err := h.msgService.GetMessageContent(ctx.Session().GetApp(), &rpc.GetOfflineMessageContentReq{
		MessageIDs: req.MessageIDs,
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	var list = make([]*pkt.MessageContent, len(resp.List))
	for i, val := range resp.List {
		list[i] = &pkt.MessageContent{
			MessageID: val.ID,
			Type:      val.Type,
			Body:      val.Body,
			Extra:     val.Extra,
		}
	}
	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageContentResp{
		Contents: list,
	})
}
