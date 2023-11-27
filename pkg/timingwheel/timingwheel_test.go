package timingwheel

import (
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	tw := NewTimingWheel(1*time.Second, 30)
	tw.start()
	defer tw.stop()

	// 添加一个定时器，如果 TimingWheel 正在运行，这个定时器应该被成功添加
	timer := &Timer{
		expiration: timeToMs(time.Now().UTC().Add(1 * time.Second)),
		task:       func() {},
	}
	if !tw.add(timer) {
		t.Error("Failed to add timer, TimingWheel may not be running")
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

	done := make(chan struct{})

	// 创建一个定时任务，该任务在 1 秒后发送一个信号到 done 通道
	tw.AfterFunc(1*time.Second, func() {
		done <- struct{}{}
	})

	if tw.overflowWheel == nil {
		t.Error("overflowWheel is nil")
	}
	// 等待定时任务执行
	select {
	case <-done:
		// 定时任务已经执行，这是预期的结果
	case <-time.After(2 * time.Second):
		// 如果在 2 秒后定时任务还没有执行，那么测试失败
		t.Error("Scheduled function not executed")
	}

}

type ScheduleFunc struct {
	NextFunc func(time.Time) time.Time
}

func (s *ScheduleFunc) Next(t time.Time) time.Time {
	return s.NextFunc(t)
}

func TestScheduleFunc(t *testing.T) {
	tw := NewTimingWheel(1*time.Second, 30)
	tw.start()
	defer tw.stop()

	flag := false

	// 创建一个调度器，它总是在当前时间的 1 秒后安排下一次执行
	scheduler := &ScheduleFunc{
		NextFunc: func(time.Time) time.Time {
			return time.Now().Add(1 * time.Second)
		},
	}

	// 创建一个定时任务，该任务将标志变量设置为 true
	tw.ScheduleFunc(scheduler, func() {
		flag = true
	})

	// 等待 2 秒，这应该足够让定时任务执行
	time.Sleep(2 * time.Second)

	// 检查标志变量，如果定时任务已经执行，标志变量应该为 true
	if !flag {
		t.Error("Scheduled function not executed")
	}
}
