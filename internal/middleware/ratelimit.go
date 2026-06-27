package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type rateLimiter struct {
	mu     sync.Mutex
	window time.Duration
	limit  int
	items  map[string][]time.Time
}

func newRateLimiter(window time.Duration, limit int) *rateLimiter {
	return &rateLimiter{
		window: window,
		limit:  limit,
		items:  make(map[string][]time.Time),
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entries := rl.items[key]
	filtered := entries[:0]
	for _, t := range entries {
		if now.Sub(t) < rl.window {
			filtered = append(filtered, t)
		}
	}
	filtered = append(filtered, now)
	rl.items[key] = filtered
	return len(filtered) <= rl.limit
}

func RateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	rl := newRateLimiter(time.Minute, requestsPerMinute)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.allow(clientIP(r)) {
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
