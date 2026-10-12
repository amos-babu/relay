package events

import (
	"context"
	"encoding/json"
	"log"
	"relay/internal/websocket"

	"github.com/redis/go-redis/v9"
)

type RedisSubscriber struct {
	client *redis.Client
	hub    *websocket.Hub
}

func NewRedisSubscriber(client *redis.Client, hub *websocket.Hub) *RedisSubscriber {
	return &RedisSubscriber{
		client: client,
		hub:    hub,
	}
}

func (s *RedisSubscriber) Subscribe(ctx context.Context) error {
	//Subscribe subscribes the client
	pubsub := s.client.Subscribe(ctx, EventChannel)
	defer pubsub.Close()

	// Wait until Redis confirms the subscription.
	if _, err := pubsub.Receive(ctx); err != nil {
		return err
	}

	log.Printf("Subscribed to Redis channel: %s", EventChannel)

	//Channel returns a Go channel for concurrently receiving messages
	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done(): // Listens for cancellation returning an error and exiting cleanly
			return ctx.Err()

		case msg, ok := <-ch:
			if !ok {
				return nil //Exits gracefully if the channel is closed
			}

			//Parsing the broadcast event
			var event websocket.BroadcastEvent

			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("Failed to decode Redis event: %v", err)
				continue
			}

			data, err := json.Marshal(event.Event)
			if err != nil {
				log.Printf("Failed to encode Redis event: %v", err)
				continue
			}

			for _, userID := range event.RecipientIDs {
				s.hub.SendToUser(userID, data)
			}
		}
	}
}
