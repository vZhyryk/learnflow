package authservice

import (
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/logger"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenTTL            = 15 * time.Minute
	refreshTokenTTL           = 7 * 24 * time.Hour
	emailVerificationTokenTTL = 24 * time.Hour
	passwordResetTokenTTL     = 1 * time.Hour
)

// Service implements the auth domain Service interface.
type Service struct {
	userRepo          authdomain.UserRepository
	sessionRepo       authdomain.SessionRepository
	tokenRepo         authdomain.TokenRepository
	dummyPasswordHash []byte
	transactor        authdomain.Transactor
	JWTSecret         string
	jsonLogger        *logger.Logger
}

// New returns a new auth Service with the given repositories and configuration.
func New(
	userRepo authdomain.UserRepository,
	sessionRepo authdomain.SessionRepository,
	tokenRepo authdomain.TokenRepository,
	transactor authdomain.Transactor,
	jwtSecret string,
	jsonLogger *logger.Logger,
) *Service {
	var dummyPasswordHash, err = bcrypt.GenerateFromPassword([]byte("dummy"), bcrypt.DefaultCost)
	if err != nil {
		jsonLogger.Fatal(err, map[string]any{"message": "dummyPasswordHash was not generated"})
	}

	return &Service{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		tokenRepo:         tokenRepo,
		transactor:        transactor,
		JWTSecret:         jwtSecret,
		dummyPasswordHash: dummyPasswordHash,
		jsonLogger:        jsonLogger,
	}
}
