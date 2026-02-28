package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"desent-api/internal/middlewares"
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
		usecases.NewUpdateBookUsecase(repo),
		usecases.NewDeleteBookUsecase(repo),
	)
	authHandler := NewAuthHandler("test-secret", time.Hour)

	r := chi.NewRouter()
	r.Post("/auth/token", authHandler.CreateToken)
	r.Post("/books", h.CreateBook)
	r.With(middlewares.RequireBearerAuth("test-secret")).Get("/books", h.ListBooks)
	r.Get("/books/{id}", h.GetBookByID)
	r.Put("/books/{id}", h.UpdateBook)
	r.Delete("/books/{id}", h.DeleteBook)
	return r
}

func getTestToken(t *testing.T, r http.Handler) string {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(`{"username":"admin","password":"password"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal token response: %v", err)
	}

	token := payload["token"]
	if token == "" {
		t.Fatalf("expected token in response, got %s", res.Body.String())
	}

	return token
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
	listReq.Header.Set("Authorization", "Bearer "+getTestToken(t, r))
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

func TestBooks_ListUnauthorized(t *testing.T) {
	r := setupBooksRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
	}
}

func TestBooks_ListFilterByAuthor(t *testing.T) {
	r := setupBooksRouter(t)
	token := getTestToken(t, r)

	for _, payload := range []string{
		`{"title":"Dune","author":"Frank Herbert","year":1965}`,
		`{"title":"Foundation","author":"Isaac Asimov","year":1951}`,
		`{"title":"Children of Dune","author":"Frank Herbert","year":1976}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)
		if res.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, res.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/books?author=Frank%20Herbert", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	got := strings.TrimSpace(res.Body.String())
	expected := `[{"id":1,"title":"Dune","author":"Frank Herbert","year":1965},{"id":3,"title":"Children of Dune","author":"Frank Herbert","year":1976}]`
	if got != expected {
		t.Fatalf("unexpected filtered list response: %s", got)
	}
}

func TestBooks_ListPagination(t *testing.T) {
	r := setupBooksRouter(t)
	token := getTestToken(t, r)

	for _, payload := range []string{
		`{"title":"Book 1","author":"A","year":2001}`,
		`{"title":"Book 2","author":"A","year":2002}`,
		`{"title":"Book 3","author":"A","year":2003}`,
		`{"title":"Book 4","author":"A","year":2004}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)
		if res.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, res.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/books?page=1&limit=2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	got := strings.TrimSpace(res.Body.String())
	expected := `[{"id":1,"title":"Book 1","author":"A","year":2001},{"id":2,"title":"Book 2","author":"A","year":2002}]`
	if got != expected {
		t.Fatalf("unexpected paginated list response: %s", got)
	}
}

func TestBooks_ListInvalidPaginationQuery(t *testing.T) {
	r := setupBooksRouter(t)
	token := getTestToken(t, r)

	req := httptest.NewRequest(http.MethodGet, "/books?page=0&limit=2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.Code)
	}

	reqTooLarge := httptest.NewRequest(http.MethodGet, "/books?page=1&limit=101", nil)
	reqTooLarge.Header.Set("Authorization", "Bearer "+token)
	resTooLarge := httptest.NewRecorder()
	r.ServeHTTP(resTooLarge, reqTooLarge)

	if resTooLarge.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resTooLarge.Code)
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

	if got := strings.TrimSpace(notFoundRes.Body.String()); got != `{"error_code":"BOOK_NOT_FOUND","message":"book not found"}` {
		t.Fatalf("unexpected not found response: %s", got)
	}
}

func TestBooks_CreateValidationErrors(t *testing.T) {
	r := setupBooksRouter(t)

	tests := []struct {
		name    string
		body    string
		message string
	}{
		{
			name:    "missing title",
			body:    `{"author":"Author","year":2001}`,
			message: "validation error: title is required",
		},
		{
			name:    "missing author",
			body:    `{"title":"Book","year":2001}`,
			message: "validation error: author is required",
		},
		{
			name:    "missing year",
			body:    `{"title":"Book","author":"Author"}`,
			message: "validation error: year must be between 1450 and 2100",
		},
		{
			name:    "empty object",
			body:    `{}`,
			message: "validation error: title is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			if res.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.Code)
			}

			expected := `{"error_code":"VALIDATION_ERROR","message":"` + tc.message + `"}`
			if got := strings.TrimSpace(res.Body.String()); got != expected {
				t.Fatalf("unexpected validation response: %s", got)
			}
		})
	}
}

func TestBooks_UpdateBook(t *testing.T) {
	r := setupBooksRouter(t)

	createReq := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(`{"title":"Old","author":"Author","year":2001}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRes := httptest.NewRecorder()
	r.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRes.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/books/1", strings.NewReader(`{"title":"New","author":"Writer","year":2002}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRes := httptest.NewRecorder()
	r.ServeHTTP(updateRes, updateReq)

	if updateRes.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateRes.Code)
	}

	if got := strings.TrimSpace(updateRes.Body.String()); got != `{"id":1,"title":"New","author":"Writer","year":2002}` {
		t.Fatalf("unexpected update response: %s", got)
	}
}

func TestBooks_UpdateBookErrors(t *testing.T) {
	r := setupBooksRouter(t)

	invalidIDReq := httptest.NewRequest(http.MethodPut, "/books/abc", strings.NewReader(`{"title":"New","author":"Writer","year":2002}`))
	invalidIDReq.Header.Set("Content-Type", "application/json")
	invalidIDRes := httptest.NewRecorder()
	r.ServeHTTP(invalidIDRes, invalidIDReq)
	if invalidIDRes.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, invalidIDRes.Code)
	}

	notFoundReq := httptest.NewRequest(http.MethodPut, "/books/999", strings.NewReader(`{"title":"New","author":"Writer","year":2002}`))
	notFoundReq.Header.Set("Content-Type", "application/json")
	notFoundRes := httptest.NewRecorder()
	r.ServeHTTP(notFoundRes, notFoundReq)
	if notFoundRes.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, notFoundRes.Code)
	}
}

func TestBooks_DeleteBook(t *testing.T) {
	r := setupBooksRouter(t)

	createReq := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(`{"title":"Delete Me","author":"Author","year":2001}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRes := httptest.NewRecorder()
	r.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRes.Code)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/books/1", nil)
	deleteRes := httptest.NewRecorder()
	r.ServeHTTP(deleteRes, deleteReq)

	if deleteRes.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRes.Code)
	}

	if got := strings.TrimSpace(deleteRes.Body.String()); got != "" {
		t.Fatalf("expected empty body, got %q", got)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/books/1", nil)
	getRes := httptest.NewRecorder()
	r.ServeHTTP(getRes, getReq)
	if getRes.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getRes.Code)
	}
}

func TestBooks_DeleteBookErrors(t *testing.T) {
	r := setupBooksRouter(t)

	invalidIDReq := httptest.NewRequest(http.MethodDelete, "/books/abc", nil)
	invalidIDRes := httptest.NewRecorder()
	r.ServeHTTP(invalidIDRes, invalidIDReq)
	if invalidIDRes.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, invalidIDRes.Code)
	}

	notFoundReq := httptest.NewRequest(http.MethodDelete, "/books/999", nil)
	notFoundRes := httptest.NewRecorder()
	r.ServeHTTP(notFoundRes, notFoundReq)
	if notFoundRes.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, notFoundRes.Code)
	}
}
