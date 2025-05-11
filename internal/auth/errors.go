package auth

import "errors"

var (
	ErrInvalidCreds      = errors.New("Invalid credentials")
	ErrUserAlreadyExists = errors.New("User already exists")
	ErrInvalidToken      = errors.New("invalid token")
)
