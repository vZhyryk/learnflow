package authservice

import (
	"context"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"learnflow_backend/internal/shared/tokens"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNew(t *testing.T) {
	Convey("New", t, func() {
		repos := Repos{UserRepo: &mockUserRepo{}, SessionRepo: &mockSessionRepo{}, TokenRepo: &mockTokenRepo{}, Transactor: &testutil.NoopTransactor{}}
		utils := Utils{Token: tokens.NewTokens("test-secret", "", "learnflow", "learnflow-users")}

		Convey("when BcryptCost is zero, it defaults to hashDefaultCost", func() {
			srv, err := New(repos, utils, Options{})
			So(err, ShouldBeNil)
			So(srv.cost, ShouldEqual, hashDefaultCost)
		})

		Convey("when generating the dummy hash fails, it returns an error", func() {
			srv, err := New(repos, utils, Options{BcryptCost: bcrypt.MaxCost + 1})
			So(srv, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "authservice.New: generate dummy hash")
		})
	})
}

// newTestService assembles a Service from mocks
func newTestService(uRepo *mockUserRepo, sRepo *mockSessionRepo, tRepo *mockTokenRepo, outbox *events.OutboxWriter, redisClient *mockRedis) *Service {
	if uRepo == nil {
		uRepo = &mockUserRepo{}
	}
	if sRepo == nil {
		sRepo = &mockSessionRepo{}
	}
	if tRepo == nil {
		tRepo = &mockTokenRepo{}
	}

	srv, err := New(
		Repos{UserRepo: uRepo, SessionRepo: sRepo, TokenRepo: tRepo, Transactor: &testutil.NoopTransactor{}},
		Utils{
			Token:       tokens.NewTokens("test-secret", "", "learnflow", "learnflow-users"),
			Outbox:      outbox,
			RedisClient: redisClient,
		},
		Options{BcryptCost: 4},
	)
	if err != nil {
		panic(err)
	}
	return srv
}

func newSuccessfulMockRedis() *mockRedis {
	return &mockRedis{
		setNX: func(_ context.Context, _ string, _ any, _ time.Duration) *redis.BoolCmd {
			return redis.NewBoolResult(true, nil)
		},
	}
}

func newChangePasswordTestUser() *authdomain.User {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-old-password"), 4)
	if err != nil {
		panic(err)
	}
	return &authdomain.User{ID: "user-123", PasswordHash: string(hash)}
}
