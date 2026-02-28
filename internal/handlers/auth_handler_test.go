package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAuth_CreateToken(t *testing.T) {
	h := NewAuthHandler("test-secret", time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(`{"username":"admin","password":"password"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	h.CreateToken(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	if !strings.Contains(res.Body.String(), `"token":"`) {
		t.Fatalf("expected token response, got %s", res.Body.String())
	}
}

func TestAuth_CreateTokenInvalidCredentials(t *testing.T) {
	h := NewAuthHandler("test-secret", time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(`{"username":"admin","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	h.CreateToken(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
	}
}

func TestAuth_CreateTokenInvalidJSON(t *testing.T) {
	h := NewAuthHandler("test-secret", time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(`{"username":`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	h.CreateToken(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.Code)
	}
}
