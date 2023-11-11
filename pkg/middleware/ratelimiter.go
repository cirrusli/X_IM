package middleware

import (
	"X_IM/pkg/logger"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"
)

const (
	// TokenBucket 使用令牌桶限流
	TokenBucket = iota
	// SlidingWindow 使用滑动窗口限流
	SlidingWindow
)

type RateLimiter interface {
	Limit(next http.Handler) http.Handler
}

type TokenBucketRateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rl       rate.Limit //令牌数
	bCap     int        //bucket's capacity
}

type SlidingWindowRateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rl       rate.Limit    // 时间戳切片的容量（阈值），将该类型看作整数使用
	interval time.Duration // 时间戳移除的时间间隔
}

type visitor struct {
	limiter    *rate.Limiter
	lastSeen   time.Time
	timestamps []time.Time
}

// NewRateLimiter params[0] 令牌生成速率（个/s），params[1] 令牌桶容量
// NewRateLimiter params[0] 请求数阈值，params[1] 滑动窗口时间间隔
func NewRateLimiter(algorithm byte, params ...int) RateLimiter {
	switch algorithm {
	case TokenBucket:
		rl := rate.Limit(params[0])
		b := params[1]
		return &TokenBucketRateLimiter{
			visitors: make(map[string]*visitor),
			rl:       rl,
			bCap:     b,
		}
	case SlidingWindow:
		rl := rate.Limit(params[0])
		i := time.Duration(params[1]) * time.Second
		return &SlidingWindowRateLimiter{
			visitors: make(map[string]*visitor),
			rl:       rl,
			interval: i,
		}
	default:
		return nil
	}
}

func (rl *TokenBucketRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rl.mu.Lock()
		defer rl.mu.Unlock()
		v, exists := rl.visitors[r.RemoteAddr]
		if !exists {
			v = &visitor{
				limiter: rate.NewLimiter(rl.rl, rl.bCap),
			}
			rl.visitors[r.RemoteAddr] = v
		}
		//rl.mu.Unlock()

		if !v.limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *SlidingWindowRateLimiter) Limit(next http.Handler) http.Handler {
	logger.Info("进入限流检测")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rl.mu.Lock()
		defer rl.mu.Unlock()
		// 查看当前请求的ip是否已有记录
		v, exists := rl.visitors[r.RemoteAddr]
		logger.Infof("Remote IP is: %s", r.RemoteAddr)
		if !exists {
			v = &visitor{
				timestamps: make([]time.Time, 0),
			}
			rl.visitors[r.RemoteAddr] = v
		}
		// Remove timestamps outside the current sliding window
		for len(v.timestamps) > 0 {
			if time.Since(v.timestamps[0]) > rl.interval {
				v.timestamps = v.timestamps[1:]
			} else {
				break
			}
		}
		if len(v.timestamps) >= int(rl.rl) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		v.timestamps = append(v.timestamps, time.Now())
		//rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
