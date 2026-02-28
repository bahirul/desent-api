package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEcho_ValidJSONReturnsExactBody(t *testing.T) {
	input := `{"message":"hello","number":42,"nested":{"key":"val"}}`

	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(input))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	Echo(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	if got := strings.TrimSpace(res.Body.String()); got != input {
		t.Fatalf("expected exact body %q, got %q", input, got)
	}
}

func TestEcho_EmptyObject(t *testing.T) {
	input := `{}`

	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(input))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	Echo(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	if got := strings.TrimSpace(res.Body.String()); got != input {
		t.Fatalf("expected exact body %q, got %q", input, got)
	}
}

func TestEcho_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"broken":`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	Echo(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.Code)
	}
}
