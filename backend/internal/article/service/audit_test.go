package articleservice

import (
	"context"
	"errors"
	articledomain "learnflow_backend/internal/article/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
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
	existing := func(_ context.Context, _ string) (*articledomain.Article, error) {
		return &articledomain.Article{ID: "target-1", Slug: "slug"}, nil
	}

	return []adminActionCase{
		{
			name: "Article create",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockArticleRepo{
					getArticleBySlug: func(_ context.Context, _ string) (*articledomain.Article, error) {
						return nil, articledomain.ErrArticleNotFound
					},
					createArticle: func(_ context.Context, _ *articledomain.Article) (*articledomain.Article, error) {
						return &articledomain.Article{ID: "target-1"}, nil
					},
				}
				_, err := newTestServiceWithActions(repo, actions).CreateArticle(ctx, articledomain.CreateArticleRequest{CreatedByUserID: "admin-1"})
				return err
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionCreateItem, TargetType: auditdomain.TargetArticle, TargetID: "target-1"},
		},
		{
			name: "Article update",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockArticleRepo{
					getArticleByID: existing,
					updateArticle:  alwaysSucceedsUpdate,
				}
				return newTestServiceWithActions(repo, actions).UpdateArticle(ctx, articledomain.UpdateArticleRequest{ID: "target-1"}, "admin-1")
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionUpdateItem, TargetType: auditdomain.TargetArticle, TargetID: "target-1"},
		},
		{
			name: "Article publish",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockArticleRepo{
					getArticleByID: validGetArticleByID,
					publishArticle: func(_ context.Context, _, _ string) error { return nil },
				}
				return newTestServiceWithActions(repo, actions).PublishArticle(ctx, "target-1", "admin-1")
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionPublishItem, TargetType: auditdomain.TargetArticle, TargetID: "target-1"},
		},
		{
			name: "Article archive",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockArticleRepo{archiveArticle: func(_ context.Context, _, _ string) error { return nil }}
				return newTestServiceWithActions(repo, actions).ArchiveArticle(ctx, "target-1", "admin-1")
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionArchiveItem, TargetType: auditdomain.TargetArticle, TargetID: "target-1"},
		},
		{
			name: "Article delete",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockArticleRepo{deleteArticle: func(_ context.Context, _, _ string) error { return nil }}
				return newTestServiceWithActions(repo, actions).DeleteArticle(ctx, "target-1", "admin-1")
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionDeleteItem, TargetType: auditdomain.TargetArticle, TargetID: "target-1"},
		},
	}
}

func TestAdminActionWritten(t *testing.T) {
	Convey("Given each mutating Article operation", t, func() {
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
	Convey("Given each mutating Article operation", t, func() {
		for _, tc := range adminActionCases() {
			Convey(tc.name+" returns the audit error", func() {
				var got []*auditdomain.AdminAction
				err := tc.run(capturingAdminActions(&got, testutil.ErrDBUnexpected))
				So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			})
		}
	})
}
