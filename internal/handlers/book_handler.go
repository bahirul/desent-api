package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"desent-api/internal/models"
	"desent-api/internal/usecases"

	"github.com/go-chi/chi/v5"
)

type BookHandler struct {
	createUsecase *usecases.CreateBookUsecase
	listUsecase   *usecases.ListBooksUsecase
	getUsecase    *usecases.GetBookUsecase
}

type errorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

func NewBookHandler(
	createUsecase *usecases.CreateBookUsecase,
	listUsecase *usecases.ListBooksUsecase,
	getUsecase *usecases.GetBookUsecase,
) *BookHandler {
	return &BookHandler{
		createUsecase: createUsecase,
		listUsecase:   listUsecase,
		getUsecase:    getUsecase,
	}
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBookRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON_BODY", "invalid JSON body")
		return
	}

	book, err := h.createUsecase.Execute(r.Context(), req)
	if err != nil {
		status, code, message := mapBookError(err)
		writeError(w, status, code, message)
		return
	}

	writeJSON(w, http.StatusCreated, models.ToBookResponse(book))
}

func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.listUsecase.Execute(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	response := make([]models.BookResponse, 0, len(books))
	for _, book := range books {
		response = append(response, models.ToBookResponse(book))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *BookHandler) GetBookByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	book, err := h.getUsecase.Execute(r.Context(), id)
	if err != nil {
		status, code, message := mapBookError(err)
		writeError(w, status, code, message)
		return
	}

	writeJSON(w, http.StatusOK, models.ToBookResponse(book))
}

func decodeJSON(body io.Reader, target any) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("unexpected trailing content")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{
		ErrorCode: code,
		Message:   message,
	})
}

func mapBookError(err error) (int, string, string) {
	switch {
	case errors.Is(err, usecases.ErrValidation):
		return http.StatusBadRequest, "VALIDATION_ERROR", err.Error()
	case errors.Is(err, usecases.ErrInvalidBookID):
		return http.StatusBadRequest, "INVALID_BOOK_ID", "invalid book id"
	case errors.Is(err, usecases.ErrBookNotFound):
		return http.StatusNotFound, "BOOK_NOT_FOUND", "book not found"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"
	}
}
