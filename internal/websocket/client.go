package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID  int64
	Conn    *websocket.Conn
	Send    chan []byte
	OnEvent func(userID int64, event Event)
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type TypingRequest struct {
	ConversationID int64 `json:"conversation_id"`
}

type TypingEvent struct {
	ConversationID int64 `json:"conversation_id"`
	UserID         int64 `json:"user_id"`
}

func (c *Client) readPump() {
	c.Conn.SetReadDeadline(
		time.Now().Add(pongWait),
	)
	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(
			time.Now().Add(pongWait),
		)
	})
	defer c.Conn.Close()
	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}
		var event Event

		if err := json.Unmarshal(data, &event); err != nil {
			continue
		}

		if c.OnEvent != nil {
			c.OnEvent(c.UserID, event)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				return
			}
			if err := c.Conn.WriteMessage(
				websocket.TextMessage,
				message,
			); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
