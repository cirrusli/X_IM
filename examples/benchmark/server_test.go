package benchmark

import (
	x "X_IM"
	"X_IM/examples/mock"
	"X_IM/pkg/logger"
	"X_IM/pkg/websocket"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"strings"
	"sync"
	"testing"
	"time"
)

const wsurl = "ws://localhost:8000"

func TestParallel(t *testing.T) {
	const doConns = 10000
	gpool, _ := ants.NewPool(50, ants.WithPreAlloc(true))
	defer gpool.Release()
	var wg sync.WaitGroup
	wg.Add(doConns)

	clis := make([]x.Client, doConns)
	t0 := time.Now()
	for i := 0; i < doConns; i++ {
		idx := i
		_ = gpool.Submit(func() {
			cli := websocket.NewClient(fmt.Sprintf("test_%v", idx), "client", websocket.ClientOptions{
				Heartbeat: x.DefaultHeartbeat,
			})
			// set dialer
			cli.SetDialer(&mock.WebsocketDialer{})

			// step2: 建立连接
			err := cli.Connect(wsurl)
			if err != nil {
				logger.Error(err)
			}
			clis[idx] = cli
			wg.Done()
		})
	}
	wg.Wait()
	t.Logf("logined %d cost %v", doConns, time.Since(t0))
	t.Logf("done connecting")
	time.Sleep(time.Second * 5)
	t.Logf("closed")

	for i := 0; i < doConns; i++ {
		clis[i].Close()
	}
}

func TestMessage(t *testing.T) {
	const msgs = 1000 * 100
	cli := websocket.NewClient(fmt.Sprintf("test_%v", 1), "client", websocket.ClientOptions{
		Heartbeat: x.DefaultHeartbeat,
	})
	// set dialer
	cli.SetDialer(&mock.WebsocketDialer{})

	// step2: 建立连接
	err := cli.Connect(wsurl)
	if err != nil {
		logger.Error(err)
	}
	msg := []byte(strings.Repeat("hello", 10))
	t0 := time.Now()
	go func() {
		for i := 0; i < msgs; i++ {
			_ = cli.Send(msg)
		}
	}()
	recv := 0
	for {
		frame, err := cli.Read()
		if err != nil {
			logger.Info("time", time.Now().UnixNano(), err)
			break
		}
		if frame.GetOpCode() != x.OpBinary {
			continue
		}
		recv++
		if recv == msgs { // 接收完消息
			break
		}
	}

	t.Logf("message %d cost %v", msgs, time.Since(t0))
}
