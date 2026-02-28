package usecases

import (
	"context"
	"errors"
	"fmt"

	"desent-api/internal/models"
	"desent-api/internal/repositories"
)

type UpdateBookUsecase struct {
	repo repositories.BookRepository
}

func NewUpdateBookUsecase(repo repositories.BookRepository) *UpdateBookUsecase {
	return &UpdateBookUsecase{repo: repo}
}

func (u *UpdateBookUsecase) Execute(ctx context.Context, rawID string, req models.CreateBookRequest) (models.Book, error) {
	id, err := parseBookID(rawID)
	if err != nil {
		return models.Book{}, err
	}

	book, err := validateCreateBookRequest(req)
	if err != nil {
		return models.Book{}, err
	}

	updated, err := u.repo.UpdateByID(ctx, id, book)
	if err != nil {
		if errors.Is(err, repositories.ErrBookNotFound) {
			return models.Book{}, ErrBookNotFound
		}

		return models.Book{}, fmt.Errorf("update book: %w", err)
	}

	return updated, nil
}
