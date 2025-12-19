package service

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserInactive      = errors.New("user is inactive")
	ErrUserNotActivated  = errors.New("user is not activated")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrUserAlreadyExists = errors.New("user already exists")
)
