package websocket

import "time"

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type TypingRequest struct {
	ConversationID int64 `json:"conversation_id"`
}

type TypingEvent struct {
	ConversationID int64 `json:"conversation_id"`
	UserID         int64 `json:"user_id"`
}

type PresenceEvent struct {
	UserID int64 `json:"user_id"`
	Online bool  `json:"online"`
}

type ReadReceiptRequest struct {
	MessageID      int64 `json:"message_id"`
	ConversationID int64 `json:"conversation_id"`
}

type ReadReceiptEvent struct {
	MessageID      int64     `json:"message_id"`
	ConversationID int64     `json:"conversation_id"`
	UserID         int64     `json:"user_id"`
	ReadAt         time.Time `json:"read_at"`
}

const (
	EventMessage     = "message"      //Deliver message in realtime
	EventTyping      = "typing"       //Show if recipient is typing
	EventReadReceipt = "read_receipt" //Delivered message
	EventPresence    = "presence"     //Online ? Offline
)
