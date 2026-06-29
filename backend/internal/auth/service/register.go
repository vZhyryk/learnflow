package authservice

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/events"
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
		return "", s.handleGetUserByEmailRegisterError(ctx, user)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.cost)
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
			return fmt.Errorf("register: create user: %w", createErr)
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

		userID = id

		return s.emitTokenEvent(ctx, id, emailVerificationTokenTTL, events.AggregationTypeUser, events.EventUserRegistered, func(ctx context.Context, rawToken, hashToken string, expiresAt time.Time) (any, error) {
			token := &authdomain.EmailVerificationToken{
				TokenBase: authdomain.TokenBase{
					UserID:    id,
					TokenHash: hashToken,
					ExpiresAt: expiresAt,
				},
			}

			_, err = s.tokenRepo.CreateEmailVerificationToken(ctx, token)
			if err != nil {
				return nil, fmt.Errorf("register: create verification token: %w", err)
			}

			return events.UserRegisteredPayload{
				UserID:    id,
				Email:     user.Email,
				ExpiresAt: expiresAt,
				RawToken:  rawToken,
				UserName:  userProfile.FirstName,
			}, nil
		})
	})

	if err != nil {
		return "", fmt.Errorf("register: transaction: %w", err)
	}

	return userID, nil
}

func (s *Service) handleGetUserByEmailRegisterError(ctx context.Context, user *authdomain.User) error {
	userEmailUserProfile, getErr := s.userRepo.GetUserProfileByUserID(ctx, user.ID)
	if getErr != nil {
		return fmt.Errorf("register: get user profile: %w", getErr)
	}

	emitErr := s.outbox.Emit(ctx, events.AggregationTypeUser, user.ID, events.EventRegistrationAttemptOnExistingEmail, events.RegistrationAttemptPayload{
		Email:    user.Email,
		UserID:   user.ID,
		UserName: userEmailUserProfile.FirstName,
	})
	if emitErr != nil {
		return fmt.Errorf("register: inform user: %w", emitErr)
	}
	return authdomain.ErrUserAlreadyExists
}
