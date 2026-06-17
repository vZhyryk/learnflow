package authservice

import (
	"time"

	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenTTL            = 15 * time.Minute
	refreshTokenTTL           = 7 * 24 * time.Hour
	emailVerificationTokenTTL = 24 * time.Hour
	passwordResetTokenTTL     = 1 * time.Hour

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
	publisher         events.Publisher
	dummyPasswordHash []byte
	jwtSecret         string
}

// New returns a new auth Service with the given repositories and configuration.
func New(
	userRepo authdomain.UserRepository,
	sessionRepo authdomain.SessionRepository,
	tokenRepo authdomain.TokenRepository,
	transactor authdomain.Transactor,
	outbox *events.OutboxWriter,
	jwtSecret string,
	jsonLogger *logger.Logger,
	redisClient *redis.Client,
) *Service {
	dummyPasswordHash, err := bcrypt.GenerateFromPassword([]byte("dummy"), hashDefaultCost)
	if err != nil {
		jsonLogger.Fatal(err, map[string]any{"message": "dummyPasswordHash was not generated"})
	}

	return &Service{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		tokenRepo:         tokenRepo,
		transactor:        transactor,
		outbox:            outbox,
		jwtSecret:         jwtSecret,
		dummyPasswordHash: dummyPasswordHash,
		jsonLogger:        jsonLogger,
		publisher:         events.NewRedisPublisher(redisClient),
	}
}
