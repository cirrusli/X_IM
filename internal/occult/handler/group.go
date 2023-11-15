package handler

import (
	"X_IM/internal/occult/database"
	"X_IM/pkg/logger"
	"X_IM/pkg/wire/rpc"
	"errors"
	"github.com/bwmarrin/snowflake"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

func (h *ServiceHandler) GroupCreate(c iris.Context) {
	app := c.Params().Get("app")
	var req rpc.CreateGroupReq
	// 根据请求头中的Content-Type解包
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}
	req.App = app
	groupId, err := h.groupCreate(&req)
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	// 根据请求头中的Accept格式来序列化返回包
	_, _ = c.Negotiate(&rpc.CreateGroupResp{
		// base36生成的string可读性高一点，没有大小写的区别
		GroupID: groupId.Base36(),
	})
}

func (h *ServiceHandler) groupCreate(req *rpc.CreateGroupReq) (snowflake.ID, error) {
	groupId := h.IDGen.Next()
	g := &database.Group{
		Model: database.Model{
			ID: groupId.Int64(),
		},
		App:          req.App,
		Group:        groupId.Base36(),
		Name:         req.Name,
		Avatar:       req.Avatar,
		Owner:        req.Owner,
		Introduction: req.Introduction,
	}
	members := make([]database.GroupMember, len(req.Members))
	for i, user := range req.Members {
		members[i] = database.GroupMember{
			Model: database.Model{
				ID: h.IDGen.Next().Int64(),
			},
			Account: user,
			Group:   groupId.Base36(),
		}
	}
	// 使用事务，保证数据的一致性
	err := h.BaseDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(g).Error; err != nil {
			// return any will rollback
			return err
		}
		if err := tx.Create(&members).Error; err != nil {
			return err
		}
		// return nil will commit the whole transaction
		return nil
	})
	if err != nil {
		return 0, err
	}
	return groupId, nil
}

func (h *ServiceHandler) GroupJoin(c iris.Context) {
	// app := c.Param("app")
	var req rpc.JoinGroupReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}
	gm := &database.GroupMember{
		Model: database.Model{
			ID: h.IDGen.Next().Int64(),
		},
		Account: req.Account,
		Group:   req.GroupID,
	}
	err := h.BaseDB.Create(gm).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func (h *ServiceHandler) GroupQuit(c iris.Context) {
	// app := c.Param("app")
	var req rpc.QuitGroupReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}
	gm := &database.GroupMember{
		Account: req.Account,
		Group:   req.GroupID,
	}
	err := h.BaseDB.Delete(&database.GroupMember{}, gm).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func (h *ServiceHandler) GroupMembers(c iris.Context) {
	group := c.Params().Get("id")
	if group == "" {
		c.StopWithError(iris.StatusBadRequest, errors.New("group is null"))
		return
	}
	var members []database.GroupMember
	err := h.BaseDB.Order("Updated_At asc").Find(&members, database.GroupMember{Group: group}).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	var users = make([]*rpc.Member, len(members))
	for i, m := range members {
		users[i] = &rpc.Member{
			Account:  m.Account,
			Alias:    m.Alias,
			JoinTime: m.CreatedAt.Unix(),
		}
	}
	_, _ = c.Negotiate(&rpc.GroupMembersResp{
		Users: users,
	})
}

func (h *ServiceHandler) GroupGet(c iris.Context) {
	groupID := c.Params().Get("id")
	if groupID == "" {
		c.StopWithError(iris.StatusBadRequest, errors.New("no group id"))
		return
	}

	result, err, _ := h.groupSingleflight.Do(groupID, func() (interface{}, error) {
		id, err := h.IDGen.ParseBase36(groupID)
		if err != nil {
			return nil, errors.New("group is invalid:" + groupID)
		}
		var group database.Group
		logger.Debugln("Starting database query for group:", groupID)
		err = h.BaseDB.First(&group, id.Int64()).Error
		logger.Debugln("Finished database query for group:", groupID)
		if err != nil {
			return nil, err
		}
		return &rpc.GetGroupResp{
			ID:           groupID,
			Name:         group.Name,
			Avatar:       group.Avatar,
			Introduction: group.Introduction,
			Owner:        group.Owner,
			CreatedAt:    group.CreatedAt.Unix(),
		}, nil
	})

	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	_, _ = c.Negotiate(result)
}
