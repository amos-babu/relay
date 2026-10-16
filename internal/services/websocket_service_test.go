package services

import (
	"context"
	"errors"
	"relay/internal/events"
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
