package reviewservice

import (
	"context"
	"errors"
	auditdomain "learnflow_backend/internal/audit/domain"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type adminActionCase struct {
	name string
	run  func(actions *mockAdminActionRepo) error
	want auditdomain.AdminAction
}

func reviewAction(actionType auditdomain.AdminActionType) auditdomain.AdminAction {
	return auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: actionType, TargetType: auditdomain.TargetReview, TargetID: "target-1"}
}

func courseActionCases() []adminActionCase {
	ctx := context.Background()
	nothing := func(_ context.Context, _, _ string) error { return nil }

	repo := &mockReviewRepo{
		getCourseReviewByUserAndCourseID: func(_ context.Context, _, _ string) (*reviewdomain.CourseReview, error) {
			return nil, reviewdomain.ErrReviewNotFound
		},
		createCourseReview: func(_ context.Context, _ *reviewdomain.CourseReview) (*reviewdomain.CourseReview, error) {
			return &reviewdomain.CourseReview{ID: "target-1"}, nil
		},
		getCourseReviewByID: func(_ context.Context, _ string) (*reviewdomain.CourseReview, error) {
			return &reviewdomain.CourseReview{ID: "target-1"}, nil
		},
		updateCourseReview: func(_ context.Context, _ *reviewdomain.CourseReview) error { return nil },
		deleteCourseReview: nothing,
	}
	return []adminActionCase{
		{
			name: "course review create (admin)",
			run: func(actions *mockAdminActionRepo) error {
				return courseSvc(repo, actions).CreateCourseReviewAdmin(ctx, reviewdomain.CreateCourseReviewRequest{UserID: "admin-1"})
			},
			want: reviewAction(auditdomain.ActionCreateItem),
		},
		{
			name: "course review update (admin)",
			run: func(actions *mockAdminActionRepo) error {
				return courseSvc(repo, actions).UpdateCourseReviewAdmin(ctx, reviewdomain.UpdateCourseReviewRequest{ReviewID: "target-1"}, "admin-1")
			},
			want: reviewAction(auditdomain.ActionUpdateItem),
		},
		{
			name: "course review delete (admin)",
			run: func(actions *mockAdminActionRepo) error {
				return courseSvc(repo, actions).DeleteCourseReviewAdmin(ctx, "target-1", "admin-1")
			},
			want: reviewAction(auditdomain.ActionDeleteItem),
		},
	}
}

func contentActionCases() []adminActionCase {
	ctx := context.Background()
	nothing := func(_ context.Context, _, _ string) error { return nil }

	repo := &mockReviewRepo{
		getContentReviewByUserAndContentID: func(_ context.Context, _, _ string) (*reviewdomain.ContentReview, error) {
			return nil, reviewdomain.ErrReviewNotFound
		},
		createContentReview: func(_ context.Context, _ *reviewdomain.ContentReview) (*reviewdomain.ContentReview, error) {
			return &reviewdomain.ContentReview{ID: "target-1"}, nil
		},
		getContentReviewByID: func(_ context.Context, _ string) (*reviewdomain.ContentReview, error) {
			return &reviewdomain.ContentReview{ID: "target-1"}, nil
		},
		updateContentReview: func(_ context.Context, _ *reviewdomain.ContentReview) error { return nil },
		deleteContentReview: nothing,
	}
	return []adminActionCase{
		{
			name: "content review create (admin)",
			run: func(actions *mockAdminActionRepo) error {
				return contentSvc(repo, actions).CreateContentReviewAdmin(ctx, reviewdomain.CreateContentReviewRequest{UserID: "admin-1"})
			},
			want: reviewAction(auditdomain.ActionCreateItem),
		},
		{
			name: "content review update (admin)",
			run: func(actions *mockAdminActionRepo) error {
				return contentSvc(repo, actions).UpdateContentReviewAdmin(ctx, reviewdomain.UpdateContentReviewRequest{ReviewID: "target-1"}, "admin-1")
			},
			want: reviewAction(auditdomain.ActionUpdateItem),
		},
		{
			name: "content review delete (admin)",
			run: func(actions *mockAdminActionRepo) error {
				return contentSvc(repo, actions).DeleteContentReviewAdmin(ctx, "target-1", "admin-1")
			},
			want: reviewAction(auditdomain.ActionDeleteItem),
		},
	}
}

func articleActionCases() []adminActionCase {
	ctx := context.Background()
	nothing := func(_ context.Context, _, _ string) error { return nil }

	repo := &mockReviewRepo{
		getArticleReviewByUserAndArticleID: func(_ context.Context, _, _ string) (*reviewdomain.ArticleReview, error) {
			return nil, reviewdomain.ErrReviewNotFound
		},
		createArticleReview: func(_ context.Context, _ *reviewdomain.ArticleReview) (*reviewdomain.ArticleReview, error) {
			return &reviewdomain.ArticleReview{ID: "target-1"}, nil
		},
		getArticleReviewByID: func(_ context.Context, _ string) (*reviewdomain.ArticleReview, error) {
			return &reviewdomain.ArticleReview{ID: "target-1"}, nil
		},
		updateArticleReview: func(_ context.Context, _ *reviewdomain.ArticleReview) error { return nil },
		deleteArticleReview: nothing,
	}
	return []adminActionCase{
		{
			name: "article review create (admin)",
			run: func(actions *mockAdminActionRepo) error {
				return articleSvc(repo, actions).CreateArticleReviewAdmin(ctx, reviewdomain.CreateArticleReviewRequest{UserID: "admin-1"})
			},
			want: reviewAction(auditdomain.ActionCreateItem),
		},
		{
			name: "article review update (admin)",
			run: func(actions *mockAdminActionRepo) error {
				return articleSvc(repo, actions).UpdateArticleReviewAdmin(ctx, reviewdomain.UpdateArticleReviewRequest{ReviewID: "target-1"}, "admin-1")
			},
			want: reviewAction(auditdomain.ActionUpdateItem),
		},
		{
			name: "article review delete (admin)",
			run: func(actions *mockAdminActionRepo) error {
				return articleSvc(repo, actions).DeleteArticleReviewAdmin(ctx, "target-1", "admin-1")
			},
			want: reviewAction(auditdomain.ActionDeleteItem),
		},
	}
}

func adminActionCases() []adminActionCase {
	cases := courseActionCases()
	cases = append(cases, contentActionCases()...)

	return append(cases, articleActionCases()...)
}

func TestAdminActionWritten(t *testing.T) {
	Convey("Given each admin review operation", t, func() {
		for _, tc := range adminActionCases() {
			Convey(tc.name+" writes exactly one audit entry attributed to the admin", func() {
				var got []*auditdomain.AdminAction
				So(tc.run(capturingAdminActions(&got, nil)), ShouldBeNil)
				So(got, ShouldResemble, []*auditdomain.AdminAction{&tc.want})
			})
		}
	})
}

func TestAdminActionFailureIsReturned(t *testing.T) {
	Convey("Given each admin review operation", t, func() {
		for _, tc := range adminActionCases() {
			Convey(tc.name+" returns the audit error", func() {
				var got []*auditdomain.AdminAction
				err := tc.run(capturingAdminActions(&got, testutil.ErrDBUnexpected))
				So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			})
		}
	})
}

func courseSvc(repo *mockReviewRepo, actions *mockAdminActionRepo) *Service {
	return newTestServiceWithActions(repo, &mockReviewRepo{}, &mockReviewRepo{}, actions)
}

func contentSvc(repo *mockReviewRepo, actions *mockAdminActionRepo) *Service {
	return newTestServiceWithActions(&mockReviewRepo{}, repo, &mockReviewRepo{}, actions)
}

func articleSvc(repo *mockReviewRepo, actions *mockAdminActionRepo) *Service {
	return newTestServiceWithActions(&mockReviewRepo{}, &mockReviewRepo{}, repo, actions)
}
