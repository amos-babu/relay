package domain

import "errors"

var (
	ErrEmailAlreadyExists         = errors.New("email already exists")
	ErrUserNotFound               = errors.New("user not found")
	ErrInvalidCredentials         = errors.New("invalid credentials")
	ErrRefreshTokenNotFound       = errors.New("refresh token not found")
	ErrUnauthorized               = errors.New("unauthorized")
	ErrInvalidRefreshToken        = errors.New("invalid refresh token")
	ErrConversationNotFound       = errors.New("conversation not found")
	ErrCannotMessageYourself      = errors.New("cannot create conversation with yourself")
	ErrNotConversationParticipant = errors.New("user is not a participant in this conversation")
	ErrEmptyMessage               = errors.New("empty message")
	ErrMessageNotFound            = errors.New("message not found")
	ErrNoParticipant              = errors.New("No participant")
)
