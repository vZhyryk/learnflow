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

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

var testEmail = "user@example.com"

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

		expected := map[admindomain.UserAdminOperation]auditdomain.AdminActionType{
			"RevokeUserRole": auditdomain.ActionRevokeSubadmin,
			"AssignUserRole": auditdomain.ActionAssignSubadmin,
			"DeleteUser":     auditdomain.ActionDeleteUser,
			"RestoreUser":    auditdomain.ActionRestoreUser,
			"BlockUser":      auditdomain.ActionBlockUser,
			"UnBlockUser":    auditdomain.ActionUnblockUser,
		}

		for operation, actionType := range expected {
			Convey(string(operation)+" runs the repository call and writes the audit entry", func() {
				err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, operation)
				So(err, ShouldBeNil)
				So(called, ShouldEqual, string(operation))
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
			err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, "Nope")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "service."+changeMethod)
			So(*auditWritten, ShouldBeFalse)
		})

		Convey("When the user id is not a UUID", func() {
			err := srv.ChangeUserField(context.Background(), "not-a-uuid", testAdminID, "BlockUser")
			So(errors.Is(err, admindomain.ErrInvalidID), ShouldBeTrue)
			So(*auditWritten, ShouldBeFalse)
		})

		Convey("When the admin targets themselves", func() {
			err := srv.ChangeUserField(context.Background(), testAdminID, testAdminID, "BlockUser")
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
			err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, "BlockUser")
			So(errors.Is(err, admindomain.ErrForbiddenUserAction), ShouldBeTrue)
			So(*auditWritten, ShouldBeFalse)
		})

		Convey("When the target user does not exist", func() {
			users.getUserDataByID = func(_ context.Context, _ string) (*admindomain.UserData, error) {
				return nil, admindomain.ErrUserNotFound
			}
			err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, "BlockUser")
			So(errors.Is(err, admindomain.ErrUserNotFound), ShouldBeTrue)
			So(*auditWritten, ShouldBeFalse)
		})

		Convey("When the repository reports the user not found", func() {
			users.blockUser = func(_ context.Context, _ string) error { return admindomain.ErrUserNotFound }
			err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, "BlockUser")
			So(errors.Is(err, admindomain.ErrUserNotFound), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "service."+changeMethod)
			So(*auditWritten, ShouldBeFalse)
		})

		Convey("When writing the audit entry fails", func() {
			actions.createAdminAction = func(_ context.Context, _ *auditdomain.AdminAction) error {
				return testutil.ErrDBUnexpected
			}
			err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, "BlockUser")
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

		return &admindomain.UserData{UserID: id, Email: &testEmail, Role: role, Status: admindomain.StatusActive}, nil
	}
}

type authCase struct {
	name      string
	actorRole admindomain.UserRole
	target    admindomain.UserRole
	operation admindomain.UserAdminOperation
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
				err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, tc.operation)
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
				So(srv.ChangeUserField(context.Background(), validUserID, testAdminID, tc.operation), ShouldBeNil)
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

		return &admindomain.UserData{UserID: id, Email: &testEmail, Role: admindomain.RoleUser, Status: admindomain.StatusActive}, nil
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
				err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, "BlockUser")
				So(errors.Is(err, admindomain.ErrForbiddenUserAction), ShouldBeTrue)
				So(*auditWritten, ShouldBeFalse)
			})
		}
	})
}

func TestChangeUserFieldActorLoadFailure(t *testing.T) {
	Convey("Given an admin service whose actor lookup fails", t, func() {
		users, _, _, srv := newFailureFixture()
		users.getUserDataByID = func(_ context.Context, id string) (*admindomain.UserData, error) {
			if id == testAdminID {
				return nil, testutil.ErrDBUnexpected
			}

			return userLookup(admindomain.RoleAdmin, admindomain.RoleUser)(context.Background(), id)
		}

		err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, "BlockUser")
		So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, "load actor")
	})
}

func TestChangeUserFieldTargetLoadFailure(t *testing.T) {
	Convey("Given an admin service whose target lookup fails", t, func() {
		users, _, _, srv := newFailureFixture()
		users.getUserDataByID = func(_ context.Context, id string) (*admindomain.UserData, error) {
			if id == validUserID {
				return nil, testutil.ErrDBUnexpected
			}

			return userLookup(admindomain.RoleAdmin, admindomain.RoleUser)(context.Background(), id)
		}

		err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, "BlockUser")
		So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, "load target")
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
	revokes := map[admindomain.UserAdminOperation]bool{
		"BlockUser":      true,
		"DeleteUser":     true,
		"UnBlockUser":    false,
		"RestoreUser":    false,
		"AssignUserRole": false,
		"RevokeUserRole": true,
	}

	Convey("Given an admin service that records session revocations", t, func() {
		for operation, revoked := range revokes {
			Convey(string(operation), func() {
				srv, calls := newSessionFixture(nil)
				So(srv.ChangeUserField(context.Background(), validUserID, testAdminID, operation), ShouldBeNil)
				So(*calls, ShouldResemble, wantSessionCalls(revoked))
			})
		}
	})
}

func TestChangeUserFieldSessionRevokeFailure(t *testing.T) {
	Convey("Given session revocation fails while blocking a user", t, func() {
		srv, _ := newSessionFixture(testutil.ErrDBUnexpected)

		err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, "BlockUser")
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
	want := map[admindomain.UserAdminOperation][]blocklistCall{
		"BlockUser":      {{op: "block", userID: validUserID, ttl: tokens.BlockMarkTTL}},
		"DeleteUser":     {{op: "block", userID: validUserID, ttl: tokens.BlockMarkTTL}},
		"UnBlockUser":    {{op: "unblock", userID: validUserID}},
		"RestoreUser":    {{op: "unblock", userID: validUserID}},
		"AssignUserRole": {{op: "clear_role", userID: validUserID}},
		"RevokeUserRole": {{op: "revoke_role", userID: validUserID, ttl: tokens.BlockMarkTTL}},
	}

	Convey("Given an admin service that records blocklist calls", t, func() {
		for operation, expected := range want {
			Convey(string(operation), func() {
				srv, calls := newBlocklistFixture(nil, nil, nil)
				So(srv.ChangeUserField(context.Background(), validUserID, testAdminID, operation), ShouldBeNil)
				So(*calls, ShouldResemble, expected)
			})
		}
	})
}

func TestChangeUserFieldBlocklistFailures(t *testing.T) {
	Convey("Given Redis fails", t, func() {
		Convey("When blocking the user", func() {
			srv, _ := newBlocklistFixture(testutil.ErrRedisUnavailable, nil, nil)
			err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, "BlockUser")
			So(errors.Is(err, testutil.ErrRedisUnavailable), ShouldBeTrue)
			So(errors.Is(err, admindomain.ErrBlocklistUnavailable), ShouldBeTrue)
		})

		Convey("When unblocking the user", func() {
			srv, _ := newBlocklistFixture(nil, testutil.ErrRedisUnavailable, nil)
			err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, "UnBlockUser")
			So(errors.Is(err, testutil.ErrRedisUnavailable), ShouldBeTrue)
			So(errors.Is(err, admindomain.ErrBlocklistUnavailable), ShouldBeTrue)
		})

		Convey("When marking the role revoked", func() {
			srv, blocklist := newRoleBlocklistFixture()
			blocklist.revokeUserRole = func(_ context.Context, _ string, _ time.Duration) error { return testutil.ErrRedisUnavailable }
			err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, "RevokeUserRole")
			So(errors.Is(err, testutil.ErrRedisUnavailable), ShouldBeTrue)
			So(errors.Is(err, admindomain.ErrBlocklistUnavailable), ShouldBeTrue)
		})

		Convey("When clearing the role-revoked mark", func() {
			srv, blocklist := newRoleBlocklistFixture()
			blocklist.clearUserRoleRevoked = func(_ context.Context, _ string) error { return testutil.ErrRedisUnavailable }
			err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, "AssignUserRole")
			So(errors.Is(err, testutil.ErrRedisUnavailable), ShouldBeTrue)
			So(errors.Is(err, admindomain.ErrBlocklistUnavailable), ShouldBeTrue)
		})
	})
}

func newRoleBlocklistFixture() (*Service, *mockBlocklist) {
	noop := func(_ context.Context, _ string) error { return nil }
	users := &mockUserRepo{
		revokeUserRole:  noop,
		assignUserRole:  noop,
		getUserDataByID: userLookup(admindomain.RoleAdmin, admindomain.RoleUser),
	}
	actions := &mockAdminActionRepo{createAdminAction: func(_ context.Context, _ *auditdomain.AdminAction) error { return nil }}
	blocklist := noopBlocklist()

	return newTestUserServiceFull(users, actions, noopSessions(), blocklist), blocklist
}

func TestChangeUserFieldSkipsBlocklistWhenAuditFails(t *testing.T) {
	Convey("Given the audit write fails while blocking a user", t, func() {
		srv, calls := newBlocklistFixture(nil, nil, testutil.ErrDBUnexpected)

		err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, "BlockUser")
		So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		So(*calls, ShouldBeEmpty)
	})
}

func TestSendEmailEventTypes(t *testing.T) {
	first := "John"
	target := &admindomain.UserData{UserID: validUserID, Email: &testEmail, FirstName: &first}
	cases := map[auditdomain.AdminActionType]string{
		auditdomain.ActionBlockUser:   "user.blocked",
		auditdomain.ActionUnblockUser: "user.unblocked",
		auditdomain.ActionDeleteUser:  "user.deleted",
		auditdomain.ActionRestoreUser: "user.restored",
	}

	Convey("Given an admin service with a capturing outbox", t, func() {
		for action, wantEvent := range cases {
			Convey("When the action is "+string(action), func() {
				var args []any
				srv := newService(&mockAnnouncementRepo{}, &mockUserRepo{}, noopAdminActions(), noopSessions(), &testutil.NoopTransactor{}, testutil.NewCapturingOutbox(&args), noopBlocklist())

				err := srv.sendEmail(context.Background(), action, target)

				So(err, ShouldBeNil)
				So(args, ShouldHaveLength, 4)
				So(args[1], ShouldEqual, validUserID)
				So(args[2], ShouldEqual, wantEvent)
				So(args[3], ShouldContainSubstring, `"email":"user@example.com"`)
				So(args[3], ShouldContainSubstring, `"user_name":"John"`)
				So(args[3], ShouldContainSubstring, `"event_id":"`)
			})
		}

		Convey("When the action has no email notification", func() {
			var args []any
			srv := newService(&mockAnnouncementRepo{}, &mockUserRepo{}, noopAdminActions(), noopSessions(), &testutil.NoopTransactor{}, testutil.NewCapturingOutbox(&args), noopBlocklist())

			err := srv.sendEmail(context.Background(), auditdomain.ActionAssignSubadmin, target)

			So(err, ShouldBeNil)
			So(args, ShouldBeNil)
		})

		Convey("When the outbox fails", func() {
			srv := newService(&mockAnnouncementRepo{}, &mockUserRepo{}, noopAdminActions(), noopSessions(), &testutil.NoopTransactor{}, testutil.NewFailingOutbox(testutil.ErrDBUnexpected), noopBlocklist())

			err := srv.sendEmail(context.Background(), auditdomain.ActionBlockUser, target)

			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})
	})
}

func TestSendEmailSkipsTargetWithoutEmail(t *testing.T) {
	empty := ""
	cases := map[string]*admindomain.UserData{
		"nil target":  nil,
		"nil email":   {UserID: validUserID},
		"empty email": {UserID: validUserID, Email: &empty},
	}

	Convey("Given an admin service with a capturing outbox", t, func() {
		for name, target := range cases {
			Convey("When the target has "+name+", nothing is emitted and no error is returned", func() {
				var args []any
				srv := newService(&mockAnnouncementRepo{}, &mockUserRepo{}, noopAdminActions(), noopSessions(), &testutil.NoopTransactor{}, testutil.NewCapturingOutbox(&args), noopBlocklist())

				err := srv.sendEmail(context.Background(), auditdomain.ActionBlockUser, target)

				So(err, ShouldBeNil)
				So(args, ShouldBeNil)
			})
		}
	})
}

func TestChangeUserFieldTouchesRedisOnlyAfterTheOutboxEmit(t *testing.T) {
	Convey("Given the outbox emit fails", t, func() {
		var calls []blocklistCall
		revoked := false
		users := &mockUserRepo{
			blockUser:       func(_ context.Context, _ string) error { return nil },
			unBlockUser:     func(_ context.Context, _ string) error { return nil },
			getUserDataByID: userLookup(admindomain.RoleAdmin, admindomain.RoleUser),
		}
		sessions := &mockSessionRepo{revokeAllUserSessionsAdmin: func(_ context.Context, _, _ string) error {
			revoked = true
			return nil
		}}
		srv := newService(&mockAnnouncementRepo{}, users, noopAdminActions(), sessions, &testutil.NoopTransactor{},
			testutil.NewFailingOutbox(testutil.ErrDBUnexpected), recordingBlocklist(&calls, nil, nil))

		for _, operation := range []admindomain.UserAdminOperation{admindomain.BlockUser, admindomain.UnBlockUser} {
			Convey(string(operation)+" fails without revoking sessions or touching the Redis mark", func() {
				err := srv.ChangeUserField(context.Background(), validUserID, testAdminID, operation)

				So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
				So(calls, ShouldBeEmpty)
				So(revoked, ShouldBeFalse)
			})
		}
	})

	Convey("Given the target has no email", t, func() {
		var calls []blocklistCall
		lookup := userLookup(admindomain.RoleAdmin, admindomain.RoleUser)
		users := &mockUserRepo{
			blockUser: func(_ context.Context, _ string) error { return nil },
			getUserDataByID: func(ctx context.Context, id string) (*admindomain.UserData, error) {
				data, err := lookup(ctx, id)
				if data != nil {
					data.Email = nil
				}

				return data, err
			},
		}
		srv := newService(&mockAnnouncementRepo{}, users, noopAdminActions(), noopSessions(), &testutil.NoopTransactor{},
			testutil.NewNoopOutbox(), recordingBlocklist(&calls, nil, nil))

		Convey("Blocking still succeeds and marks the user blocked in Redis", func() {
			So(srv.ChangeUserField(context.Background(), validUserID, testAdminID, "BlockUser"), ShouldBeNil)
			So(calls, ShouldResemble, []blocklistCall{{op: "block", userID: validUserID, ttl: tokens.BlockMarkTTL}})
		})
	})
}

var userOperations = []admindomain.UserAdminOperation{
	admindomain.RevokeUserRole, admindomain.AssignUserRole, admindomain.DeleteUser,
	admindomain.RestoreUser, admindomain.BlockUser, admindomain.UnBlockUser,
}

func activeUserWithRole(role admindomain.UserRole) *admindomain.UserData {
	return &admindomain.UserData{Role: role, Status: admindomain.StatusActive}
}

func isRoleChangeOperation(operation admindomain.UserAdminOperation) bool {
	return operation == admindomain.RevokeUserRole || operation == admindomain.AssignUserRole
}

func TestCanChangeUserNobodyMayTouchAnAdmin(t *testing.T) {
	Convey("Given an admin target", t, func() {
		for _, op := range userOperations {
			Convey(string(op)+": no actor role may act on it", func() {
				for _, actorRole := range []admindomain.UserRole{admindomain.RoleAdmin, admindomain.RoleSubAdmin, admindomain.RoleUser} {
					So(canChangeUser(activeUserWithRole(actorRole), activeUserWithRole(admindomain.RoleAdmin), op), ShouldBeFalse)
				}
			})
		}
	})
}

func TestCanChangeUserAdminActor(t *testing.T) {
	Convey("Given an active admin actor", t, func() {
		for _, op := range userOperations {
			Convey(string(op)+": it may act on a subadmin and on a plain user", func() {
				So(canChangeUser(activeUserWithRole(admindomain.RoleAdmin), activeUserWithRole(admindomain.RoleSubAdmin), op), ShouldBeTrue)
				So(canChangeUser(activeUserWithRole(admindomain.RoleAdmin), activeUserWithRole(admindomain.RoleUser), op), ShouldBeTrue)
			})
		}
	})
}

func TestCanChangeUserSubAdminActor(t *testing.T) {
	Convey("Given an active subadmin actor", t, func() {
		for _, op := range userOperations {
			Convey(string(op)+": it may act on a plain user only when it is not a role change", func() {
				So(canChangeUser(activeUserWithRole(admindomain.RoleSubAdmin), activeUserWithRole(admindomain.RoleUser), op), ShouldEqual, !isRoleChangeOperation(op))
			})

			Convey(string(op)+": it may not act on a peer subadmin", func() {
				So(canChangeUser(activeUserWithRole(admindomain.RoleSubAdmin), activeUserWithRole(admindomain.RoleSubAdmin), op), ShouldBeFalse)
			})
		}
	})
}

func TestCanChangeUserPlainUserActor(t *testing.T) {
	Convey("Given an active plain user actor", t, func() {
		for _, op := range userOperations {
			Convey(string(op)+": it may not act on anyone", func() {
				So(canChangeUser(activeUserWithRole(admindomain.RoleUser), activeUserWithRole(admindomain.RoleUser), op), ShouldBeFalse)
				So(canChangeUser(activeUserWithRole(admindomain.RoleUser), activeUserWithRole(admindomain.RoleSubAdmin), op), ShouldBeFalse)
			})
		}
	})
}

func TestCanChangeUserInactiveActor(t *testing.T) {
	deletedAt := time.Now()
	inactive := map[string]*admindomain.UserData{
		"blocked": {Role: admindomain.RoleAdmin, Status: admindomain.StatusBlocked},
		"deleted": {Role: admindomain.RoleAdmin, Status: admindomain.StatusActive, DeletedAt: &deletedAt},
	}

	Convey("Given an actor that is not an active account", t, func() {
		for name, actor := range inactive {
			Convey("A "+name+" admin may not act on a plain user", func() {
				So(canChangeUser(actor, activeUserWithRole(admindomain.RoleUser), "BlockUser"), ShouldBeFalse)
			})
		}
	})
}

const grantItemID = "33333333-3333-3333-3333-333333333333"

type grantFixture struct {
	srv       *Service
	users     *mockUserRepo
	courses   *mockCourseRepo
	content   *mockContentItemRepo
	actions   []*auditdomain.AdminAction
	actionErr error
	outbox    []any
	granted   bool
	user      *admindomain.UserData
}

func newGrantFixture() *grantFixture {
	f := &grantFixture{user: &admindomain.UserData{UserID: validUserID, Email: &testEmail}}
	f.courses = &mockCourseRepo{getCourseTitleByID: func(_ context.Context, _ string) (string, error) { return "Go Basics", nil }}
	f.content = &mockContentItemRepo{getContentItemTitleByID: func(_ context.Context, _ string) (string, error) { return "Intro Video", nil }}
	f.users = &mockUserRepo{
		getUserDataByID: func(_ context.Context, _ string) (*admindomain.UserData, error) { return f.user, nil },
		grantUserCourseAccess: func(_ context.Context, _, _ string) error {
			f.granted = true
			return nil
		},
		grantUserContentAccess: func(_ context.Context, _, _ string) error {
			f.granted = true
			return nil
		},
	}
	f.srv = New(
		Repos{
			AnnounRepo:      &mockAnnouncementRepo{},
			UserRepo:        f.users,
			ActionRepo:      f.capturingActions(),
			SessionRepo:     noopSessions(),
			CourseRepo:      f.courses,
			ContentItemRepo: f.content,
		},
		Utils{Transactor: &testutil.NoopTransactor{}, Outbox: testutil.NewCapturingOutbox(&f.outbox), Blocklist: noopBlocklist()},
	)

	return f
}

func (f *grantFixture) capturingActions() *mockAdminActionRepo {
	return &mockAdminActionRepo{createAdminAction: func(_ context.Context, action *auditdomain.AdminAction) error {
		f.actions = append(f.actions, action)
		return f.actionErr
	}}
}

type grantCase struct {
	name       string
	call       func(f *grantFixture, userID string) error
	detailsKey string
	itemType   string
	itemName   string
	setLookup  func(f *grantFixture, err error)
	setGrant   func(f *grantFixture, err error)
}

func grantCases() []grantCase {
	return []grantCase{
		{
			name: "GrantUserCourseAccess",
			call: func(f *grantFixture, userID string) error {
				return f.srv.GrantUserCourseAccess(context.Background(), userID, grantItemID, testAdminID, auditdomain.ActionGrantItemAccess)
			},
			detailsKey: "course_id",
			itemType:   "course",
			itemName:   "Go Basics",
			setLookup: func(f *grantFixture, err error) {
				f.courses.getCourseTitleByID = func(_ context.Context, _ string) (string, error) { return "", err }
			},
			setGrant: func(f *grantFixture, err error) {
				f.users.grantUserCourseAccess = func(_ context.Context, _, _ string) error { return err }
			},
		},
		{
			name: "GrantUserContentAccess",
			call: func(f *grantFixture, userID string) error {
				return f.srv.GrantUserContentAccess(context.Background(), userID, grantItemID, testAdminID, auditdomain.ActionGrantItemAccess)
			},
			detailsKey: "content_item_id",
			itemType:   "content",
			itemName:   "Intro Video",
			setLookup: func(f *grantFixture, err error) {
				f.content.getContentItemTitleByID = func(_ context.Context, _ string) (string, error) { return "", err }
			},
			setGrant: func(f *grantFixture, err error) {
				f.users.grantUserContentAccess = func(_ context.Context, _, _ string) error { return err }
			},
		},
	}
}

func TestGrantAccessSuccess(t *testing.T) {
	for _, tc := range grantCases() {
		Convey("Given "+tc.name, t, func() {
			f := newGrantFixture()

			err := tc.call(f, validUserID)

			So(err, ShouldBeNil)
			So(f.granted, ShouldBeTrue)

			Convey("it writes an audit entry naming the admin, the user and the item", func() {
				So(f.actions, ShouldHaveLength, 1)
				So(f.actions[0].AdminUserID, ShouldEqual, testAdminID)
				So(f.actions[0].ActionType, ShouldEqual, auditdomain.ActionGrantItemAccess)
				So(f.actions[0].TargetType, ShouldEqual, auditdomain.TargetUser)
				So(f.actions[0].TargetID, ShouldEqual, validUserID)
				So(f.actions[0].Details, ShouldResemble, map[string]any{tc.detailsKey: grantItemID})
			})

			Convey("it queues the notification email with the item title", func() {
				So(f.outbox, ShouldHaveLength, 4)
				So(f.outbox[1], ShouldEqual, validUserID)
				So(f.outbox[2], ShouldEqual, "user.grant.access")
				So(f.outbox[3], ShouldContainSubstring, `"item_name":"`+tc.itemName+`"`)
				So(f.outbox[3], ShouldContainSubstring, `"item_type":"`+tc.itemType+`"`)
				So(f.outbox[3], ShouldContainSubstring, `"email":"user@example.com"`)
			})
		})
	}
}

func TestGrantAccessFailures(t *testing.T) {
	for _, tc := range grantCases() {
		Convey("Given "+tc.name, t, func() {
			f := newGrantFixture()

			Convey("When the user id is not a UUID", func() {
				err := tc.call(f, "not-a-uuid")
				So(errors.Is(err, admindomain.ErrInvalidID), ShouldBeTrue)
				So(f.granted, ShouldBeFalse)
			})

			Convey("When the item does not exist", func() {
				tc.setLookup(f, pgx.ErrNoRows)
				err := tc.call(f, validUserID)
				So(errors.Is(err, admindomain.ErrItemNotFound), ShouldBeTrue)
				So(f.granted, ShouldBeFalse)
			})

			Convey("When the item lookup fails", func() {
				tc.setLookup(f, testutil.ErrDBUnexpected)
				err := tc.call(f, validUserID)
				So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
				So(err.Error(), ShouldContainSubstring, "service."+tc.name)
				So(f.granted, ShouldBeFalse)
			})

			Convey("When the user does not exist", func() {
				f.users.getUserDataByID = func(_ context.Context, _ string) (*admindomain.UserData, error) {
					return nil, admindomain.ErrUserNotFound
				}
				err := tc.call(f, validUserID)
				So(errors.Is(err, admindomain.ErrUserNotFound), ShouldBeTrue)
				So(f.granted, ShouldBeFalse)
			})

			Convey("When the user is soft-deleted", func() {
				deletedAt := time.Now()
				f.user.DeletedAt = &deletedAt
				err := tc.call(f, validUserID)
				So(errors.Is(err, admindomain.ErrInvalidUserState), ShouldBeTrue)
				So(f.granted, ShouldBeFalse)
			})

			Convey("When the user has no email", func() {
				f.user.Email = nil
				err := tc.call(f, validUserID)
				So(errors.Is(err, admindomain.ErrInvalidUserState), ShouldBeTrue)
				So(f.granted, ShouldBeFalse)
			})

			Convey("When the access is already granted", func() {
				tc.setGrant(f, admindomain.ErrAccessAlreadyGranted)
				err := tc.call(f, validUserID)
				So(errors.Is(err, admindomain.ErrAccessAlreadyGranted), ShouldBeTrue)
				So(f.actions, ShouldBeEmpty)
				So(f.outbox, ShouldBeNil)
			})

			Convey("When the grant fails", func() {
				tc.setGrant(f, testutil.ErrDBUnexpected)
				err := tc.call(f, validUserID)
				So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
				So(f.actions, ShouldBeEmpty)
			})

			Convey("When the audit write fails no email is queued", func() {
				f.actionErr = testutil.ErrDBUnexpected
				err := tc.call(f, validUserID)
				So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
				So(err.Error(), ShouldContainSubstring, "audit")
				So(f.outbox, ShouldBeNil)
			})
		})
	}
}

func TestGrantAccessOutboxFailure(t *testing.T) {
	for _, tc := range grantCases() {
		Convey("Given "+tc.name+" with a failing outbox", t, func() {
			f := newGrantFixture()
			f.srv.outbox = testutil.NewFailingOutbox(testutil.ErrDBUnexpected)

			err := tc.call(f, validUserID)

			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})
	}
}

func TestApplyChangeEffectsIgnoresActionsWithoutEffects(t *testing.T) {
	Convey("Given an action that neither blocks, unblocks nor changes the role", t, func() {
		var calls []blocklistCall
		revoked := false
		sessions := &mockSessionRepo{revokeAllUserSessionsAdmin: func(_ context.Context, _, _ string) error {
			revoked = true
			return nil
		}}
		srv := newTestUserServiceFull(&mockUserRepo{}, noopAdminActions(), sessions, recordingBlocklist(&calls, nil, nil))

		err := srv.applyChangeEffects(context.Background(), auditdomain.ActionGrantItemAccess, validUserID, testAdminID)

		So(err, ShouldBeNil)
		So(revoked, ShouldBeFalse)
		So(calls, ShouldBeEmpty)
	})
}

func TestRevokeSubAdminAccessSessionFailure(t *testing.T) {
	Convey("Given the session revocation fails while revoking a subadmin", t, func() {
		var calls []blocklistCall
		sessions := &mockSessionRepo{revokeAllUserSessionsAdmin: func(_ context.Context, _, _ string) error {
			return testutil.ErrDBUnexpected
		}}
		srv := newTestUserServiceFull(&mockUserRepo{}, noopAdminActions(), sessions, recordingBlocklist(&calls, nil, nil))

		err := srv.revokeSubAdminAccess(context.Background(), validUserID, testAdminID)

		So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, "revoke sessions")
		So(calls, ShouldBeEmpty)
	})
}
