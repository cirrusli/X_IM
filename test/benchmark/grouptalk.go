package benchmark

import (
	"X_IM/pkg/wire/common"
	"X_IM/pkg/wire/pkt"
	"X_IM/pkg/x"
	"X_IM/test/benchmark/report"
	"X_IM/test/mock/dialer"
	"bytes"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"os"
	"sync"
	"time"
)

func groupTalk(wsurl, appSecret string, threads, count int, memberCount int, onlinePercent float32) error {
	cli1, err := dialer.Login(wsurl, "test1", appSecret)
	if err != nil {
		return err
	}
	var members = make([]string, memberCount)
	for i := 0; i < memberCount; i++ {
		members[i] = fmt.Sprintf("test_%d", i+1)
	}
	// 创建群
	p := pkt.New(common.CommandGroupCreate)
	p.WriteBody(&pkt.GroupCreateReq{
		Name:    "group1",
		Owner:   "test1",
		Members: members,
	})
	if err = cli1.Send(pkt.Marshal(p)); err != nil {
		return err
	}
	// 读取返回信息
	ack, _ := cli1.Read()
	ackPkt, _ := pkt.MustReadLogicPkt(bytes.NewBuffer(ack.GetPayload()))
	if pkt.Status_Success != ackPkt.GetStatus() {
		return fmt.Errorf("create group failed")
	}
	var createResp pkt.GroupCreateResp
	_ = ackPkt.ReadBody(&createResp)
	group := createResp.GetGroupID()

	onlines := int(float32(memberCount) * onlinePercent)
	if onlines < 1 {
		onlines = 1
	}
	for i := 1; i < onlines; i++ {
		clix, err := dialer.Login(wsurl, fmt.Sprintf("test_%d", i), appSecret)
		if err != nil {
			return err
		}
		go func(cli x.Client) {
			for {
				_, err := cli.Read()
				if err != nil {
					return
				}
			}
		}(clix)
	}

	clis, err := loginMulti(wsurl, appSecret, 2, threads)
	if err != nil {
		return err
	}

	pool, _ := ants.NewPool(threads, ants.WithPreAlloc(true))
	defer pool.Release()

	r := report.New(os.Stdout, count)
	t1 := time.Now()

	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		cli := clis[i%threads]
		_ = pool.Submit(func() {
			defer func() {
				wg.Done()
			}()

			t0 := time.Now()
			p := pkt.New(common.CommandChatGroupTalk, pkt.WithDest(group))
			p.WriteBody(&pkt.MessageReq{
				Type: 1,
				Body: "hello world",
			})
			// 发送消息
			err := cli.Send(pkt.Marshal(p))
			if err != nil {
				r.Add(&report.Result{
					Err:           err,
					ContentLength: 11,
				})
				return
			}
			// 读取Resp消息
			_, err = cli.Read()
			if err != nil {
				r.Add(&report.Result{
					Err:           err,
					ContentLength: 11,
				})
				return
			}
			r.Add(&report.Result{
				Duration:   time.Since(t0),
				Err:        err,
				StatusCode: 0,
			})
		})
	}

	wg.Wait()
	r.Finalize(time.Since(t1))
	return nil
}
