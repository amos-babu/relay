package websocket

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

const (
	EventMessage     = "message"
	EventTyping      = "typing"
	EventReadReceipt = "read_receipt"
	EventPresence    = "presence"
)
