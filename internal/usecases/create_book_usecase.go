package usecases

import (
	"context"
	"fmt"
	"strings"

	"desent-api/internal/models"
	"desent-api/internal/repositories"
)

type CreateBookUsecase struct {
	repo repositories.BookRepository
}

func NewCreateBookUsecase(repo repositories.BookRepository) *CreateBookUsecase {
	return &CreateBookUsecase{repo: repo}
}

func (u *CreateBookUsecase) Execute(ctx context.Context, req models.CreateBookRequest) (models.Book, error) {
	title := strings.TrimSpace(req.Title)
	author := strings.TrimSpace(req.Author)

	if title == "" {
		return models.Book{}, fmt.Errorf("%w: title is required", ErrValidation)
	}

	if author == "" {
		return models.Book{}, fmt.Errorf("%w: author is required", ErrValidation)
	}

	if req.Year < 1450 || req.Year > 2100 {
		return models.Book{}, fmt.Errorf("%w: year must be between 1450 and 2100", ErrValidation)
	}

	book := models.Book{
		Title:  title,
		Author: author,
		Year:   req.Year,
	}

	return u.repo.Create(ctx, book)
}
