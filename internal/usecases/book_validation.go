package usecases

import (
	"fmt"
	"strconv"
	"strings"

	"desent-api/internal/models"
)

func parseBookID(rawID string) (int64, error) {
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidBookID
	}

	return id, nil
}

func validateCreateBookRequest(req models.CreateBookRequest) (models.Book, error) {
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

	return models.Book{
		Title:  title,
		Author: author,
		Year:   req.Year,
	}, nil
}
