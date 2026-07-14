//go:build integration

package worker

import (
	"context"
	"errors"
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	. "github.com/smartystreets/goconvey/convey"
)

// newRealRedisEmailWorker returns an EmailWorker wired to the docker-compose Redis
// (localhost:6379) with a unique idempotency key for this test run, cleaned up after.
func newRealRedisEmailWorker(t *testing.T, keyPrefix string) (*EmailWorker[map[string]string], string) {
	t.Helper()
	w := newTestEmailWorker()
	w.redisClient = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	key := fmt.Sprintf("%s:%d", keyPrefix, time.Now().UnixNano())
	t.Cleanup(func() {
		w.redisClient.Del(context.Background(), key)
		w.redisClient.Close()
	})
	w.cfg.Validate = func(p map[string]string) error { return nil }
	w.cfg.IdempotencyKey = func(p map[string]string) string { return key }
	return w, key
}

func TestHandleMessage_Integration(t *testing.T) {
	Convey("handleMessage", t, func() {
		Convey("Unreachable redis (broken SetNX)", func() {
			w := newTestEmailWorker()
			w.redisClient = testutil.UnreachableRedis()
			w.cfg.Validate = func(p map[string]string) error {
				return nil
			}

			w.cfg.IdempotencyKey = func(p map[string]string) string {
				return "test_key:test_key"
			}

			result, idempotencyKey, err := w.handleMessage(context.Background(), `{"value": "value"}`)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "idempotency check")
			So(result, ShouldBeNil)
			So(idempotencyKey, ShouldBeEmpty)
		})

		Convey("Duplicate message (SetNX ok=false) returns errAlreadyProcessed", func() {
			w, _ := newRealRedisEmailWorker(t, "test_key:duplicate")

			_, _, firstErr := w.handleMessage(context.Background(), `{"value": "value"}`)
			So(firstErr, ShouldBeNil)

			result, idempotencyKey, err := w.handleMessage(context.Background(), `{"value": "value"}`)
			So(errors.Is(err, errAlreadyProcessed), ShouldBeTrue)
			So(result, ShouldBeNil)
			So(idempotencyKey, ShouldBeEmpty)
		})

		Convey("Success returns the payload and idempotency key", func() {
			w, key := newRealRedisEmailWorker(t, "test_key:success")

			result, idempotencyKey, err := w.handleMessage(context.Background(), `{"value": "value"}`)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, &map[string]string{"value": "value"})
			So(idempotencyKey, ShouldEqual, key)
		})

		Convey("Unknown JSON fields on a struct payload are silently ignored, not rejected", func() {
			// handleMessage parses event payloads via plain json.Unmarshal
			// (email_worker.go:94), unlike the HTTP layer's helpers.ReadJSON
			// which sets DisallowUnknownFields. This documents that gap for a
			// real struct-typed event payload (events.RegistrationAttemptPayload
			// — map[string]string, used elsewhere in this file for convenience,
			// has no "known fields" concept and would keep everything, which
			// isn't representative of production payload types): if a producer
			// adds a field an older, still-deployed worker doesn't know about,
			// the worker does NOT error out — it silently drops the unknown
			// field and processes the message using only the fields its struct
			// declares.
			w := &EmailWorker[events.RegistrationAttemptPayload]{
				logger: testutil.NewTestLogger(),
				cfg: Config[events.RegistrationAttemptPayload]{
					EventType: "registration_attempt",
					Validate:  ValidateRegistrationAttemptsPayload,
				},
			}
			w.redisClient = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
			key := fmt.Sprintf("test_key:unknown-field:%d", time.Now().UnixNano())
			t.Cleanup(func() {
				w.redisClient.Del(context.Background(), key) //nolint:errcheck // best-effort cleanup
				w.redisClient.Close()                        //nolint:errcheck // best-effort cleanup
			})
			w.cfg.IdempotencyKey = func(_ events.RegistrationAttemptPayload) string { return key }

			message := `{"user_id": "user-123", "email": "user@example.com", "unexpected_new_field": "from a newer producer"}`
			result, idempotencyKey, err := w.handleMessage(context.Background(), message)

			So(err, ShouldBeNil)
			So(result, ShouldResemble, &events.RegistrationAttemptPayload{UserID: "user-123", Email: "user@example.com"})
			So(idempotencyKey, ShouldEqual, key)
		})
	})
}
