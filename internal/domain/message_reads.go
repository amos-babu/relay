package domain

import "time"

type MessageRead struct {
	MessageID int64
	UserID    int64
	ReadAt    time.Time
}
