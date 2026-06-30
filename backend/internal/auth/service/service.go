package authservice

import (
	"context"
	"fmt"
	"time"

	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/tokens"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenTTL            = 15 * time.Minute
	refreshTokenTTL           = 7 * 24 * time.Hour
	emailVerificationTokenTTL = 24 * time.Hour
	passwordResetTokenTTL     = 1 * time.Hour
	emailChangeTokenTTL       = 1 * time.Hour
	accountRecoverTokenTTL    = 24 * time.Hour

	hashDefaultCost = 12
)

// Service implements the auth domain Service interface.
type Service struct {
	userRepo          authdomain.UserRepository
	sessionRepo       authdomain.SessionRepository
	tokenRepo         authdomain.TokenRepository
	transactor        authdomain.Transactor
	outbox            *events.OutboxWriter
	dummyPasswordHash []byte
	cost              int
	token             *tokens.Tokens
	redisClient       RedisOps
}

// Repos groups the repository dependencies required by the auth Service.
type Repos struct {
	UserRepo    authdomain.UserRepository
	SessionRepo authdomain.SessionRepository
	TokenRepo   authdomain.TokenRepository
	Transactor  authdomain.Transactor
}

// Utils groups the infrastructure utilities required by the auth Service.
type Utils struct {
	Outbox      *events.OutboxWriter
	Token       *tokens.Tokens
	RedisClient RedisOps
}

// New returns a new auth Service with the given repositories and configuration.
func New(
	repos Repos,
	utils Utils,
	opts Options,
) (*Service, error) {
	cost := opts.BcryptCost
	if cost == 0 {
		cost = hashDefaultCost
	}
	dummyPasswordHash, err := bcrypt.GenerateFromPassword([]byte("dummy"), cost)
	if err != nil {
		return nil, fmt.Errorf("authservice.New: generate dummy hash: %w", err)
	}

	service := &Service{
		userRepo:          repos.UserRepo,
		sessionRepo:       repos.SessionRepo,
		tokenRepo:         repos.TokenRepo,
		transactor:        repos.Transactor,
		outbox:            utils.Outbox,
		token:             utils.Token,
		redisClient:       utils.RedisClient,
		cost:              cost,
		dummyPasswordHash: dummyPasswordHash,
	}

	return service, nil
}

// Options configures optional parameters for the auth Service.
type Options struct {
	BcryptCost int // default hashDefaultCost (12), bcrypt.MinCost (4) in tests
}

// RedisOps is the subset of redis.Client methods used by the auth Service.
type RedisOps interface {
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd
}
