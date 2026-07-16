package authservice

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/ptr"
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
	}

	var userID string

	err = s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		id, createErr := s.userRepo.CreateUser(ctx, user)
		if createErr != nil {
			return fmt.Errorf("register: create user: %w", createErr)
		}

		userProfile := newUserProfileFromRegisterRequest(id, req)

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
				UserName:  ptr.StringOrEmpty(userProfile.FirstName),
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
		UserName: ptr.StringOrEmpty(userEmailUserProfile.FirstName),
	})
	if emitErr != nil {
		return fmt.Errorf("register: inform user: %w", emitErr)
	}
	return authdomain.ErrUserAlreadyExists
}

// defaultUILanguage matches user_profiles.ui_language's DB DEFAULT — applied here
// (business rule), not in the repository (persistence concern).
const defaultUILanguage = "uk"

// newUserProfileFromRegisterRequest maps RegisterRequest ("" = not provided) onto
// UserProfile (*string, nil = not provided) for nullable user_profiles columns.
func newUserProfileFromRegisterRequest(userID string, req authdomain.RegisterRequest) *authdomain.UserProfile {
	uiLanguage := req.UILanguage
	if uiLanguage == "" {
		uiLanguage = defaultUILanguage
	}

	return &authdomain.UserProfile{
		UserID:      userID,
		FirstName:   ptr.StringOrNil(req.FirstName),
		LastName:    ptr.StringOrNil(req.LastName),
		PhoneNumber: ptr.StringOrNil(req.PhoneNumber),
		Country:     ptr.StringOrNil(req.Country),
		City:        ptr.StringOrNil(req.City),
		Gender:      ptr.StringOrNil(req.Gender),
		DateOfBirth: req.DateOfBirth,
		UILanguage:  uiLanguage,
		AvatarURL:   ptr.StringOrNil(req.AvatarURL),
		Timezone:    ptr.StringOrNil(req.Timezone),
		Bio:         ptr.StringOrNil(req.Bio),
	}
}
