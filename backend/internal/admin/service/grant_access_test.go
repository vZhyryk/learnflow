package adminservice

import (
	"context"
	"errors"
	admindomain "learnflow_backend/internal/admin/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

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
