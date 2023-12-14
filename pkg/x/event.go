package x

import (
	"sync"
	"sync/atomic"
)

//思想来源于k8s的events

//Event仅是做简单封闭，方便调用
//Event与ctx, cancel := context.WithCancel(context.Background())的作用相同
//Fire方法等价cancel()
//Done方法等价<-ctx.Done()
//HasFired方法等价select{case <-ctx.Done(): return true; default: return false}

// Event represents a one-time event that may occur in the future.
type Event struct {
	fired int32
	c     chan struct{}
	o     sync.Once
}

// Fire causes e to complete.  It is safe to call multiple times, and concurrently.
// It returns true if this call to Fire caused the signaling
// channel returned by Done to close.
func (e *Event) Fire() bool {
	ret := false
	e.o.Do(func() {
		atomic.StoreInt32(&e.fired, 1)
		close(e.c)
		ret = true
	})
	return ret
}

// Done returns a channel that will be closed when Fire is called.
func (e *Event) Done() <-chan struct{} {
	return e.c
}

// HasFired returns true if Fire has been called.
func (e *Event) HasFired() bool {
	return atomic.LoadInt32(&e.fired) == 1
}

// NewEvent returns a new, ready-to-use Event.
func NewEvent() *Event {
	return &Event{c: make(chan struct{})}
}
