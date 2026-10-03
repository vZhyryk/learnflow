//go:build integration

package events

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	. "github.com/smartystreets/goconvey/convey"
)

// uniqueEventType returns a collision-free Redis list key per test run.
func uniqueEventType(prefix string) EventType {
	return EventType(fmt.Sprintf("%s:%d", prefix, time.Now().UnixNano()))
}

func publishAndAssertRaw(t *testing.T, eventTypeName, payloadRawCompare string, payload any) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	eventType := uniqueEventType(eventTypeName)
	t.Cleanup(func() {
		client.Del(context.Background(), string(eventType))
		client.Close()
	})

	publisher := NewRedisPublisher(client)

	err := publisher.Publish(context.Background(), eventType, payload)
	So(err, ShouldBeNil)

	raw, err := client.RPop(context.Background(), string(eventType)).Result()
	So(err, ShouldBeNil)
	So(raw, ShouldEqual, payloadRawCompare)
}

func TestRedisPublisherPublish_Integration(t *testing.T) {
	Convey("RedisPublisher.Publish", t, func() {
		Convey("On success, LPush stores the marshaled payload under the event type key", func() {
			publishAndAssertRaw(t, "publish-success", `{"id":"evt-1","value":"hello"}`, map[string]string{"id": "evt-1", "value": "hello"})
		})

		Convey("On a nil payload, Publish succeeds and pushes the literal JSON null — no required-field check happens before publish", func() {
			publishAndAssertRaw(t, "publish-nil-payload", "null", nil)
		})

		Convey("On a payload struct with a required field left as its zero value, Publish still succeeds — validation is the consumer's job, not the publisher's", func() {
			publishAndAssertRaw(t, "publish-zero-value-payload", `{"user_id":"","email":"","user_name":""}`, RegistrationAttemptPayload{})
		})

		Convey("When the Redis client cannot reach the server, the error is wrapped", func() {
			client := redis.NewClient(&redis.Options{
				Addr:        "127.0.0.1:1",
				DialTimeout: 200 * time.Millisecond,
			})
			t.Cleanup(func() { client.Close() })

			publisher := NewRedisPublisher(client)
			eventType := uniqueEventType("publish-unreachable")

			err := publisher.Publish(context.Background(), eventType, map[string]string{"id": "evt-2"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "publisher.Publish:")
		})
	})
}
