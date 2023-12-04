package timingwheel

import (
	"X_IM/pkg/logger"
	"X_IM/pkg/timingwheel/delayqueue"
	"errors"
	"sync/atomic"
	"time"
	"unsafe"
)

type TimingWheeler interface {
	add(t *Timer) bool
	addOrRun(t *Timer)
	advanceClock(expiration int64)
	start()
	stop()
	afterFunc(d time.Duration, f func()) *Timer
	scheduleFunc(s Scheduler, f func()) (t *Timer)
}

// TimingWheel is an implementation of Hierarchical Timing Wheels.
type TimingWheel struct {
	tick      int64 // in milliseconds
	wheelSize int64

	interval    int64     // in milliseconds
	currentTime int64     // in milliseconds
	buckets     []*bucket //存储定时器
	queue       *delayqueue.DelayQueue

	// The higher-level overflow wheel.
	//
	// NOTE: This field may be updated and read concurrently through add.
	overflowWheel unsafe.Pointer // type: *TimingWheel

	exitC     chan struct{}
	waitGroup waitGroupWrapper
}

// 全局的时间轮，对外部屏蔽，只提供Start、Stop、AfterFunc、ScheduleFunc函数
var tw *TimingWheel

// Start configs: tick:1s wheelSize:30
func Start() {
	if tw != nil {
		logger.WithField("module", "timingwheel").Warnln("timingwheel already started")
		return
	}
	tw = NewTimingWheel(1*time.Second, 30)
	tw.start()
}

func Stop() {
	tw.stop()
}

// AfterFunc 等待持续时间过去，然后在自己的 goroutine 中调用 f
// 它返回一个 Timer，可使用其 Stop 方法取消调用
func AfterFunc(d time.Duration, f func()) *Timer {
	return tw.afterFunc(d, f)
}

// ScheduleFunc 根据 s 安排的执行计划(在它自己的 goroutine 中)调用 f
// 它返回一个 Timer，可以使用其 Stop 方法取消调用。
func ScheduleFunc(s Scheduler, f func()) (t *Timer) {
	return tw.scheduleFunc(s, f)
}

// NewTimingWheel creates an instance of TimingWheel with the given tick and wheelSize.
func NewTimingWheel(tick time.Duration, wheelSize int64) *TimingWheel {
	tickMs := int64(tick / time.Millisecond)
	if tickMs <= 0 {
		panic(errors.New("tick must be greater than or equal to 1ms"))
	}

	startMs := timeToMs(time.Now().UTC())

	return newTimingWheel(
		tickMs,
		wheelSize,
		startMs,
		delayqueue.New(int(wheelSize)),
	)
}

func newTimingWheel(tickMs int64, wheelSize int64, startMs int64, queue *delayqueue.DelayQueue) *TimingWheel {
	buckets := make([]*bucket, wheelSize)
	for i := range buckets {
		buckets[i] = newBucket()
	}
	return &TimingWheel{
		tick:        tickMs,
		wheelSize:   wheelSize,
		currentTime: truncate(startMs, tickMs),
		interval:    tickMs * wheelSize,
		buckets:     buckets,
		queue:       queue,
		exitC:       make(chan struct{}),
	}
}

// add inserts the timer t into the current timing wheel.
// NOTE: not depend on start()
func (tw *TimingWheel) add(t *Timer) bool {
	curTime := atomic.LoadInt64(&tw.currentTime)
	if t.expiration < curTime+tw.tick {
		// 该timer处于当前时间格，已经过期
		return false
	} else if t.expiration < curTime+tw.interval {
		// timer未超过当前时间轮的跨度,加入当前时间轮的bucket
		virtualID := t.expiration / tw.tick
		b := tw.buckets[virtualID%tw.wheelSize]
		b.Add(t)
		// 该方法内部使用的CAS，当bucket过期时间改变时重新加入延时队列
		// bucket过期时间改变只有两种情况：1. bucket未加入过延时队列  2. bucket被flush，过期时间被置为-1
		if b.SetExpiration(virtualID * tw.tick) {

			tw.queue.Offer(b, b.Expiration())
		}

		return true
	} else {
		// 超过了当前时间轮的跨度，加入上层时间队列
		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel == nil {
			//TODO 调试的时候如果步入add方法就会赋值成功？？？
			//上层时间队列还不存在，创建并赋值
			atomic.CompareAndSwapPointer(
				&tw.overflowWheel,
				nil,
				unsafe.Pointer(newTimingWheel(
					tw.interval,
					tw.wheelSize,
					curTime,
					tw.queue,
				)),
			)
			overflowWheel = atomic.LoadPointer(&tw.overflowWheel)
		}
		return (*TimingWheel)(overflowWheel).add(t)
	}
}

// addOrRun 将定时器 t 插入到当前的时间轮中，或者如果它已经过期，则运行定时器的任务
func (tw *TimingWheel) addOrRun(t *Timer) {
	if !tw.add(t) {
		// Already expired

		// Like the standard time.AfterFunc (https://golang.org/pkg/time/#AfterFunc),
		// always execute the timer's task in its own goroutine.
		go t.task()
	}
}

func (tw *TimingWheel) advanceClock(expiration int64) {
	curTime := atomic.LoadInt64(&tw.currentTime)
	//  expiration < curTime+tw.tick 表示过期时间正处于当前单元格内，不需要更新时间
	if expiration >= curTime+tw.tick {
		curTime = truncate(expiration, tw.tick)
		atomic.StoreInt64(&tw.currentTime, curTime)

		// Try to advance the clock of the overflow wheel if present
		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel != nil {
			(*TimingWheel)(overflowWheel).advanceClock(curTime)
		}
	}
}

// start the current timing wheel.
func (tw *TimingWheel) start() {
	tw.waitGroup.Wrap(func() {
		tw.queue.Poll(tw.exitC,
			func() int64 {
				return timeToMs(time.Now().UTC())
			})
	})

	tw.waitGroup.Wrap(func() {
		for {
			select {
			//延时队列中有bucket到期
			case elem := <-tw.queue.C:
				b := elem.(*bucket)
				// 推动时间轮的时间
				tw.advanceClock(b.Expiration())
				//刷新处理过期的bucket
				b.Flush(tw.addOrRun)
			case <-tw.exitC:
				//调用 Stop 方法，关闭 exitC
				return
			}
		}
	})
}

// stop stops the current timing wheel.
//
// NOTE: timing wheel must have be added with elem yet, otherwise stop will panic.
//
// If there is any timer's task being running in its own goroutine, stop does
// not wait for the task to complete before returning. If the caller needs to
// know whether the task is completed, it must coordinate with the task explicitly.
func (tw *TimingWheel) stop() {
	close(tw.exitC)
	tw.waitGroup.Wait()
}

// AfterFunc waits for the duration to elapse and then calls f in its own goroutine.
// It returns a Timer that can be used to cancel the call using its Stop method.
func (tw *TimingWheel) afterFunc(d time.Duration, f func()) *Timer {
	t := &Timer{
		expiration: timeToMs(time.Now().UTC().Add(d)),
		task:       f,
	}
	tw.addOrRun(t)
	return t
}

// Scheduler determines the execution plan of a task.
type Scheduler interface {
	// Next returns the next execution time after the given (previous) time.
	// It will return a zero time if no next time is scheduled.
	//
	// All times must be UTC.
	Next(time.Time) time.Time
}

// ScheduleFunc 根据 s 安排的执行计划(在它自己的 goroutine 中)调用 f。它返回一个 Timer ，可以使用其 Stop 方法取消调用。
//
// 如果调用者想要中途终止执行计划，它必须停止定时器，并确保定时器实际上已经停止，
// 因为在当前的实现中，定时器的过期和重新启动之间有一个间隙。确保的等待时间很短，因为间隙非常小。
//
// 在内部，scheduleFunc 最初会询问第一次执行时间(通过调用 s.Next())，
// 并在执行时间非零时创建一个定时器。之后，每次 f 即将被执行时，它都会询问下一个执行时间，如果时间非零，f 将在下一个执行时间被调用。
func (tw *TimingWheel) scheduleFunc(s Scheduler, f func()) (t *Timer) {
	expiration := s.Next(time.Now().UTC())
	if expiration.IsZero() {
		// No time is scheduled
		return
	}

	t = &Timer{
		expiration: timeToMs(expiration),
		task: func() {
			// Schedule the task to execute at the next time if possible.
			expiration := s.Next(msToTime(t.expiration))
			if !expiration.IsZero() {
				t.expiration = timeToMs(expiration)
				tw.addOrRun(t)
			}

			// Actually execute the task.
			f()
		},
	}
	tw.addOrRun(t)

	return
}
