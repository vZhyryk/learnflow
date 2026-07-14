package events

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPublish(t *testing.T) {
	Convey("Publish", t, func() {
		Convey("returns an error if the payload cannot be marshaled", func() {
			publisher := NewRedisPublisher(nil)
			err := publisher.Publish(context.Background(), EventUserRegistered, make(chan int))
			So(err, ShouldNotBeNil)
		})
	})
}
