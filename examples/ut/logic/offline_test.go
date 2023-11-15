package logic

import (
	x "X_IM"
	"X_IM/examples/dialer"
	"X_IM/pkg/logger"
	"X_IM/pkg/wire/common"
	pkt2 "X_IM/pkg/wire/pkt"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"testing"
	"time"
)

func TestOffline(t *testing.T) {
	src := fmt.Sprintf("u%d", time.Now().Unix())
	cli, err := dialer.Login(wsurl, src)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	dest := fmt.Sprintf("u%d", time.Now().Unix()+1)
	count := 10
	for i := 0; i < count; i++ {
		p := pkt2.New(common.CommandChatUserTalk, pkt2.WithDest(dest))
		p.WriteBody(&pkt2.MessageReq{
			Type: 1,
			Body: "hello world",
		})
		err := cli.Send(pkt2.Marshal(p))
		if err != nil {
			logger.Error(err)
			return
		}
		// wait ack
		_, _ = cli.Read()
	}

	destCli, err := dialer.Login(wsurl, dest)
	assert.Nil(t, err)

	// request offline message index
	p := pkt2.New(common.CommandOfflineIndex)
	p.WriteBody(&pkt2.MessageIndexReq{})
	_ = destCli.Send(pkt2.Marshal(p))

	var indexResp pkt2.MessageIndexResp
	err = Read(destCli, &indexResp)
	assert.Nil(t, err)

	assert.Equal(t, count, len(indexResp.Indexes))
	assert.Equal(t, src, indexResp.Indexes[0].AccountB)
	assert.Equal(t, int32(0), indexResp.Indexes[0].Direction)
	t.Log(indexResp.Indexes)

	var ids = make([]int64, count)
	for i, idx := range indexResp.Indexes {
		ids[i] = idx.MessageID
	}
	t.Log(ids)

	lastMessageId := ids[count-1]

	// read again
	p = pkt2.New(common.CommandOfflineIndex)
	p.WriteBody(&pkt2.MessageIndexReq{
		MessageID: lastMessageId,
	})
	_ = destCli.Send(pkt2.Marshal(p))

	var indexResp2 pkt2.MessageIndexResp
	err = Read(destCli, &indexResp2)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(indexResp2.Indexes))

	// request offline message content
	p = pkt2.New(common.CommandOfflineContent)
	p.WriteBody(&pkt2.MessageContentReq{
		MessageIDs: ids,
	})
	_ = destCli.Send(pkt2.Marshal(p))
	var contentResp pkt2.MessageContentResp
	err = Read(destCli, &contentResp)
	assert.Nil(t, err)
	t.Log(contentResp.Contents)
	assert.Equal(t, count, len(contentResp.Contents))
	assert.Equal(t, "hello world", contentResp.Contents[0].Body)
	assert.Equal(t, int32(1), contentResp.Contents[0].Type)
}

func Read(cli x.Client, body proto.Message) error {
	frame, err := cli.Read()
	if err != nil {
		return err
	}
	packet, _ := pkt2.MustReadLogicPkt(bytes.NewBuffer(frame.GetPayload()))
	if packet.GetStatus() != pkt2.Status_Success {
		return fmt.Errorf("received status :%v", packet.GetStatus())
	}
	return packet.ReadBody(body)
}
