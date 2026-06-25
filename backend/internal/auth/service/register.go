package authservice

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/tokens"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Register creates a new user account and sends an email verification token.
func (s *Service) Register(ctx context.Context, req authdomain.RegisterRequest) (string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, authdomain.ErrUserNotFound) {
		return "", fmt.Errorf("register: get user by email: %w", err)
	}

	if user != nil {
		userEmailUserProfile, getErr := s.userRepo.GetUserProfileByUserID(ctx, user.ID)
		if getErr != nil {
			return "", fmt.Errorf("register: get user profile: %w", err)
		}

		err = s.outbox.Emit(ctx, events.AggregationTypeUser, user.ID, events.EventRegistrationAttemptOnExistingEmail, events.RegistrationAttemptPayload{
			Email:    user.Email,
			UserID:   user.ID,
			UserName: userEmailUserProfile.FirstName,
		})
		if err != nil {
			return "", fmt.Errorf("register: inform user: %w", err)
		}
		return "", authdomain.ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), hashDefaultCost)
	if err != nil {
		return "", fmt.Errorf("register: hash password: %w", err)
	}

	user = &authdomain.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         authdomain.RoleUser,
		Status:       authdomain.StatusPendingVerification,
	}

	var userID string

	err = s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		id, createErr := s.userRepo.CreateUser(ctx, user)
		if createErr != nil {
			return fmt.Errorf("register: create user: %w", err)
		}

		userProfile := &authdomain.UserProfile{
			UserID:      id,
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			PhoneNumber: req.PhoneNumber,
			Country:     req.Country,
			City:        req.City,
			Gender:      req.Gender,
			DateOfBirth: req.DateOfBirth,
			UILanguage:  req.UILanguage,
			AvatarURL:   req.AvatarURL,
			Timezone:    req.Timezone,
			Bio:         req.Bio,
		}

		err = s.userRepo.CreateUserProfile(ctx, userProfile)
		if err != nil {
			return fmt.Errorf("register: create user profile: %w", err)
		}

		rawToken, hashToken, tokenErr := tokens.GenerateSecureToken()
		if tokenErr != nil {
			return fmt.Errorf("register: generate token: %w", err)
		}

		expiresAt := time.Now().UTC().Add(emailVerificationTokenTTL)

		token := &authdomain.EmailVerificationToken{
			TokenBase: authdomain.TokenBase{
				UserID:    id,
				TokenHash: hashToken,
				ExpiresAt: expiresAt,
			},
		}

		if _, err = s.tokenRepo.CreateEmailVerificationToken(ctx, token); err != nil {
			return fmt.Errorf("register: create verification token: %w", err)
		}

		payload := events.UserRegisteredPayload{
			UserID:    id,
			Email:     user.Email,
			ExpiresAt: expiresAt,
			RawToken:  rawToken,
			UserName:  userProfile.FirstName,
		}

		err = s.outbox.Emit(ctx, events.AggregationTypeUser, id, events.EventUserRegistered, payload)
		if err != nil {
			return fmt.Errorf("register: emit event: %w", err)
		}

		userID = id

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("register: transaction: %w", err)
	}

	return userID, nil
}
