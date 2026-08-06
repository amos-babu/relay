package websocket

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

const (
	EventMessage     = "message"
	EventTyping      = "typing"
	EventReadReceipt = "read_receipt"
	EventPresence    = "presence"
)
