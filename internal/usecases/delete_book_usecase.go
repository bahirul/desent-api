package usecases

import (
	"context"
	"errors"
	"fmt"

	"desent-api/internal/repositories"
)

type DeleteBookUsecase struct {
	repo repositories.BookRepository
}

func NewDeleteBookUsecase(repo repositories.BookRepository) *DeleteBookUsecase {
	return &DeleteBookUsecase{repo: repo}
}

func (u *DeleteBookUsecase) Execute(ctx context.Context, rawID string) error {
	id, err := parseBookID(rawID)
	if err != nil {
		return err
	}

	err = u.repo.DeleteByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrBookNotFound) {
			return ErrBookNotFound
		}

		return fmt.Errorf("delete book: %w", err)
	}

	return nil
}
