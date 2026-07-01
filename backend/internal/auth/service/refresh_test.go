package authservice

import (
	"context"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/tokens"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func newRefreshTestService(sRepo *mockSessionRepo, uRepo *mockUserRepo) *Service {
	srv, err := New(
		Repos{UserRepo: uRepo, SessionRepo: sRepo, TokenRepo: &mockTokenRepo{}, Transactor: &mockTransactor{}},
		Utils{Token: tokens.NewTokens("test-secret", "", "learnflow", "learnflow-users")},
		Options{BcryptCost: 4},
	)
	if err != nil {
		panic(err)
	}
	return srv
}

func newRefreshTestFixtures() (authdomain.RefreshRequest, *authdomain.User, *authdomain.UserSession) {
	now := time.Now().UTC().Truncate(time.Second)
	validReq := authdomain.RefreshRequest{RefreshToken: "raw-refresh-token", UserAgent: "test-agent", IPAddress: "127.0.0.1"}
	activeUser := &authdomain.User{ID: "user-123", Role: authdomain.RoleUser, Status: authdomain.StatusActive}
	activeSession := &authdomain.UserSession{ID: "session-123", UserID: "user-123", ExpiresAt: now.Add(7 * 24 * time.Hour)}
	return validReq, activeUser, activeSession
}

func TestRefreshSuccess(t *testing.T) {
	validReq, activeUser, activeSession := newRefreshTestFixtures()

	Convey("Given an auth service", t, func() {
		Convey("When the refresh token rotates successfully", func() {
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return activeSession, nil
				},
				getSessionByPrevHash: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return nil, authdomain.ErrSessionNotFound
				},
				updateSessionToken: func(_ context.Context, sessionID, _, ua, ip string) error {
					So(sessionID, ShouldEqual, activeSession.ID)
					So(ua, ShouldEqual, validReq.UserAgent)
					So(ip, ShouldEqual, validReq.IPAddress)
					return nil
				},
			}
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, userID string) (*authdomain.User, error) {
					So(userID, ShouldEqual, activeSession.UserID)
					return activeUser, nil
				},
			}
			srv := newRefreshTestService(sRepo, uRepo)

			got, err := srv.Refresh(context.Background(), validReq)

			So(err, ShouldBeNil)
			So(got.AccessToken, ShouldNotBeEmpty)
			So(got.RefreshToken, ShouldNotBeEmpty)
			So(got.RefreshToken, ShouldNotEqual, validReq.RefreshToken)
			So(got.UserID, ShouldEqual, activeSession.UserID)
			So(got.ExpiresAt, ShouldEqual, activeSession.ExpiresAt)
		})
	})
}

func TestRefreshNoReuseWhenSessionNotFound(t *testing.T) {
	validReq, _, _ := newRefreshTestFixtures()

	Convey("Given an auth service", t, func() {
		Convey("When the session is not found and the token was not reused", func() {
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return nil, authdomain.ErrSessionNotFound
				},
				getSessionByPrevHash: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return nil, authdomain.ErrSessionNotFound
				},
			}
			srv := newRefreshTestService(sRepo, &mockUserRepo{})

			_, err := srv.Refresh(context.Background(), validReq)

			So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
		})
	})
}

func TestRefreshReuseBeforeUserFetch(t *testing.T) {
	validReq, _, activeSession := newRefreshTestFixtures()

	Convey("Given an auth service", t, func() {
		Convey("When a rotated-out token is replayed (reuse before user fetch)", func() {
			var gotUserID string
			var gotRevokedBy *string
			var gotReason authdomain.RevokeReason
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return nil, authdomain.ErrSessionNotFound
				},
				getSessionByPrevHash: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return activeSession, nil
				},
				revokeAllUserSessions: func(_ context.Context, userID string, revokedBy *string, reason authdomain.RevokeReason) error {
					gotUserID, gotRevokedBy, gotReason = userID, revokedBy, reason
					return nil
				},
			}
			srv := newRefreshTestService(sRepo, &mockUserRepo{})

			_, err := srv.Refresh(context.Background(), validReq)

			So(errors.Is(err, authdomain.ErrSessionRevoked), ShouldBeTrue)
			So(gotUserID, ShouldEqual, activeSession.UserID)
			So(gotRevokedBy, ShouldBeNil)
			So(gotReason, ShouldEqual, authdomain.RevokeReasonSuspiciousActivity)
		})
	})
}

func TestRefreshReuseAfterUserFetch(t *testing.T) {
	validReq, activeUser, activeSession := newRefreshTestFixtures()

	Convey("Given an auth service", t, func() {
		Convey("When the same token rotates two sessions (reuse after user fetch)", func() {
			revoked := false
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return activeSession, nil
				},
				getSessionByPrevHash: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return activeSession, nil
				},
				revokeAllUserSessions: func(_ context.Context, _ string, _ *string, _ authdomain.RevokeReason) error {
					revoked = true
					return nil
				},
			}
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return activeUser, nil
				},
			}
			srv := newRefreshTestService(sRepo, uRepo)

			_, err := srv.Refresh(context.Background(), validReq)

			So(errors.Is(err, authdomain.ErrSessionRevoked), ShouldBeTrue)
			So(revoked, ShouldBeTrue)
		})
	})
}

func TestRefreshTokenReuseErrors(t *testing.T) {
	validReq, _, activeSession := newRefreshTestFixtures()

	Convey("Given an auth service", t, func() {
		Convey("When revoking sessions during reuse detection fails", func() {
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return nil, authdomain.ErrSessionNotFound
				},
				getSessionByPrevHash: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return activeSession, nil
				},
				revokeAllUserSessions: func(_ context.Context, _ string, _ *string, _ authdomain.RevokeReason) error {
					return errors.New("db connection lost")
				},
			}
			srv := newRefreshTestService(sRepo, &mockUserRepo{})

			_, err := srv.Refresh(context.Background(), validReq)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "revoke all sessions (reuse)")
		})

		Convey("When checking for a reused token fails unexpectedly", func() {
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return nil, authdomain.ErrSessionNotFound
				},
				getSessionByPrevHash: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return nil, errors.New("db connection lost")
				},
			}
			srv := newRefreshTestService(sRepo, &mockUserRepo{})

			_, err := srv.Refresh(context.Background(), validReq)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get prev session (reuse check)")
		})
	})
}

func TestRefreshUserStatus(t *testing.T) {
	validReq, _, activeSession := newRefreshTestFixtures()

	Convey("Given an auth service", t, func() {
		Convey("When the user lookup fails unexpectedly", func() {
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return activeSession, nil
				},
			}
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, errors.New("db connection lost")
				},
			}
			srv := newRefreshTestService(sRepo, uRepo)

			_, err := srv.Refresh(context.Background(), validReq)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get user")
		})

		Convey("When the account is blocked", func() {
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return activeSession, nil
				},
			}
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return &authdomain.User{ID: "user-123", Status: authdomain.StatusBlocked}, nil
				},
			}
			srv := newRefreshTestService(sRepo, uRepo)

			_, err := srv.Refresh(context.Background(), validReq)

			So(errors.Is(err, authdomain.ErrAccountBlocked), ShouldBeTrue)
		})

		Convey("When the account is deleted", func() {
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return activeSession, nil
				},
			}
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return &authdomain.User{ID: "user-123", Status: authdomain.StatusDeleted}, nil
				},
			}
			srv := newRefreshTestService(sRepo, uRepo)

			_, err := srv.Refresh(context.Background(), validReq)

			So(errors.Is(err, authdomain.ErrInvalidCredentials), ShouldBeTrue)
		})
	})
}

func TestRefreshSessionPersistence(t *testing.T) {
	validReq, activeUser, activeSession := newRefreshTestFixtures()

	Convey("Given an auth service", t, func() {
		Convey("When persisting the rotated session token fails", func() {
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return activeSession, nil
				},
				getSessionByPrevHash: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return nil, authdomain.ErrSessionNotFound
				},
				updateSessionToken: func(_ context.Context, _, _, _, _ string) error {
					return errors.New("db connection lost")
				},
			}
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return activeUser, nil
				},
			}
			srv := newRefreshTestService(sRepo, uRepo)

			_, err := srv.Refresh(context.Background(), validReq)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "update session token")
		})
	})
}
