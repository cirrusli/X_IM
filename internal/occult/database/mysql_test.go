package database

import (
	"X_IM/pkg/logger"
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"
)

var db *gorm.DB
var idgen *IDGenerator

const (
	dsn = "root:lzq@tcp(8.146.198.70:3306)/x_message?charset=utf8mb4&parseTime=True&loc=Local"
)

func init() {
	var err error
	db, err = InitDB("mysql", dsn)
	if err != nil {
		logger.Fatalln("in init(): ", err)
	}
	_ = db.AutoMigrate(&MessageIndex{}) //分区表时，不需要在此创建MessageIndex表
	_ = db.AutoMigrate(&MessageContent{})

	idgen, _ = NewIDGenerator(1)
}

func BenchmarkInsert(b *testing.B) {
	sendTime := time.Now().UnixNano()
	b.ResetTimer()
	b.SetBytes(1024)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idxs := make([]MessageIndex, 100)
			cid := idgen.Next().Int64()
			for i := 0; i < len(idxs); i++ {
				account := fmt.Sprintf("x_bench_%d", cid)
				idxs[i] = MessageIndex{
					ID: idgen.Next().Int64(),
					//ShardID:   HashCode(account),
					AccountA:  account,
					AccountB:  fmt.Sprintf("x_mark_%d", i),
					SendTime:  sendTime,
					MessageID: cid,
				}
			}
			db.Create(&idxs)
		}
	})
}

//MySQL分区表
//
//func TestShardingMessageInsert(t *testing.T) {
//	for i := 0; i < 10000; i++ {
//		sendTime := time.Now().UnixNano()
//		idxs := make([]MessageIndex, 100)
//		cid := idgen.Next().Int64()
//		for i := 0; i < len(idxs); i++ {
//			account := fmt.Sprintf("test_%d", cid)
//			idxs[i] = MessageIndex{
//				ShardID:   HashCode(account),
//				AccountA:  account,
//				AccountB:  fmt.Sprintf("test_%d", i),
//				SendTime:  sendTime,
//				MessageID: cid,
//			}
//		}
//		db.Create(&idxs)
//	}
//}
