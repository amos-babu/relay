package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrRoomNotFound       = errors.New("room not found")
	ErrRoomFull           = errors.New("room is full")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
)
