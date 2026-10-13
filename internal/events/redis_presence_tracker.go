package events

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	presenceKeyPrefix = "relay:presence:"
	presenceTTL       = 2 * time.Minute
)

type RedisPresenceTracker struct {
	client *redis.Client
}

func NewRedisPresenceTracker(
	client *redis.Client,
) *RedisPresenceTracker {
	return &RedisPresenceTracker{
		client: client,
	}
}

func presenceKey(userID int64) string {
	return fmt.Sprintf("%s%d", presenceKeyPrefix, userID)
}

// Connect registers a WebSocket connection.
//
// It returns true when this connection makes the user online.
func (t *RedisPresenceTracker) Connect(
	ctx context.Context,
	userID int64,
	connectionID string,
) (bool, error) {
	now := time.Now().UnixMilli()
	cutoff := time.Now().Add(-presenceTTL).UnixMilli()

	script := redis.NewScript(`
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local cutoff = tonumber(ARGV[2])
		local connectionID = ARGV[3]
		local ttlSeconds = tonumber(ARGV[4])

		redis.call(
			'ZREMRANGEBYSCORE',
			key,
			'-inf',
			cutoff
		)

		local count = redis.call('ZCARD', key)

		redis.call(
			'ZADD',
			key,
			now,
			connectionID
		)

		redis.call('EXPIRE', key, ttlSeconds)

		if count == 0 then
			return 1
		end

		return 0
	`)

	result, err := script.Run(
		ctx,
		t.client,
		[]string{presenceKey(userID)},
		now,
		cutoff,
		connectionID,
		int(presenceTTL.Seconds()),
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// Heartbeat refreshes a connection's last-seen timestamp.
func (t *RedisPresenceTracker) Heartbeat(
	ctx context.Context,
	userID int64,
	connectionID string,
) error {
	now := time.Now().UnixMilli()

	script := redis.NewScript(`
		local key = KEYS[1]
		local connectionID = ARGV[1]
		local now = tonumber(ARGV[2])
		local ttlSeconds = tonumber(ARGV[3])

		if redis.call('ZSCORE', key, connectionID) == false then
			return 0
		end

		redis.call('ZADD', key, now, connectionID)
		redis.call('EXPIRE', key, ttlSeconds)

		return 1
	`)

	result, err := script.Run(
		ctx,
		t.client,
		[]string{presenceKey(userID)},
		connectionID,
		now,
		int(presenceTTL.Seconds()),
	).Int()

	if err != nil {
		return err
	}

	if result == 0 {
		return fmt.Errorf(
			"connection %q is not registered for user %d",
			connectionID,
			userID,
		)
	}

	return nil
}

// Disconnect removes a WebSocket connection.
//
// It returns true when this was the user's last active connection.
func (t *RedisPresenceTracker) Disconnect(
	ctx context.Context,
	userID int64,
	connectionID string,
) (bool, error) {
	cutoff := time.Now().Add(-presenceTTL).UnixMilli()

	script := redis.NewScript(`
		local key = KEYS[1]
		local connectionID = ARGV[1]
		local cutoff = tonumber(ARGV[2])

		redis.call(
			'ZREMRANGEBYSCORE',
			key,
			'-inf',
			cutoff
		)

		redis.call('ZREM', key, connectionID)

		local count = redis.call('ZCARD', key)

		if count == 0 then
			redis.call('DEL', key)
			return 1
		end

		return 0
	`)

	result, err := script.Run(
		ctx,
		t.client,
		[]string{presenceKey(userID)},
		connectionID,
		cutoff,
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}
