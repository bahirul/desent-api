package usecases

import (
	"context"

	"desent-api/internal/models"
	"desent-api/internal/repositories"
)

type ListBooksUsecase struct {
	repo repositories.BookRepository
}

func NewListBooksUsecase(repo repositories.BookRepository) *ListBooksUsecase {
	return &ListBooksUsecase{repo: repo}
}

func (u *ListBooksUsecase) Execute(ctx context.Context) ([]models.Book, error) {
	return u.repo.FindAll(ctx)
}
