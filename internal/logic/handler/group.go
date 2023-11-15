package handler

import (
	x "X_IM"
	"X_IM/internal/logic/client"
	"X_IM/pkg/wire/pkt"
	"X_IM/pkg/wire/rpc"
	"errors"
)

type GroupHandler struct {
	groupService client.Group
}

func NewGroupHandler(groupService client.Group) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
	}
}

func (h *GroupHandler) DoCreate(ctx x.Context) {
	var req pkt.GroupCreateReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	resp, err := h.groupService.Create(ctx.Session().GetApp(), &rpc.CreateGroupReq{
		Name:         req.GetName(),
		Avatar:       req.GetAvatar(),
		Introduction: req.GetIntroduction(),
		Owner:        req.GetOwner(),
		Members:      req.GetMembers(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	locs, err := ctx.GetLocations(req.GetMembers()...)
	if err != nil && !errors.Is(err, x.ErrSessionNil) {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	// push to receiver
	if len(locs) > 0 {
		if err = ctx.Dispatch(&pkt.GroupCreateNotify{
			GroupID: resp.GroupID,
			Members: req.GetMembers(),
		}, locs...); err != nil {
			_ = ctx.RespWithError(pkt.Status_SystemException, err)
			return
		}
	}

	_ = ctx.Resp(pkt.Status_Success, &pkt.GroupCreateResp{
		GroupID: resp.GroupID,
	})
}

func (h *GroupHandler) DoJoin(ctx x.Context) {
	var req pkt.GroupJoinReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	err := h.groupService.Join(ctx.Session().GetApp(), &rpc.JoinGroupReq{
		Account: req.Account,
		GroupID: req.GetGroupID(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	_ = ctx.Resp(pkt.Status_Success, nil)
}

func (h *GroupHandler) DoQuit(ctx x.Context) {
	var req pkt.GroupQuitReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	err := h.groupService.Quit(ctx.Session().GetApp(), &rpc.QuitGroupReq{
		Account: req.Account,
		GroupID: req.GetGroupID(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	_ = ctx.Resp(pkt.Status_Success, nil)
}

// DoDetail 获取群基本信息及成员列表
func (h *GroupHandler) DoDetail(ctx x.Context) {
	var req pkt.GroupGetReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	resp, err := h.groupService.Detail(ctx.Session().GetApp(), &rpc.GetGroupReq{
		GroupID: req.GetGroupID(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	membersResp, err := h.groupService.Members(ctx.Session().GetApp(), &rpc.GroupMembersReq{
		GroupID: req.GetGroupID(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	var members = make([]*pkt.Member, len(membersResp.GetUsers()))
	for i, m := range membersResp.GetUsers() {
		members[i] = &pkt.Member{
			Account:  m.Account,
			Alias:    m.Alias,
			JoinTime: m.JoinTime,
			Avatar:   m.Avatar,
		}
	}
	_ = ctx.Resp(pkt.Status_Success, &pkt.GroupGetResp{
		Id:           resp.ID,
		Name:         resp.Name,
		Introduction: resp.Introduction,
		Avatar:       resp.Avatar,
		Owner:        resp.Owner,
		Members:      members,
	})
}
