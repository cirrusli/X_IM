package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenBucketRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(TokenBucket, 1, 1)
	testRateLimiter(t, limiter)
}

func TestSlidingWindowRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(SlidingWindow, 1, 1)
	testRateLimiter(t, limiter)
}

func testRateLimiter(t *testing.T, limiter RateLimiter) {
	handler := limiter.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("GET", "http://localhost", nil)
	rr := httptest.NewRecorder()

	// 第一次请求应该成功
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// 第二次请求应该被限制
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	t.Log("be limited: ", rr.Code)
	if status := rr.Code; status != http.StatusTooManyRequests {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusTooManyRequests)
	}

	// 等待1秒后，应该可以再次请求
	time.Sleep(1 * time.Second)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
