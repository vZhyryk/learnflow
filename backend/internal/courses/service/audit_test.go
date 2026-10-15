package courseservice

import (
	"context"
	"errors"
	auditdomain "learnflow_backend/internal/audit/domain"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type adminActionCase struct {
	name string
	run  func(actions *mockAdminActionRepo) error
	want auditdomain.AdminAction
}

func adminActionCases() []adminActionCase {
	ctx := context.Background()
	existing := func(_ context.Context, _ string) (*coursedomain.Course, error) {
		return &coursedomain.Course{ID: "target-1", Slug: "slug"}, nil
	}

	return []adminActionCase{
		{
			name: "Course create",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockCourseRepoRepo{
					getCourseBySlug: func(_ context.Context, _ string) (*coursedomain.Course, error) {
						return nil, coursedomain.ErrCourseNotFound
					},
					createCourse: func(_ context.Context, _ *coursedomain.Course) (*coursedomain.Course, error) {
						return &coursedomain.Course{ID: "target-1"}, nil
					},
				}
				_, err := newTestServiceWithActions(repo, actions).CreateCourse(ctx, coursedomain.CreateCourseRequest{CreatedByUserID: "admin-1"})
				return err
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionCreateItem, TargetType: auditdomain.TargetCourse, TargetID: "target-1"},
		},
		{
			name: "Course update",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockCourseRepoRepo{
					getCourseByID: existing,
					updateCourse:  alwaysSucceedsUpdate,
				}
				return newTestServiceWithActions(repo, actions).UpdateCourse(ctx, coursedomain.UpdateCourseRequest{ID: "target-1"}, "admin-1")
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionUpdateItem, TargetType: auditdomain.TargetCourse, TargetID: "target-1"},
		},
		{
			name: "Course publish",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockCourseRepoRepo{
					getCourseByID: validGetCourseByID,
					publishCourse: func(_ context.Context, _, _ string) error { return nil },
				}
				return newTestServiceWithActions(repo, actions).PublishCourse(ctx, "target-1", "admin-1")
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionPublishItem, TargetType: auditdomain.TargetCourse, TargetID: "target-1"},
		},
		{
			name: "Course archive",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockCourseRepoRepo{archiveCourse: func(_ context.Context, _, _ string) error { return nil }}
				return newTestServiceWithActions(repo, actions).ArchiveCourse(ctx, "target-1", "admin-1")
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionArchiveItem, TargetType: auditdomain.TargetCourse, TargetID: "target-1"},
		},
		{
			name: "Course delete",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockCourseRepoRepo{deleteCourse: func(_ context.Context, _, _ string) error { return nil }}
				return newTestServiceWithActions(repo, actions).DeleteCourse(ctx, "target-1", "admin-1")
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionDeleteItem, TargetType: auditdomain.TargetCourse, TargetID: "target-1"},
		},
	}
}

func TestAdminActionWritten(t *testing.T) {
	Convey("Given each mutating Course operation", t, func() {
		for _, tc := range adminActionCases() {
			Convey(tc.name+" writes exactly one audit entry", func() {
				var got []*auditdomain.AdminAction
				So(tc.run(capturingAdminActions(&got, nil)), ShouldBeNil)
				So(got, ShouldResemble, []*auditdomain.AdminAction{&tc.want})
			})
		}
	})
}

func TestAdminActionFailureIsReturned(t *testing.T) {
	Convey("Given each mutating Course operation", t, func() {
		for _, tc := range adminActionCases() {
			Convey(tc.name+" returns the audit error", func() {
				var got []*auditdomain.AdminAction
				err := tc.run(capturingAdminActions(&got, testutil.ErrDBUnexpected))
				So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			})
		}
	})
}
