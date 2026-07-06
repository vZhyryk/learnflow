package events

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPublish(t *testing.T) {
	Convey("Publish", t, func() {
		Convey("returns an error if the payload cannot be marshaled (permanent — not retried)", func() {
			publisher := NewRedisPublisher(nil)
			err := publisher.Publish(context.Background(), EventUserRegistered, make(chan int))
			So(err, ShouldNotBeNil)
		})

		// The LPush success/failure branches are not covered here: RedisPublisher.client
		// is a concrete *redis.Client, not an interface, so it can't be swapped for a
		// function-field mock the way mockRedis is elsewhere. Covering LPush requires a
		// real Redis-compatible backend (miniredis) — left for integration tests.
	})
}
