package middlewares

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type rateLimitBucket struct {
	windowStart time.Time
	count       int
}

type RateLimiter struct {
	limit   int
	nowFunc func() time.Time

	mu      sync.Mutex
	buckets map[string]rateLimitBucket
}

type rateLimitResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

func NewRateLimiter(limit int, nowFunc func() time.Time) *RateLimiter {
	if limit <= 0 {
		limit = 200
	}
	if nowFunc == nil {
		nowFunc = time.Now
	}

	return &RateLimiter{
		limit:   limit,
		nowFunc: nowFunc,
		buckets: make(map[string]rateLimitBucket),
	}
}

func RateLimitPerIP(limit int) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(limit, time.Now)
	return limiter.Middleware()
}

func (l *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			now := l.nowFunc().UTC()
			window := now.Truncate(time.Minute)

			allowed, remaining, retryAfter := l.allow(ip, window, now)
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(l.limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			if !allowed {
				if retryAfter > 0 {
					w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(rateLimitResponse{
					ErrorCode: "RATE_LIMIT_EXCEEDED",
					Message:   "rate limit exceeded",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (l *RateLimiter) allow(ip string, window, now time.Time) (bool, int, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.cleanup(window)

	bucket, ok := l.buckets[ip]
	if !ok || !bucket.windowStart.Equal(window) {
		bucket = rateLimitBucket{windowStart: window, count: 0}
	}

	if bucket.count >= l.limit {
		retryAfter := int(window.Add(time.Minute).Sub(now).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		return false, 0, retryAfter
	}

	bucket.count++
	l.buckets[ip] = bucket
	remaining := l.limit - bucket.count
	if remaining < 0 {
		remaining = 0
	}

	return true, remaining, 0
}

func (l *RateLimiter) cleanup(window time.Time) {
	cutoff := window.Add(-2 * time.Minute)
	for key, bucket := range l.buckets {
		if bucket.windowStart.Before(cutoff) {
			delete(l.buckets, key)
		}
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}

	if r.RemoteAddr == "" {
		return "unknown"
	}

	return r.RemoteAddr
}
