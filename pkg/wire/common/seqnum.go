package common

import (
	"math"
	"sync/atomic"
)

type sequence struct {
	num uint32
}

// Next return the next Seq id
func (s *sequence) Next() uint32 {
	next := atomic.AddUint32(&s.num, 1)
	if next == math.MaxUint32 {
		//原子地重置sequence num
		if atomic.CompareAndSwapUint32(&s.num, next, 1) {
			return 1
		}
		return s.Next()
	}
	return next
}

// Seq 饿汉式单例模式，提供给外部一个全局可用的序列号生成器
var Seq = sequence{num: 1}
