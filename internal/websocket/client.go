package websocket

import "github.com/gorilla/websocket"

type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
}
