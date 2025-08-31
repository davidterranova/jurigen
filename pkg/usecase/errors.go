package usecase

import "errors"

var (
	ErrInvalidCommand = errors.New("invalid command")
	ErrNotFound       = errors.New("not found")
	ErrInternal       = errors.New("internal server error")
)
