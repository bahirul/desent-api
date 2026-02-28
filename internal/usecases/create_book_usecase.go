package usecases

import (
	"context"
	"fmt"

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
	book, err := validateCreateBookRequest(req)
	if err != nil {
		return models.Book{}, err
	}

	created, err := u.repo.Create(ctx, book)
	if err != nil {
		return models.Book{}, fmt.Errorf("create book: %w", err)
	}

	return created, nil
}
