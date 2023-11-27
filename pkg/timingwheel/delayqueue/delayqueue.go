package delayqueue

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"time"
)

// DelayQueue 是一个无大小限制的阻塞队列，用priorityQueue作为内部的存储结构
// 元素只有在其延迟过期后才能被取出。队列的头部是延迟最早过期的元素。
type DelayQueue struct {
	C chan any // 用于发送过期元素的通道

	mu sync.Mutex
	// 队列中的元素按照其延迟时间排序，延迟时间最短的元素位于队列头部
	pq priorityQueue

	sleeping int32         // 是否被挂起
	wakeupC  chan struct{} // 用于唤醒阻塞挂起的 Poll 方法的通道
}

// New 创建一个指定大小的 DelayQueue 实例
func New(size int) *DelayQueue {
	return &DelayQueue{
		C:       make(chan any),
		pq:      newPriorityQueue(size),
		wakeupC: make(chan struct{}),
	}
}

// Offer 将元素插入到当前队列中
func (dq *DelayQueue) Offer(elem any, expiration int64) {
	item := &item{Value: elem, Priority: expiration}

	dq.mu.Lock()
	heap.Push(&dq.pq, item)
	index := item.Index
	dq.mu.Unlock()

	if index == 0 {
		// 新加入的元素排在首部，判断延时队列是否处于挂起状态
		if atomic.CompareAndSwapInt32(&dq.sleeping, 1, 0) {
			// 唤醒延时队列
			dq.wakeupC <- struct{}{}
		}
	}
}

// Poll 持续等待一个元素过期，然后将过期的元素发送到通道 C
func (dq *DelayQueue) Poll(exit chan struct{}, now func() int64) {
	for {
		now := now()

		dq.mu.Lock()
		item, delta := dq.pq.PeekAndShift(now)
		if item == nil {
			// 队列中没有元素，或者没有到期的元素，将队列置为休眠状态
			atomic.StoreInt32(&dq.sleeping, 1)
		}
		dq.mu.Unlock()

		if item == nil {
			if delta == 0 {
				// no item in dq
				select {
				case <-dq.wakeupC:
					continue
				case <-exit:
					dq.resetState()
				}
			} else if delta > 0 {
				// At least one item is pending
				select {
				case <-dq.wakeupC:
					// 一个比当前堆顶元素优先级更高的元素加入时，唤醒dq
					continue
				case <-time.After(time.Duration(delta) * time.Millisecond):
					// 堆顶元素过期
					// 这里与 Offer 操作并没有互斥，如果定时器到期和插入优先级更早的元素两个事件同时发生，所以需要如下操作
					if atomic.SwapInt32(&dq.sleeping, 0) == 0 {
						<-dq.wakeupC
					}
					continue
				case <-exit:
					dq.resetState()
					return
				}
			}
		}

		select {
		case dq.C <- item.Value:
			// 过期的元素已经成功发送
		case <-exit:
			dq.resetState()
		}
	}

}

func (dq *DelayQueue) resetState() {
	atomic.StoreInt32(&dq.sleeping, 0)
}
