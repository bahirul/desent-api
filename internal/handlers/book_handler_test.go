package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"desent-api/internal/repositories"
	"desent-api/internal/usecases"

	"github.com/go-chi/chi/v5"
	_ "modernc.org/sqlite"
)

func setupBooksRouter(t *testing.T) http.Handler {
	t.Helper()

	dbPath := t.TempDir() + "/books.db"
	db, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := repositories.InitBooksSchema(context.Background(), db); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	repo := repositories.NewSQLiteBookRepository(db)
	h := NewBookHandler(
		usecases.NewCreateBookUsecase(repo),
		usecases.NewListBooksUsecase(repo),
		usecases.NewGetBookUsecase(repo),
	)

	r := chi.NewRouter()
	r.Post("/books", h.CreateBook)
	r.Get("/books", h.ListBooks)
	r.Get("/books/{id}", h.GetBookByID)
	return r
}

func TestBooks_CreateListGet(t *testing.T) {
	r := setupBooksRouter(t)

	createReq := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(`{"title":"Clean Code","author":"Robert C. Martin","year":2008}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRes := httptest.NewRecorder()
	r.ServeHTTP(createRes, createReq)

	if createRes.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRes.Code)
	}

	if got := strings.TrimSpace(createRes.Body.String()); got != `{"id":1,"title":"Clean Code","author":"Robert C. Martin","year":2008}` {
		t.Fatalf("unexpected create response: %s", got)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/books", nil)
	listRes := httptest.NewRecorder()
	r.ServeHTTP(listRes, listReq)

	if listRes.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRes.Code)
	}

	if got := strings.TrimSpace(listRes.Body.String()); got != `[{"id":1,"title":"Clean Code","author":"Robert C. Martin","year":2008}]` {
		t.Fatalf("unexpected list response: %s", got)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/books/1", nil)
	getRes := httptest.NewRecorder()
	r.ServeHTTP(getRes, getReq)

	if getRes.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRes.Code)
	}

	if got := strings.TrimSpace(getRes.Body.String()); got != `{"id":1,"title":"Clean Code","author":"Robert C. Martin","year":2008}` {
		t.Fatalf("unexpected get response: %s", got)
	}
}

func TestBooks_GetByIDErrors(t *testing.T) {
	r := setupBooksRouter(t)

	invalidIDReq := httptest.NewRequest(http.MethodGet, "/books/abc", nil)
	invalidIDRes := httptest.NewRecorder()
	r.ServeHTTP(invalidIDRes, invalidIDReq)
	if invalidIDRes.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, invalidIDRes.Code)
	}

	notFoundReq := httptest.NewRequest(http.MethodGet, "/books/999", nil)
	notFoundRes := httptest.NewRecorder()
	r.ServeHTTP(notFoundRes, notFoundReq)
	if notFoundRes.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, notFoundRes.Code)
	}
}
