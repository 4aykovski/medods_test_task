package repository

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrSessionAlreadyExists = errors.New("refresh session already exists")
	ErrSessionNotFound      = errors.New("refresh session not found")
	ErrUserSessionsNotFound = errors.New("user doesn't have refresh sessions")
)
