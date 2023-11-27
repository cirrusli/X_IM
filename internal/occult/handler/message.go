package handler

import (
	"X_IM/internal/occult/database"
	"X_IM/pkg/wire/common"
	"X_IM/pkg/wire/rpc"
	"github.com/go-redis/redis/v7"
	"github.com/kataras/iris/v12"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"time"
)

type ServiceHandler struct {
	BaseDB            *gorm.DB
	MessageDB         *gorm.DB
	Cache             *redis.Client
	IDGen             *database.IDGenerator
	groupSingleflight singleflight.Group
}

func (h *ServiceHandler) InsertUserMessage(c iris.Context) {
	var req rpc.InsertMessageReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}
	id, err := h.insertUserMessage(&req)
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	_, _ = c.Negotiate(&rpc.InsertMessageResp{
		MessageID: id,
	})
}

func (h *ServiceHandler) insertUserMessage(req *rpc.InsertMessageReq) (int64, error) {
	messageId := h.IDGen.Next().Int64()
	messageContent := database.MessageContent{
		ID: messageId,
		//ShardID:   database.HashCode(req.Dest),
		Type:     byte(req.Message.Type),
		Body:     req.Message.Body,
		Extra:    req.Message.Extra,
		SendTime: req.SendTime,
	}
	// 扩散写
	idxs := make([]database.MessageIndex, 2)
	idxs[0] = database.MessageIndex{
		ID: h.IDGen.Next().Int64(),
		//ShardID:   database.HashCode(req.Sender),
		MessageID: messageId,
		AccountA:  req.Dest,
		AccountB:  req.Sender,
		Direction: 0,
		SendTime:  req.SendTime,
	}
	idxs[1] = database.MessageIndex{
		ID: h.IDGen.Next().Int64(),
		//ShardID:   database.HashCode(m.Account),
		MessageID: messageId,
		AccountA:  req.Sender,
		AccountB:  req.Dest,
		Direction: 1,
		SendTime:  req.SendTime,
	}

	err := h.MessageDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&messageContent).Error; err != nil {
			return err
		}
		if err := tx.Create(&idxs).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return messageId, nil
}

func (h *ServiceHandler) InsertGroupMessage(c iris.Context) {
	var req rpc.InsertMessageReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}
	messageId, err := h.insertGroupMessage(&req)
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	_, _ = c.Negotiate(&rpc.InsertMessageResp{
		MessageID: messageId,
	})
}

func (h *ServiceHandler) insertGroupMessage(req *rpc.InsertMessageReq) (int64, error) {
	messageId := h.IDGen.Next().Int64()

	var members []database.GroupMember
	err := h.BaseDB.Where(&database.GroupMember{Group: req.Dest}).Find(&members).Error
	if err != nil {
		return 0, err
	}
	// 扩散写
	var idxs = make([]database.MessageIndex, len(members))
	for i, m := range members {
		idxs[i] = database.MessageIndex{
			ID:        h.IDGen.Next().Int64(),
			MessageID: messageId,
			AccountA:  m.Account,
			AccountB:  req.Sender,
			Direction: 0,
			Group:     m.Group,
			SendTime:  req.SendTime,
		}
		if m.Account == req.Sender {
			idxs[i].Direction = 1
		}
	}

	messageContent := database.MessageContent{
		ID:       messageId,
		Type:     byte(req.Message.Type),
		Body:     req.Message.Body,
		Extra:    req.Message.Extra,
		SendTime: req.SendTime,
	}

	err = h.MessageDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&messageContent).Error; err != nil {
			return err
		}
		if err := tx.Create(&idxs).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return messageId, nil
}

func (h *ServiceHandler) MessageACK(c iris.Context) {
	var req rpc.AckMessageReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}
	// 将读索引直接存到Redis
	err := setMessageAck(h.Cache, req.Account, req.MessageID)
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func setMessageAck(cache *redis.Client, account string, msgId int64) error {
	if msgId == 0 {
		return nil
	}
	key := database.KeyMessageAckIndex(account)
	return cache.Set(key, msgId, common.OfflineReadIndexExpiresIn).Err()
}

// GetOfflineMessageIndex
// 获取读索引的全局时钟
// 根据这个全局时钟，从DB中把消息索引读取出来
// 重置读索引位置
func (h *ServiceHandler) GetOfflineMessageIndex(c iris.Context) {
	var req rpc.GetOfflineMessageIndexReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}
	msgId := req.MessageID
	start, err := h.getSentTime(req.Account, req.MessageID)
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	var indexes []*rpc.MessageIndex
	tx := h.MessageDB.Model(&database.MessageIndex{}).
		Select("send_time", "account_b", "direction", "message_id", "group")

	err = tx.Where("account_a=? and send_time>? and direction=?", req.Account, start, 0).
		Order("send_time asc").Limit(common.OfflineSyncIndexCount).Find(&indexes).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	err = setMessageAck(h.Cache, req.Account, msgId)
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	_, _ = c.Negotiate(&rpc.GetOfflineMessageIndexResp{
		List: indexes,
	})
}

func (h *ServiceHandler) getSentTime(account string, msgID int64) (int64, error) {
	// 1. 冷启动情况，从服务端拉取消息索引
	if msgID == 0 {
		key := database.KeyMessageAckIndex(account)
		// 如果一次都没有发ack包，这里就是0
		msgID, _ = h.Cache.Get(key).Int64()
	}
	var start int64
	if msgID > 0 {
		// 2.根据消息ID读取此条消息的发送时间。
		var content database.MessageContent
		err := h.MessageDB.Select("send_time").First(&content, msgID).Error
		if err != nil {
			//3.如果此条消息不存在，返回最近一天
			start = time.Now().AddDate(0, 0, -1).UnixNano()
		} else {
			start = content.SendTime
		}
	}
	// 4.返回默认的离线消息过期时间
	earliestKeepTime := time.Now().
		AddDate(0, 0, -1*common.OfflineMessageExpiresIn).UnixNano()
	if start == 0 || start < earliestKeepTime {
		start = earliestKeepTime
	}
	return start, nil
}

// GetOfflineMessageContent 直接从DB中下载消息内容，限制每次的最大读取数
func (h *ServiceHandler) GetOfflineMessageContent(c iris.Context) {
	var req rpc.GetOfflineMessageContentReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}
	mlen := len(req.MessageIDs)
	if mlen > common.MessageMaxCountPerPage {
		c.StopWithText(iris.StatusBadRequest, "too many MessageIDs")
		return
	}
	var contents []*rpc.Message
	err := h.MessageDB.Model(&database.MessageContent{}).Where(req.MessageIDs).Find(&contents).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	_, _ = c.Negotiate(&rpc.GetOfflineMessageContentResp{
		List: contents,
	})
}
