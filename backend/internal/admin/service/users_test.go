package adminservice

import (
	"context"
	"errors"
	admindomain "learnflow_backend/internal/admin/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"learnflow_backend/internal/shared/tokens"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	validUserID  = "11111111-1111-1111-1111-111111111111"
	testAdminID  = "22222222-2222-2222-2222-222222222222"
	changeMethod = "ChangeUserField"
)

func TestGetUsersData(t *testing.T) {
	Convey("Given an admin service", t, func() {
		users := &mockUserRepo{}
		srv := newTestUserService(users, &mockAdminActionRepo{})

		Convey("When the repository fails", func() {
			users.getUsersData = func(_ context.Context, _ pagination.Params) ([]*admindomain.UserData, int, error) {
				return nil, 0, testutil.ErrDBUnexpected
			}
			got, total, err := srv.GetUsersData(context.Background(), pagination.NewParams(1, 20))
			So(got, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "service.GetUsersData")
		})

		Convey("When it succeeds", func() {
			want := []*admindomain.UserData{{UserID: "user-1"}, {UserID: "user-2"}}
			users.getUsersData = func(_ context.Context, _ pagination.Params) ([]*admindomain.UserData, int, error) {
				return want, 57, nil
			}
			got, total, err := srv.GetUsersData(context.Background(), pagination.NewParams(1, 20))
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 57)
			So(got, ShouldResemble, want)
		})
	})
}

func TestGetUserDataByID(t *testing.T) {
	Convey("Given an admin service", t, func() {
		users := &mockUserRepo{}
		srv := newTestUserService(users, &mockAdminActionRepo{})

		Convey("When the id is not a UUID", func() {
			_, err := srv.GetUserDataByID(context.Background(), "not-a-uuid")
			So(errors.Is(err, admindomain.ErrInvalidID), ShouldBeTrue)
		})

		Convey("When the user is not found", func() {
			users.getUserDataByID = func(_ context.Context, _ string) (*admindomain.UserData, error) {
				return nil, admindomain.ErrUserNotFound
			}
			_, err := srv.GetUserDataByID(context.Background(), validUserID)
			So(errors.Is(err, admindomain.ErrUserNotFound), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "service.GetUserDataByID")
		})

		Convey("When it succeeds", func() {
			var gotID string
			users.getUserDataByID = func(_ context.Context, id string) (*admindomain.UserData, error) {
				gotID = id
				return &admindomain.UserData{UserID: id}, nil
			}
			got, err := srv.GetUserDataByID(context.Background(), validUserID)
			So(err, ShouldBeNil)
			So(got.UserID, ShouldEqual, validUserID)
			So(gotID, ShouldEqual, validUserID)
		})
	})
}

func TestChangeUserFieldOperations(t *testing.T) {
	Convey("Given an admin service with all user operations mocked", t, func() {
		var called string
		var gotUserID string
		var gotAction *auditdomain.AdminAction

		record := func(name string) func(context.Context, string) error {
			return func(_ context.Context, id string) error {
				called, gotUserID = name, id
				return nil
			}
		}
		users := &mockUserRepo{
			revokeUserRole:  record("RevokeUserRole"),
			assignUserRole:  record("AssignUserRole"),
			deleteUser:      record("DeleteUser"),
			restoreUser:     record("RestoreUser"),
			blockUser:       record("BlockUser"),
			unBlockUser:     record("UnBlockUser"),
			getUserDataByID: userLookup(admindomain.RoleAdmin, admindomain.RoleUser),
		}
		actions := &mockAdminActionRepo{
			createAdminAction: func(_ context.Context, action *auditdomain.AdminAction) error {
				gotAction = action
				return nil
			},
		}
		srv := newTestUserService(users, actions)

		expected := map[string]auditdomain.AdminActionType{
			"RevokeUserRole": auditdomain.ActionRevokeSubadmin,
			"AssignUserRole": auditdomain.ActionAssignSubadmin,
			"DeleteUser":     auditdomain.ActionDeleteUser,
			"RestoreUser":    auditdomain.ActionRestoreUser,
			"BlockUser":      auditdomain.ActionBlockUser,
			"UnBlockUser":    auditdomain.ActionUnblockUser,
		}

		for operation, actionType := range expected {
			Convey(operation+" runs the repository call and writes the audit entry", func() {
				err := srv.ChangeUserField(context.Background(), operation, validUserID, testAdminID)
				So(err, ShouldBeNil)
				So(called, ShouldEqual, operation)
				So(gotUserID, ShouldEqual, validUserID)
				So(gotAction, ShouldResemble, &auditdomain.AdminAction{
					AdminUserID: testAdminID,
					ActionType:  actionType,
					TargetType:  auditdomain.TargetUser,
					TargetID:    validUserID,
				})
			})
		}
	})
}

func newFailureFixture() (*mockUserRepo, *mockAdminActionRepo, *bool, *Service) {
	auditWritten := false
	users := &mockUserRepo{
		blockUser:       func(_ context.Context, _ string) error { return nil },
		getUserDataByID: userLookup(admindomain.RoleAdmin, admindomain.RoleUser),
	}
	actions := &mockAdminActionRepo{
		createAdminAction: func(_ context.Context, _ *auditdomain.AdminAction) error {
			auditWritten = true
			return nil
		},
	}

	return users, actions, &auditWritten, newTestUserService(users, actions)
}

func TestChangeUserFieldInputFailures(t *testing.T) {
	Convey("Given an admin service", t, func() {
		_, _, auditWritten, srv := newFailureFixture()

		Convey("When the operation name is unknown", func() {
			err := srv.ChangeUserField(context.Background(), "Nope", validUserID, testAdminID)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "service."+changeMethod)
			So(*auditWritten, ShouldBeFalse)
		})

		Convey("When the user id is not a UUID", func() {
			err := srv.ChangeUserField(context.Background(), "BlockUser", "not-a-uuid", testAdminID)
			So(errors.Is(err, admindomain.ErrInvalidID), ShouldBeTrue)
			So(*auditWritten, ShouldBeFalse)
		})

		Convey("When the admin targets themselves", func() {
			err := srv.ChangeUserField(context.Background(), "BlockUser", testAdminID, testAdminID)
			So(errors.Is(err, admindomain.ErrForbiddenUserAction), ShouldBeTrue)
			So(*auditWritten, ShouldBeFalse)
		})
	})
}

func TestChangeUserFieldTargetFailures(t *testing.T) {
	Convey("Given an admin service", t, func() {
		users, actions, auditWritten, srv := newFailureFixture()

		Convey("When the target is an admin", func() {
			users.getUserDataByID = func(_ context.Context, id string) (*admindomain.UserData, error) {
				return &admindomain.UserData{UserID: id, Role: admindomain.RoleAdmin}, nil
			}
			users.blockUser = func(_ context.Context, _ string) error {
				panic("mutation must not run for an admin target")
			}
			err := srv.ChangeUserField(context.Background(), "BlockUser", validUserID, testAdminID)
			So(errors.Is(err, admindomain.ErrForbiddenUserAction), ShouldBeTrue)
			So(*auditWritten, ShouldBeFalse)
		})

		Convey("When the target user does not exist", func() {
			users.getUserDataByID = func(_ context.Context, _ string) (*admindomain.UserData, error) {
				return nil, admindomain.ErrUserNotFound
			}
			err := srv.ChangeUserField(context.Background(), "BlockUser", validUserID, testAdminID)
			So(errors.Is(err, admindomain.ErrUserNotFound), ShouldBeTrue)
			So(*auditWritten, ShouldBeFalse)
		})

		Convey("When the repository reports the user not found", func() {
			users.blockUser = func(_ context.Context, _ string) error { return admindomain.ErrUserNotFound }
			err := srv.ChangeUserField(context.Background(), "BlockUser", validUserID, testAdminID)
			So(errors.Is(err, admindomain.ErrUserNotFound), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "service."+changeMethod)
			So(*auditWritten, ShouldBeFalse)
		})

		Convey("When writing the audit entry fails", func() {
			actions.createAdminAction = func(_ context.Context, _ *auditdomain.AdminAction) error {
				return testutil.ErrDBUnexpected
			}
			err := srv.ChangeUserField(context.Background(), "BlockUser", validUserID, testAdminID)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "audit")
		})
	})
}

// userLookup returns testAdminID as an active user with actorRole and any other ID as an active user with targetRole.
func userLookup(actorRole, targetRole admindomain.UserRole) func(context.Context, string) (*admindomain.UserData, error) {
	return func(_ context.Context, id string) (*admindomain.UserData, error) {
		role := targetRole
		if id == testAdminID {
			role = actorRole
		}

		return &admindomain.UserData{UserID: id, Role: role, Status: admindomain.StatusActive}, nil
	}
}

type authCase struct {
	name      string
	actorRole admindomain.UserRole
	target    admindomain.UserRole
	operation string
}

func newAuthFixture(actorRole, targetRole admindomain.UserRole) (srv *Service, auditWritten *bool) {
	users, _, auditWritten, srv := newFailureFixture()
	users.getUserDataByID = userLookup(actorRole, targetRole)
	users.blockUser = func(_ context.Context, _ string) error { return nil }
	users.revokeUserRole = func(_ context.Context, _ string) error { return nil }
	users.assignUserRole = func(_ context.Context, _ string) error { return nil }
	users.deleteUser = func(_ context.Context, _ string) error { return nil }

	return srv, auditWritten
}

func TestChangeUserFieldDenied(t *testing.T) {
	cases := []authCase{
		{"subadmin cannot assign a role", admindomain.RoleSubAdmin, admindomain.RoleUser, "AssignUserRole"},
		{"subadmin cannot revoke a role", admindomain.RoleSubAdmin, admindomain.RoleSubAdmin, "RevokeUserRole"},
		{"subadmin cannot block a peer subadmin", admindomain.RoleSubAdmin, admindomain.RoleSubAdmin, "BlockUser"},
		{"subadmin cannot delete a peer subadmin", admindomain.RoleSubAdmin, admindomain.RoleSubAdmin, "DeleteUser"},
		{"plain user cannot act at all", admindomain.RoleUser, admindomain.RoleUser, "BlockUser"},
	}

	Convey("Given an admin service", t, func() {
		for _, tc := range cases {
			Convey(tc.name, func() {
				srv, auditWritten := newAuthFixture(tc.actorRole, tc.target)
				err := srv.ChangeUserField(context.Background(), tc.operation, validUserID, testAdminID)
				So(errors.Is(err, admindomain.ErrForbiddenUserAction), ShouldBeTrue)
				So(*auditWritten, ShouldBeFalse)
			})
		}
	})
}

func TestChangeUserFieldAllowed(t *testing.T) {
	cases := []authCase{
		{"subadmin can block a plain user", admindomain.RoleSubAdmin, admindomain.RoleUser, "BlockUser"},
		{"admin can revoke a subadmin", admindomain.RoleAdmin, admindomain.RoleSubAdmin, "RevokeUserRole"},
		{"admin can block a subadmin", admindomain.RoleAdmin, admindomain.RoleSubAdmin, "BlockUser"},
	}

	Convey("Given an admin service", t, func() {
		for _, tc := range cases {
			Convey(tc.name, func() {
				srv, auditWritten := newAuthFixture(tc.actorRole, tc.target)
				So(srv.ChangeUserField(context.Background(), tc.operation, validUserID, testAdminID), ShouldBeNil)
				So(*auditWritten, ShouldBeTrue)
			})
		}
	})
}

func actorLookup(actor admindomain.UserData) func(context.Context, string) (*admindomain.UserData, error) {
	return func(_ context.Context, id string) (*admindomain.UserData, error) {
		if id == testAdminID {
			actor.UserID = id
			return &actor, nil
		}

		return &admindomain.UserData{UserID: id, Role: admindomain.RoleUser, Status: admindomain.StatusActive}, nil
	}
}

func TestChangeUserFieldActorState(t *testing.T) {
	deletedAt := time.Now()
	cases := map[string]admindomain.UserData{
		"blocked":      {Role: admindomain.RoleAdmin, Status: admindomain.StatusBlocked},
		"soft-deleted": {Role: admindomain.RoleAdmin, Status: admindomain.StatusActive, DeletedAt: &deletedAt},
	}

	Convey("Given an admin service", t, func() {
		for name, actor := range cases {
			Convey("When the actor is "+name, func() {
				users, _, auditWritten, srv := newFailureFixture()
				users.getUserDataByID = actorLookup(actor)
				err := srv.ChangeUserField(context.Background(), "BlockUser", validUserID, testAdminID)
				So(errors.Is(err, admindomain.ErrForbiddenUserAction), ShouldBeTrue)
				So(*auditWritten, ShouldBeFalse)
			})
		}
	})
}

func TestChangeUserFieldActorLoadFailure(t *testing.T) {
	Convey("Given an admin service whose actor lookup fails", t, func() {
		users, _, _, srv := newFailureFixture()
		users.getUserDataByID = func(_ context.Context, _ string) (*admindomain.UserData, error) {
			return nil, testutil.ErrDBUnexpected
		}

		err := srv.ChangeUserField(context.Background(), "BlockUser", validUserID, testAdminID)
		So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, "load actor")
	})
}

type sessionCall struct {
	userID  string
	adminID string
}

func newSessionFixture(sessionErr error) (srv *Service, calls *[]sessionCall) {
	noop := func(_ context.Context, _ string) error { return nil }
	users := &mockUserRepo{
		revokeUserRole:  noop,
		assignUserRole:  noop,
		deleteUser:      noop,
		restoreUser:     noop,
		blockUser:       noop,
		unBlockUser:     noop,
		getUserDataByID: userLookup(admindomain.RoleAdmin, admindomain.RoleUser),
	}
	actions := &mockAdminActionRepo{createAdminAction: func(_ context.Context, _ *auditdomain.AdminAction) error { return nil }}
	var recorded []sessionCall
	sessions := &mockSessionRepo{revokeAllUserSessionsAdmin: func(_ context.Context, userID, adminID string) error {
		recorded = append(recorded, sessionCall{userID: userID, adminID: adminID})
		return sessionErr
	}}

	return newTestUserServiceWithSessions(users, actions, sessions), &recorded
}

func wantSessionCalls(revoked bool) []sessionCall {
	if !revoked {
		return nil
	}

	return []sessionCall{{userID: validUserID, adminID: testAdminID}}
}

func TestChangeUserFieldRevokesSessions(t *testing.T) {
	revokes := map[string]bool{
		"BlockUser":      true,
		"DeleteUser":     true,
		"UnBlockUser":    false,
		"RestoreUser":    false,
		"AssignUserRole": false,
		"RevokeUserRole": false,
	}

	Convey("Given an admin service that records session revocations", t, func() {
		for operation, revoked := range revokes {
			Convey(operation, func() {
				srv, calls := newSessionFixture(nil)
				So(srv.ChangeUserField(context.Background(), operation, validUserID, testAdminID), ShouldBeNil)
				So(*calls, ShouldResemble, wantSessionCalls(revoked))
			})
		}
	})
}

func TestChangeUserFieldSessionRevokeFailure(t *testing.T) {
	Convey("Given session revocation fails while blocking a user", t, func() {
		srv, _ := newSessionFixture(testutil.ErrDBUnexpected)

		err := srv.ChangeUserField(context.Background(), "BlockUser", validUserID, testAdminID)
		So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
	})
}

func newBlocklistFixture(setErr, delErr, auditErr error) (srv *Service, calls *[]blocklistCall) {
	noop := func(_ context.Context, _ string) error { return nil }
	users := &mockUserRepo{
		revokeUserRole:  noop,
		assignUserRole:  noop,
		deleteUser:      noop,
		restoreUser:     noop,
		blockUser:       noop,
		unBlockUser:     noop,
		getUserDataByID: userLookup(admindomain.RoleAdmin, admindomain.RoleUser),
	}
	actions := &mockAdminActionRepo{createAdminAction: func(_ context.Context, _ *auditdomain.AdminAction) error {
		return auditErr
	}}
	var recorded []blocklistCall
	blocklist := recordingBlocklist(&recorded, setErr, delErr)

	return newTestUserServiceFull(users, actions, noopSessions(), blocklist), &recorded
}

func TestChangeUserFieldUpdatesBlocklist(t *testing.T) {
	want := map[string][]blocklistCall{
		"BlockUser":      {{op: "block", userID: validUserID, ttl: tokens.AccessTokenTTL}},
		"DeleteUser":     {{op: "block", userID: validUserID, ttl: tokens.AccessTokenTTL}},
		"UnBlockUser":    {{op: "unblock", userID: validUserID}},
		"RestoreUser":    {{op: "unblock", userID: validUserID}},
		"AssignUserRole": nil,
		"RevokeUserRole": nil,
	}

	Convey("Given an admin service that records blocklist calls", t, func() {
		for operation, expected := range want {
			Convey(operation, func() {
				srv, calls := newBlocklistFixture(nil, nil, nil)
				So(srv.ChangeUserField(context.Background(), operation, validUserID, testAdminID), ShouldBeNil)
				So(*calls, ShouldResemble, expected)
			})
		}
	})
}

func TestChangeUserFieldBlocklistFailures(t *testing.T) {
	Convey("Given Redis fails", t, func() {
		Convey("When blocking the user", func() {
			srv, _ := newBlocklistFixture(testutil.ErrRedisUnavailable, nil, nil)
			err := srv.ChangeUserField(context.Background(), "BlockUser", validUserID, testAdminID)
			So(errors.Is(err, testutil.ErrRedisUnavailable), ShouldBeTrue)
		})

		Convey("When unblocking the user", func() {
			srv, _ := newBlocklistFixture(nil, testutil.ErrRedisUnavailable, nil)
			err := srv.ChangeUserField(context.Background(), "UnBlockUser", validUserID, testAdminID)
			So(errors.Is(err, testutil.ErrRedisUnavailable), ShouldBeTrue)
		})
	})
}

func TestChangeUserFieldSkipsBlocklistWhenAuditFails(t *testing.T) {
	Convey("Given the audit write fails while blocking a user", t, func() {
		srv, calls := newBlocklistFixture(nil, nil, testutil.ErrDBUnexpected)

		err := srv.ChangeUserField(context.Background(), "BlockUser", validUserID, testAdminID)
		So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		So(*calls, ShouldBeEmpty)
	})
}
