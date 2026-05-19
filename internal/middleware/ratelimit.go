package middleware

import (
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	count    int
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	requests int
	window   time.Duration
}

func NewRateLimiter(requests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		requests: requests,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.Sub(v.lastSeen) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		now := time.Now()
		if !exists || now.Sub(v.lastSeen) > rl.window {
			rl.visitors[ip] = &visitor{count: 1, lastSeen: now}
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}
		v.count++
		v.lastSeen = now
		if v.count > rl.requests {
			rl.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", rl.window.String())
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded, try again later"}`))
			return
		}
		rl.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
