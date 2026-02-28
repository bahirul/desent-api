package usecases

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"desent-api/internal/models"
	"desent-api/internal/repositories"
)

type GetBookUsecase struct {
	repo repositories.BookRepository
}

func NewGetBookUsecase(repo repositories.BookRepository) *GetBookUsecase {
	return &GetBookUsecase{repo: repo}
}

func (u *GetBookUsecase) Execute(ctx context.Context, rawID string) (models.Book, error) {
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		return models.Book{}, ErrInvalidBookID
	}

	book, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrBookNotFound) {
			return models.Book{}, ErrBookNotFound
		}

		return models.Book{}, fmt.Errorf("get book: %w", err)
	}

	return book, nil
}
