package database

import (
	"fmt"
	"log"
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
	db, _ = InitDB("mysql", dsn)

	err := db.AutoMigrate(&MessageIndex{})
	if err != nil {
		log.Fatalln("in init(): ", err)
	}
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
				idxs[i] = MessageIndex{
					ID:        idgen.Next().Int64(),
					AccountA:  fmt.Sprintf("x_bench_%d", cid),
					AccountB:  fmt.Sprintf("x_mark_%d", i),
					SendTime:  sendTime,
					MessageID: cid,
				}
			}
			db.Create(&idxs)
		}
	})
}
