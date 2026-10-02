package services

import (
	"context"
	"relay/internal/websocket"
)

type EventPublisher interface {
	Publish(ctx context.Context, event websocket.BroadcastEvent) error
}
