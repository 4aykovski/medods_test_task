package repository

import "errors"

var (
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrSessionAlreadyExists = errors.New("refresh session already exists")
	ErrSessionNotFound      = errors.New("refresh session not found")
)
