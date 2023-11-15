package logic

import (
	x "X_IM"
	"X_IM/examples/dialer"
	"X_IM/pkg/wire/common"
	pkt2 "X_IM/pkg/wire/pkt"
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUserTalk(t *testing.T) {
	cli1, err := dialer.Login(wsurl, "test1")
	assert.Nil(t, err)
	if err != nil {
		return
	}
	cli2, err := dialer.Login(wsurl, "test2")
	assert.Nil(t, err)
	if err != nil {
		return
	}
	p := pkt2.New(common.CommandChatUserTalk, pkt2.WithDest("test2"))
	p.WriteBody(&pkt2.MessageReq{
		Type: 1,
		Body: "hello world",
	})
	err = cli1.Send(pkt2.Marshal(p))
	assert.Nil(t, err)

	// resp
	frame, _ := cli1.Read()
	assert.Equal(t, x.OpBinary, frame.GetOpCode())
	packet, err := pkt2.MustReadLogicPkt(bytes.NewBuffer(frame.GetPayload()))
	assert.Nil(t, err)
	assert.Equal(t, pkt2.Status_Success, packet.Header.Status)
	var resp pkt2.MessageResp
	_ = packet.ReadBody(&resp)
	assert.Greater(t, resp.MessageID, int64(1000))
	assert.Greater(t, resp.SendTime, int64(1000))
	t.Log(&resp)

	// push message
	frame, err = cli2.Read()
	assert.Nil(t, err)
	packet, err = pkt2.MustReadLogicPkt(bytes.NewBuffer(frame.GetPayload()))
	assert.Nil(t, err)
	var push pkt2.MessagePush
	_ = packet.ReadBody(&push)
	assert.Equal(t, resp.MessageID, push.MessageID)
	assert.Equal(t, resp.SendTime, push.SendTime)
	assert.Equal(t, "hello world", push.Body)
	assert.Equal(t, int32(1), push.Type)
	t.Log(&push)
}

func TestGroupTalk(t *testing.T) {
	// 1. test1 登录
	cli1, err := dialer.Login(wsurl, "test1")
	assert.Nil(t, err)

	// 2. 创建群
	p := pkt2.New(common.CommandGroupCreate)
	p.WriteBody(&pkt2.GroupCreateReq{
		Name:    "group1",
		Owner:   "test1",
		Members: []string{"test1", "test2", "test3", "test4"},
	})
	err = cli1.Send(pkt2.Marshal(p))
	assert.Nil(t, err)

	// 3. 读取创建群返回信息
	ack, err := cli1.Read()
	assert.Nil(t, err)
	ackp, _ := pkt2.MustReadLogicPkt(bytes.NewBuffer(ack.GetPayload()))
	assert.Equal(t, pkt2.Status_Success, ackp.GetStatus())
	assert.Equal(t, common.CommandGroupCreate, ackp.GetCommand())
	// 4. 解包
	var createResp pkt2.GroupCreateResp
	err = ackp.ReadBody(&createResp)
	assert.Nil(t, err)
	group := createResp.GetGroupID()
	assert.NotEmpty(t, group)
	if group == "" {
		return
	}
	// 5. 群成员test2、test3 登录
	cli2, err := dialer.Login(wsurl, "test2")
	assert.Nil(t, err)
	cli3, err := dialer.Login(wsurl, "test3")
	assert.Nil(t, err)
	t1 := time.Now()

	// 6. 发送群消息 CommandChatGroupTalk
	grpTalk := pkt2.New(common.CommandChatGroupTalk, pkt2.WithDest(group)).WriteBody(&pkt2.MessageReq{
		Type: 1,
		Body: "hello,group~",
	})
	err = cli1.Send(pkt2.Marshal(grpTalk))
	assert.Nil(t, err)
	// 7. 读取resp消息，确认消息发送成功
	ack, _ = cli1.Read()
	ackp, _ = pkt2.MustReadLogicPkt(bytes.NewBuffer(ack.GetPayload()))
	assert.Equal(t, pkt2.Status_Success, ackp.GetStatus())

	// 7. test2 读取消息
	notify1, _ := cli2.Read()
	n1, _ := pkt2.MustReadLogicPkt(bytes.NewBuffer(notify1.GetPayload()))
	assert.Equal(t, common.CommandChatGroupTalk, n1.GetCommand())
	var notify pkt2.MessagePush
	_ = n1.ReadBody(&notify)
	// 8. 校验消息内容
	assert.Equal(t, "hello,group~", notify.Body)
	assert.Equal(t, int32(1), notify.Type)
	assert.Empty(t, notify.Extra)
	assert.Greater(t, notify.SendTime, t1.UnixNano())
	assert.Greater(t, notify.MessageID, int64(10000))

	// 9. test3 读取消息
	notify2, _ := cli3.Read()
	n2, _ := pkt2.MustReadLogicPkt(bytes.NewBuffer(notify2.GetPayload()))
	_ = n2.ReadBody(&notify)
	assert.Equal(t, "hello,group~", notify.Body)

	t.Logf("cost %v", time.Since(t1))
}