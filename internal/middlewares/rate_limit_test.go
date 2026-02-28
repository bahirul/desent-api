package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_AllowThenLimit(t *testing.T) {
	now := time.Date(2026, 2, 28, 12, 0, 1, 0, time.UTC)
	limiter := NewRateLimiter(2, func() time.Time { return now })
	h := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, res.Code)
	}
}

func TestRateLimiter_PerIPIsolation(t *testing.T) {
	now := time.Date(2026, 2, 28, 12, 0, 1, 0, time.UTC)
	limiter := NewRateLimiter(1, func() time.Time { return now })
	h := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	reqA1 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	reqA1.RemoteAddr = "10.0.0.1:1234"
	resA1 := httptest.NewRecorder()
	h.ServeHTTP(resA1, reqA1)
	if resA1.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resA1.Code)
	}

	reqA2 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	reqA2.RemoteAddr = "10.0.0.1:1234"
	resA2 := httptest.NewRecorder()
	h.ServeHTTP(resA2, reqA2)
	if resA2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, resA2.Code)
	}

	reqB := httptest.NewRequest(http.MethodGet, "/ping", nil)
	reqB.RemoteAddr = "10.0.0.2:1234"
	resB := httptest.NewRecorder()
	h.ServeHTTP(resB, reqB)
	if resB.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resB.Code)
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	current := time.Date(2026, 2, 28, 12, 0, 59, 0, time.UTC)
	limiter := NewRateLimiter(1, func() time.Time { return current })
	h := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req1.RemoteAddr = "10.0.0.1:1234"
	res1 := httptest.NewRecorder()
	h.ServeHTTP(res1, req1)
	if res1.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req2.RemoteAddr = "10.0.0.1:1234"
	res2 := httptest.NewRecorder()
	h.ServeHTTP(res2, req2)
	if res2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, res2.Code)
	}

	current = time.Date(2026, 2, 28, 12, 1, 0, 0, time.UTC)
	req3 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req3.RemoteAddr = "10.0.0.1:1234"
	res3 := httptest.NewRecorder()
	h.ServeHTTP(res3, req3)
	if res3.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res3.Code)
	}
}
