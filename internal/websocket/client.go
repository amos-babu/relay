package websocket

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
}

func (c *Client) readPump() {
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}
	}
}

func (c *Client) writePump() {
	for message := range c.Send {
		log.Printf("Sending to user %d: %s", c.UserID, message)
		if err := c.Conn.WriteMessage(
			websocket.TextMessage,
			message,
		); err != nil {
			return
		}
	}
}
