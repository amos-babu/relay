package models

import "time"

type Message struct {
	ID             int64
	ConversationID int64
	SenderID       int64
	Text           string
	CreatedAt      time.Time
}
