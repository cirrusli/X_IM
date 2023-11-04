package handler

import (
	"X_IM/services/occult/database"
	"X_IM/wire/rpc"
	"fmt"
	"testing"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
)

var handler ServiceHandler

func init() {
	baseDb, _ := database.InitDB("mysql", "root:lzq@tcp(127.0.0.1:3306)/x_base?charset=utf8mb4&parseTime=True&loc=Local")
	messageDb, _ := database.InitDB("mysql", "root:lzq@tcp(127.0.0.1:3306)/x_message?charset=utf8mb4&parseTime=True&loc=Local")
	idgen, _ := database.NewIDGenerator(1)
	handler = ServiceHandler{
		MessageDB: messageDb,
		BaseDB:    baseDb,
		IDGen:     idgen,
	}
}

func BenchmarkInsUsrMsg(b *testing.B) {

	b.ResetTimer()
	b.SetBytes(1024)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = handler.insertUserMessage(&rpc.InsertMessageReq{
				Sender:   "test1",
				Dest:     ksuid.New().String(),
				SendTime: time.Now().UnixNano(),
				Message: &rpc.Message{
					Type: 1,
					Body: "hello",
				},
			})
		}
	})
}

func BenchmarkInsGrp10Msg(b *testing.B) {
	memberCount := 10

	var members = make([]string, memberCount)
	for i := 0; i < memberCount; i++ {
		members[i] = fmt.Sprintf("test%d", i+1)
	}

	groupId, err := handler.groupCreate(&rpc.CreateGroupReq{
		App:     "x_t",
		Name:    "testg",
		Owner:   "test1",
		Members: members,
	})
	assert.Nil(b, err)

	b.ResetTimer()
	b.SetBytes(1024)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = handler.insertGroupMessage(&rpc.InsertMessageReq{
				Sender:   "test1",
				Dest:     groupId.Base36(),
				SendTime: time.Now().UnixNano(),
				Message: &rpc.Message{
					Type: 1,
					Body: "hello",
				},
			})
		}
	})
}

func BenchmarkInsGrp50Msg(b *testing.B) {
	memberCount := 50

	var members = make([]string, memberCount)
	for i := 0; i < memberCount; i++ {
		members[i] = fmt.Sprintf("test%d", i+1)
	}

	groupId, err := handler.groupCreate(&rpc.CreateGroupReq{
		App:     "x_t",
		Name:    "testg",
		Owner:   "test1",
		Members: members,
	})
	assert.Nil(b, err)

	b.ResetTimer()
	b.SetBytes(1024)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = handler.insertGroupMessage(&rpc.InsertMessageReq{
				Sender:   "test1",
				Dest:     groupId.Base36(),
				SendTime: time.Now().UnixNano(),
				Message: &rpc.Message{
					Type: 1,
					Body: "hello",
				},
			})
		}
	})
}
