package models

import "time"

type ConversationParticipant struct {
	ConversationID int64
	UserID         int64
	joinedAt       time.Time
}
