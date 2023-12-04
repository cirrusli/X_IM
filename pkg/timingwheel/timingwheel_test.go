package timingwheel

import (
	"sync"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	tw := NewTimingWheel(1*time.Second, 30)
	tw.start()
	defer tw.stop()

	done := make(chan struct{})

	timer := &Timer{
		expiration: timeToMs(time.Now().UTC().Add(1 * time.Second)),
		task:       func() {},
	}
	if !tw.add(timer) {
		t.Error("Failed to add timer, TimingWheel may not be running")
	}

	select {
	case <-done:
		// timer的task已经执行，说明TimingWheel已经成功启动
	case <-time.After(2 * time.Second):
		// 如果在2秒后还没有接收到信号，那么测试失败
		t.Error("Timer task not executed, TimingWheel may not have started")
	}
}

func TestStop(t *testing.T) {
	tw := NewTimingWheel(1*time.Second, 30)
	tw.start()

	timer := &Timer{
		expiration: timeToMs(time.Now().UTC().Add(1 * time.Second)),
		task:       func() {},
	}
	tw.add(timer)

	tw.stop()

	// 检查 exitC 通道是否已经被关闭，如果已关闭可以立即接收到false
	select {
	case _, ok := <-tw.exitC:
		if ok {
			//channel中接收到了值
			t.Error("exitC channel is not closed, TimingWheel may not have stopped")
		}
	default:
		//channel未关闭且没有值
		t.Error("exitC channel is not readable, TimingWheel may not have stopped")
	}
}

func TestAfterFunc(t *testing.T) {
	tw := NewTimingWheel(1*time.Second, 30)
	tw.start()
	defer tw.stop()

	var wg sync.WaitGroup
	wg.Add(1)

	// 创建一个定时任务，该任务在 1 秒后执行并通知 WaitGroup
	tw.afterFunc(1*time.Second, func() {
		wg.Done()
	})

	// 创建一个计时器，如果在 2 秒后 WaitGroup 还没有被通知，那么测试失败
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 定时任务已经执行，这是预期的结果
	case <-time.After(2 * time.Second):
		// 如果在 2 秒后定时任务还没有执行，那么测试失败
		t.Error("Scheduled function not executed")
	}
}

type ScheduleFuncTest struct {
	NextFunc func(time.Time) time.Time
}

func (s *ScheduleFuncTest) Next(t time.Time) time.Time {
	return s.NextFunc(t)
}

func TestScheduleFunc(t *testing.T) {
	tw := NewTimingWheel(1*time.Second, 30)
	tw.start()
	defer tw.stop()

	flag := false

	// 创建一个调度器，它总是在当前时间的 1 秒后安排下一次执行
	scheduler := &ScheduleFuncTest{
		NextFunc: func(time.Time) time.Time {
			return time.Now().Add(1 * time.Second)
		},
	}

	// 创建一个定时任务，该任务将标志变量设置为 true
	tw.scheduleFunc(scheduler, func() {
		flag = true
	})

	// 等待 2 秒，这应该足够让定时任务执行
	time.Sleep(2 * time.Second)

	// 检查标志变量，如果定时任务已经执行，标志变量应该为 true
	if !flag {
		t.Error("Scheduled function not executed")
	}
}

// https://github.com/RussellLuo/timingwheel/blob/master/timingwheel_test.go
// timingwheel_test origin version
func TestTimingWheel_AfterFunc(t *testing.T) {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.start()
	defer tw.stop()

	durations := []time.Duration{
		1 * time.Millisecond,
		5 * time.Millisecond,
		10 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}
	for _, d := range durations {
		t.Run("", func(t *testing.T) {
			exitC := make(chan time.Time)

			start := time.Now().UTC()
			tw.afterFunc(d, func() {
				exitC <- time.Now().UTC()
			})

			got := (<-exitC).Truncate(time.Millisecond)
			minTime := start.Add(d).Truncate(time.Millisecond)

			err := 5 * time.Millisecond
			if got.Before(minTime) || got.After(minTime.Add(err)) {
				t.Errorf("Timer(%s) expiration: want [%s, %s], got %s", d, minTime, minTime.Add(err), got)
			}
		})
	}
}

type scheduler struct {
	intervals []time.Duration
	current   int
}

func (s *scheduler) Next(prev time.Time) time.Time {
	if s.current >= len(s.intervals) {
		return time.Time{}
	}
	next := prev.Add(s.intervals[s.current])
	s.current += 1
	return next
}

func TestTimingWheel_ScheduleFunc(t *testing.T) {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.start()
	defer tw.stop()

	s := &scheduler{intervals: []time.Duration{
		1 * time.Millisecond,   // start + 1ms
		4 * time.Millisecond,   // start + 5ms
		5 * time.Millisecond,   // start + 10ms
		40 * time.Millisecond,  // start + 50ms
		50 * time.Millisecond,  // start + 100ms
		400 * time.Millisecond, // start + 500ms
		500 * time.Millisecond, // start + 1s
	}}

	exitC := make(chan time.Time, len(s.intervals))

	start := time.Now().UTC()
	tw.scheduleFunc(s, func() {
		exitC <- time.Now().UTC()
	})

	accum := time.Duration(0)
	for _, d := range s.intervals {
		got := (<-exitC).Truncate(time.Millisecond)
		accum += d
		minTime := start.Add(accum).Truncate(time.Millisecond)

		err := 5 * time.Millisecond
		if got.Before(minTime) || got.After(minTime.Add(err)) {
			t.Errorf("Timer(%s) expiration: want [%s, %s], got %s", accum, minTime, minTime.Add(err), got)
		}
	}
}
