package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Publisher publishes domain events to a message queue.
type Publisher interface {
	Publish(ctx context.Context, eventType EventType, payload any) error
}

// RedisPublisher publishes events to Redis lists via LPUSH.
type RedisPublisher struct {
	client *redis.Client
}

// NewRedisPublisher returns a new RedisPublisher backed by the given Redis client.
func NewRedisPublisher(client *redis.Client) *RedisPublisher {
	return &RedisPublisher{client: client}
}

// Publish serializes payload and pushes it to the Redis list keyed by eventType.
func (p *RedisPublisher) Publish(ctx context.Context, eventType EventType, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("publisher.Publish marshal: %w", err)
	}
	if err := p.client.LPush(ctx, string(eventType), data).Err(); err != nil {
		return fmt.Errorf("publisher.Publish: %w", err)
	}
	return nil
}
