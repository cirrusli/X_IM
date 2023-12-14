package timer

import (
	"X_IM/pkg/timingwheel"
	"testing"
	"time"
)

//result shows that timingwheel is about 50% faster than time package timer

// Benchmark for time package timer
func BenchmarkTimeTimer(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		timer := time.NewTimer(1 * time.Second)
		if !timer.Stop() {
			<-timer.C
		}
	}
}

// Benchmark for timingwheel timer
func BenchmarkTimingWheelTimer(b *testing.B) {
	b.ReportAllocs()
	tw := timingwheel.NewTimingWheel(1*time.Millisecond, 1024)
	tw.Start()
	defer tw.Stop()

	for i := 0; i < b.N; i++ {
		timer := tw.AfterFunc(1*time.Second, func() {})
		timer.Stop()
	}
}

// go test -bench=. -benchmem -run=none
//
//goos: windows
//goarch: amd64
//pkg: X_IM/test/benchmark/timer
//cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
//BenchmarkTimeTimer
//BenchmarkTimeTimer-8             4938370               230.9 ns/op           200
//B/op          3 allocs/op
//BenchmarkTimingWheelTimer
//BenchmarkTimingWheelTimer-8      7517062               154.2 ns/op            80
//B/op          2 allocs/op
