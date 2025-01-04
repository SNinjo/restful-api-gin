package user

import "errors"

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidIdFormat = errors.New("invalid ID format")
)
