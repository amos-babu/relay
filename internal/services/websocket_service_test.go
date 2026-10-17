package services

import (
	"context"
	"encoding/json"
	"errors"
	"relay/internal/events"
	"relay/internal/models"
	"relay/internal/websocket"
	"testing"
)

type fakePresenceTracker struct {
	connectFunc    func(ctx context.Context, userID int64, connectionID string) (bool, error)
	heartbeatFunc  func(ctx context.Context, userID int64, connectionID string) error
	disconnectFunc func(ctx context.Context, userID int64, connectionID string) (bool, error)
}

func (f *fakePresenceTracker) Connect(
	ctx context.Context,
	userID int64,
	connectionID string,
) (bool, error) {
	return f.connectFunc(ctx, userID, connectionID)
}

func (f *fakePresenceTracker) Heartbeat(
	ctx context.Context,
	userID int64,
	connectionID string,
) error {
	return f.heartbeatFunc(ctx, userID, connectionID)
}

func (f *fakePresenceTracker) Disconnect(
	ctx context.Context,
	userID int64,
	connectionID string,
) (bool, error) {
	return f.disconnectFunc(ctx, userID, connectionID)
}

func TestWebSocketService_HandleHeartbeat_RecoversMissingConnection(t *testing.T) {
	heartbeatCalls := 0
	connectCalls := 0

	tracker := &fakePresenceTracker{
		heartbeatFunc: func(
			ctx context.Context,
			userID int64,
			connectionID string,
		) error {
			heartbeatCalls++

			if userID != 1 {
				t.Errorf("expected user ID 1, got %d", userID)
			}

			if connectionID != "connection-123" {
				t.Errorf("unexpected connection ID: %s", connectionID)
			}

			return events.ErrConnectionNotRegistered
		},
		connectFunc: func(
			ctx context.Context,
			userID int64,
			connectionID string,
		) (bool, error) {
			connectCalls++
			return false, nil
		},
	}

	service := &WebSocketService{
		presenceTracker: tracker,
	}

	service.HandleHeartbeat(1, "connection-123")

	if heartbeatCalls != 1 {
		t.Errorf("expected 1 heartbeat call, got %d", heartbeatCalls)
	}

	if connectCalls != 1 {
		t.Errorf("expected 1 recovery registration, got %d", connectCalls)
	}
}

func TestWebSocketService_HandleHeartbeat_RedisErrorDoesNotReconnect(
	t *testing.T,
) {
	heartbeatErr := errors.New("redis unavailable")
	connectCalls := 0

	tracker := &fakePresenceTracker{
		heartbeatFunc: func(
			ctx context.Context,
			userID int64,
			connectionID string,
		) error {
			return heartbeatErr
		},
		connectFunc: func(
			ctx context.Context,
			userID int64,
			connectionID string,
		) (bool, error) {
			connectCalls++
			return true, nil
		},
	}

	service := &WebSocketService{
		presenceTracker: tracker,
	}

	service.HandleHeartbeat(1, "connection-123")

	if connectCalls != 0 {
		t.Errorf(
			"expected no reconnection attempt for a Redis error, got %d",
			connectCalls,
		)
	}
}

func TestWebSocketService_HandleHeartbeat_BroadcastsOnlineAfterRecovery(
	t *testing.T,
) {
	var publishedEvents []websocket.BroadcastEvent

	conversationRepo := &fakeConversationRepository{
		ListForUserFunc: func(
			ctx context.Context,
			userID int64,
		) ([]*models.Conversation, error) {
			return []*models.Conversation{
				{ID: 10},
			}, nil
		},
		ParticipantsFunc: func(
			ctx context.Context,
			conversationID int64,
		) ([]int64, error) {
			return []int64{1, 2}, nil
		},
	}

	publisher := &fakeEventPublisher{
		publishFunc: func(
			ctx context.Context,
			event websocket.BroadcastEvent,
		) error {
			publishedEvents = append(publishedEvents, event)
			return nil
		},
	}

	tracker := &fakePresenceTracker{
		heartbeatFunc: func(
			ctx context.Context,
			userID int64,
			connectionID string,
		) error {
			return events.ErrConnectionNotRegistered
		},
		connectFunc: func(
			ctx context.Context,
			userID int64,
			connectionID string,
		) (bool, error) {
			return true, nil
		},
	}

	service := &WebSocketService{
		conversations:   conversationRepo,
		publisher:       publisher,
		presenceTracker: tracker,
	}

	service.HandleHeartbeat(1, "connection-123")

	if len(publishedEvents) != 1 {
		t.Fatalf(
			"expected 1 published presence event, got %d",
			len(publishedEvents),
		)
	}

	got := publishedEvents[0]

	if len(got.RecipientIDs) != 1 || got.RecipientIDs[0] != 2 {
		t.Errorf("expected recipient [2], got %v", got.RecipientIDs)
	}

	if got.Event.Type != websocket.EventPresence {
		t.Errorf(
			"expected event type %q, got %q",
			websocket.EventPresence,
			got.Event.Type,
		)
	}

	payload, err := json.Marshal(got.Event.Payload)
	if err != nil {
		t.Fatalf("failed to marshal presence payload: %v", err)
	}

	var presence websocket.PresenceEvent
	if err := json.Unmarshal(payload, &presence); err != nil {
		t.Fatalf("failed to unmarshal presence payload: %v", err)
	}

	if presence.UserID != 1 {
		t.Errorf("expected user ID 1, got %d", presence.UserID)
	}

	if !presence.Online {
		t.Error("expected recovered user to be online")
	}
}
