package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type bucket struct {
	mu      sync.Mutex
	count   int
	resetAt time.Time
}

// RateLimiter is an in-memory fixed-window rate limiter keyed by client IP.
type RateLimiter struct {
	buckets sync.Map
	limit   int
	window  time.Duration
}

// NewRateLimiter creates a RateLimiter and starts a background cleanup goroutine.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{limit: limit, window: window}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) allow(ip string) bool {
	now := time.Now()
	v, _ := rl.buckets.LoadOrStore(ip, &bucket{resetAt: now.Add(rl.window)})
	b := v.(*bucket)
	b.mu.Lock()
	defer b.mu.Unlock()
	if now.After(b.resetAt) {
		b.count = 0
		b.resetAt = now.Add(rl.window)
	}
	b.count++
	return b.count <= rl.limit
}

// cleanup removes expired buckets every 10 minutes to prevent unbounded memory growth.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		rl.buckets.Range(func(k, v any) bool {
			b := v.(*bucket)
			b.mu.Lock()
			expired := now.After(b.resetAt)
			b.mu.Unlock()
			if expired {
				rl.buckets.Delete(k)
			}
			return true
		})
	}
}

// RateLimit returns middleware that enforces per-IP request limits.
func RateLimit(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if !rl.allow(ip) {
				retryAfter := int(rl.window.Seconds())
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				http.Error(w, "rate limit exceeded — try again later", http.StatusTooManyRequests)
				slog.Warn("rate limit exceeded", "ip", ip, "path", r.URL.Path)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP extracts the real client IP, honoring X-Forwarded-For / X-Real-IP
// set by a trusted reverse proxy.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For may be a comma-separated list; the first is the origin.
		if host, _, err := net.SplitHostPort(xff); err == nil {
			return host
		}
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
