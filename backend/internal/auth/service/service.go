package authservice

import (
	"fmt"
	"time"

	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/logger"
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
	jsonLogger        *logger.Logger
	dummyPasswordHash []byte
	token             *tokens.Tokens
	redis             *redis.Client
}

// New returns a new auth Service with the given repositories and configuration.
func New(
	userRepo authdomain.UserRepository,
	sessionRepo authdomain.SessionRepository,
	tokenRepo authdomain.TokenRepository,
	transactor authdomain.Transactor,
	outbox *events.OutboxWriter,
	token *tokens.Tokens,
	jsonLogger *logger.Logger,
	redisClient *redis.Client,
) (*Service, error) {
	dummyPasswordHash, err := bcrypt.GenerateFromPassword([]byte("dummy"), hashDefaultCost)
	if err != nil {
		return nil, fmt.Errorf("authservice.New: generate dummy hash: %w", err)
	}

	service := &Service{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		tokenRepo:         tokenRepo,
		transactor:        transactor,
		outbox:            outbox,
		token:             token,
		dummyPasswordHash: dummyPasswordHash,
		jsonLogger:        jsonLogger,
		redis:             redisClient,
	}

	return service, nil
}
