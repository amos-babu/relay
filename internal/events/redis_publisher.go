package events

import (
	"context"
	"encoding/json"

	"relay/internal/websocket"

	"github.com/redis/go-redis/v9"
)

const EventChannel = "relay:events"

type RedisPublisher struct {
	client *redis.Client
}

func NewRedisPublisher(client *redis.Client) *RedisPublisher {
	return &RedisPublisher{
		client: client,
	}
}

func (p *RedisPublisher) Publish(
	ctx context.Context,
	event websocket.BroadcastEvent,
) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, EventChannel, payload).Err()
}
