package benchmark

import (
	"X_IM/examples/benchmark/report"
	"X_IM/examples/mock/dialer"
	"X_IM/pkg/wire/common"
	"X_IM/pkg/wire/pkt"
	"X_IM/pkg/x"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"os"
	"sync"
	"time"
)

func loginMulti(wsurl, appSecret string, start, count int) ([]x.Client, error) {
	clis := make([]x.Client, count)
	for i := 0; i < count; i++ {
		account := fmt.Sprintf("test%d", start)
		start++
		cli, err := dialer.Login(wsurl, account, appSecret)
		if err != nil {
			return nil, err
		}
		clis[i] = cli
	}
	return clis, nil
}

func singleTalk(wsurl, appSecret string, threads, count int, online bool) error {
	p, _ := ants.NewPool(threads, ants.WithPreAlloc(true))
	defer p.Release()

	if online {
		cli2, _ := dialer.Login(wsurl, "test1")

		go func() {
			for {
				_, err := cli2.Read()
				if err != nil {
					return
				}
			}
		}()
	}

	clis, err := loginMulti(wsurl, appSecret, 2, threads)
	if err != nil {
		return err
	}

	r := report.New(os.Stdout, count)
	t1 := time.Now()

	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		cli := clis[i%threads]
		_ = p.Submit(func() {
			defer func() {
				wg.Done()
			}()

			t0 := time.Now()
			p := pkt.New(common.CommandChatUserTalk, pkt.WithDest("test1"))
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
