package usecases

import "errors"

var ErrValidation = errors.New("validation error")
var ErrInvalidBookID = errors.New("invalid book id")
var ErrBookNotFound = errors.New("book not found")
