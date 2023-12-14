package benchmark

import (
	"X_IM/examples/benchmark/report"
	"X_IM/examples/mock/dialer"
	"X_IM/pkg/logger"
	"X_IM/pkg/x"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"os"
	"sync"
	"time"
)

func login(wsurl, appSecret string, threads int, count int, keep time.Duration) error {
	p, _ := ants.NewPool(threads, ants.WithPreAlloc(true))
	defer p.Release()

	r := report.New(os.Stdout, count)
	t1 := time.Now()

	var wg sync.WaitGroup
	wg.Add(count)
	clis := make([]x.Client, count)
	for i := 0; i < count; i++ {
		idx := i
		_ = p.Submit(func() {
			t0 := time.Now()
			cli, err := dialer.Login(wsurl, fmt.Sprintf("test%d", idx+1), appSecret)
			r.Add(&report.Result{
				Duration:   time.Since(t0),
				Err:        err,
				StatusCode: 0,
			})
			if err != nil {
				logger.Error(err)
			} else {
				clis[idx] = cli
			}
			wg.Done()
		})
	}
	wg.Wait()

	r.Finalize(time.Since(t1))

	logger.Infof("keep login for %v", keep)
	time.Sleep(keep)

	for _, cli := range clis {
		cli.Close()
	}
	logger.Infoln("shutdown..")
	return nil
}
