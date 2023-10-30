package handler

import (
	x "X_IM"
	"X_IM/logger"
	"X_IM/wire/pkt"
	"errors"
)

type LoginHandler struct {
}

func NewLoginHandler() *LoginHandler {
	return &LoginHandler{}
}

func (h *LoginHandler) DoLogin(ctx x.Context) {
	var session pkt.Session
	if err := ctx.ReadBody(&session); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	logger.WithFields(logger.Fields{
		"Func":      "Login",
		"ChannelID": session.GetChannelID(),
		"Account":   session.GetAccount(),
		"RemoteIP":  session.GetRemoteIP(),
	}).Info("do Login")

	//检测当前账号是否已在其他地方登录
	old, err := ctx.GetLocation(session.Account, "")
	if err != nil && !errors.Is(err, x.ErrSessionNil) {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	if old != nil {
		//将旧的连接关闭
		_ = ctx.Dispatch(&pkt.KickOutNotify{
			//在web客户端由于网络IO都是异步操作，并且会自动重连
			//不能使用account作为唯一标识，否则很容易导致自己踢自己下线
			ChannelID: old.ChannelID,
		}, old)
	}

	//将新的连接加入到sessionStorage中
	err = ctx.Add(&session)
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	//return login succeed
	var resp = &pkt.LoginResp{
		ChannelID: session.ChannelID,
	}
	_ = ctx.Resp(pkt.Status_Success, resp)
}

func (h *LoginHandler) DoLogout(ctx x.Context) {
	logger.WithFields(logger.Fields{
		"Func":      "Logout",
		"ChannelID": ctx.Session().GetChannelID(),
		"Account":   ctx.Session().GetAccount(),
	}).Info("do Logout ")

	err := ctx.Delete(ctx.Session().GetAccount(), ctx.Session().GetChannelID())
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	_ = ctx.Resp(pkt.Status_Success, nil)
}
