package websocket

import (
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

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
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			return
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
