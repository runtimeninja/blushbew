package httpserver

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64 // tokens per second
	capacity float64
}

func NewRateLimiter(reqPerMin int, burst int) *RateLimiter {
	if reqPerMin <= 0 {
		reqPerMin = 60
	}
	if burst <= 0 {
		burst = 30
	}
	return &RateLimiter{
		buckets:  map[string]*bucket{},
		rate:     float64(reqPerMin) / 60.0,
		capacity: float64(burst),
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok {
		rl.buckets[key] = &bucket{tokens: rl.capacity - 1, lastRefill: now}
		return true
	}

	// refill
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = min(rl.capacity, b.tokens+elapsed*rl.rate)
	b.lastRefill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens -= 1
	return true
}

func RateLimit(rl *RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r)
		if !rl.Allow(key) {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate_limited"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	// later we can respect X-Forwarded-For when behind proxy
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
