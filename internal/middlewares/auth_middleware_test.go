package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"desent-api/internal/utils"
)

func TestRequireBearerAuth(t *testing.T) {
	secret := "test-secret"
	validToken, err := utils.GenerateToken("admin", secret, time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mw := RequireBearerAuth(secret)

	t.Run("missing authorization", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books", nil)
		res := httptest.NewRecorder()
		mw(next).ServeHTTP(res, req)
		if res.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books", nil)
		req.Header.Set("Authorization", "Bearer badtoken")
		res := httptest.NewRecorder()
		mw(next).ServeHTTP(res, req)
		if res.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
		}
	})

	t.Run("valid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		res := httptest.NewRecorder()
		mw(next).ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
		}
	})
}
