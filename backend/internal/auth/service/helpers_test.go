package authservice

import (
	"context"
	"errors"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRevokeUserSessions(t *testing.T) {
	Convey("revokeUserSessions", t, func() {
		Convey("when fn fails, it wraps and returns the error without touching Redis", func() {
			// mockRedis with setNX left unset — if revokeUserSessions ever calls SetNX
			// here, the mock panics with a clear "not set" message, failing the test.
			srv := newTestService(nil, nil, nil, nil, &mockRedis{})

			fnErr := errors.New("db failure")
			err := srv.revokeUserSessions(context.Background(), "test_caller", "jti-1", time.Now().Add(time.Hour), func(context.Context) error {
				return fnErr
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "test_caller: revoke sessions")
			So(errors.Is(err, fnErr), ShouldBeTrue)
		})

		Convey("when fn succeeds and jti is empty, Redis is skipped entirely", func() {
			srv := newTestService(nil, nil, nil, nil, nil)

			called := false
			err := srv.revokeUserSessions(context.Background(), "test_caller", "", time.Now().Add(time.Hour), func(context.Context) error {
				called = true
				return nil
			})

			So(err, ShouldBeNil)
			So(called, ShouldBeTrue)
		})

		Convey("when fn succeeds and accessTokenExpiresAt is already in the past, Redis is skipped", func() {
			srv := newTestService(nil, nil, nil, nil, nil)

			err := srv.revokeUserSessions(context.Background(), "test_caller", "jti-1", time.Now().Add(-time.Hour), func(context.Context) error {
				return nil
			})

			So(err, ShouldBeNil)
		})

		Convey("when fn succeeds and Redis SetNX succeeds", func() {
			srv := newTestService(nil, nil, nil, nil, newSuccessfulMockRedis())

			err := srv.revokeUserSessions(context.Background(), "test_caller", "jti-1", time.Now().Add(time.Hour), func(context.Context) error {
				return nil
			})

			So(err, ShouldBeNil)
		})

		Convey("when fn succeeds but Redis SetNX fails, the error is wrapped and returned", func() {
			redisErr := testutil.ErrRedisUnavailable
			srv := newTestService(nil, nil, nil, nil, mockRedisSetNXError(redisErr))

			err := srv.revokeUserSessions(context.Background(), "test_caller", "jti-1", time.Now().Add(time.Hour), func(context.Context) error {
				return nil
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "test_caller: session blocklist")
			So(errors.Is(err, redisErr), ShouldBeTrue)
		})
	})
}

func TestEmitTokenEvent(t *testing.T) {
	Convey("emitTokenEvent", t, func() {
		Convey("when token generation succeeds and fn succeeds, the event is emitted via outbox", func() {
			var captured []any
			srv := newTestService(nil, nil, nil, testutil.NewCapturingOutbox(&captured), nil)

			var gotRaw, gotHash string
			var gotExpiry time.Time
			err := srv.emitTokenEvent(context.Background(), "user-1", time.Hour, events.AggregationTypeUser, events.EventUserRegistered,
				func(_ context.Context, rawToken, hashToken string, expiresAt time.Time) (any, error) {
					gotRaw, gotHash, gotExpiry = rawToken, hashToken, expiresAt
					return map[string]string{"token": rawToken}, nil
				})

			So(err, ShouldBeNil)
			So(gotRaw, ShouldNotBeEmpty)
			So(gotHash, ShouldNotBeEmpty)
			So(gotRaw, ShouldNotEqual, gotHash)
			So(gotExpiry.After(time.Now()), ShouldBeTrue)
			So(captured, ShouldNotBeEmpty)
		})

		Convey("when fn returns an error, it is wrapped and outbox.Emit is never called", func() {
			srv := newTestService(nil, nil, nil, testutil.NewFailingOutbox(errors.New("should not be reached")), nil)

			fnErr := errors.New("token persistence failed")
			err := srv.emitTokenEvent(context.Background(), "user-1", time.Hour, events.AggregationTypeUser, events.EventUserRegistered,
				func(context.Context, string, string, time.Time) (any, error) {
					return nil, fnErr
				})

			So(errors.Is(err, fnErr), ShouldBeTrue)
		})

		Convey("when fn succeeds but outbox.Emit fails, the error propagates", func() {
			outboxErr := errors.New("outbox insert failed")
			srv := newTestService(nil, nil, nil, testutil.NewFailingOutbox(outboxErr), nil)

			err := srv.emitTokenEvent(context.Background(), "user-1", time.Hour, events.AggregationTypeUser, events.EventUserRegistered,
				func(context.Context, string, string, time.Time) (any, error) {
					return map[string]string{"a": "b"}, nil
				})

			So(err, ShouldNotBeNil)
			So(errors.Is(err, outboxErr), ShouldBeTrue)
		})
	})
}
