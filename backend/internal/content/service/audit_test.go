package contentservice

import (
	"context"
	"errors"
	auditdomain "learnflow_backend/internal/audit/domain"
	contentdomain "learnflow_backend/internal/content/domain"
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
	existing := func(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
		return &contentdomain.ContentItem{ID: "target-1", Slug: "slug"}, nil
	}

	return []adminActionCase{
		{
			name: "ContentItem create",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockContentItemRepo{
					getContentItemBySlug: func(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
						return nil, contentdomain.ErrContentItemNotFound
					},
					createContentItem: func(_ context.Context, _ *contentdomain.ContentItem) (*contentdomain.ContentItem, error) {
						return &contentdomain.ContentItem{ID: "target-1"}, nil
					},
				}
				_, err := newTestServiceWithActions(repo, actions).CreateContentItem(ctx, contentdomain.CreateContentItemRequest{CreatedByUserID: "admin-1"})
				return err
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionCreateItem, TargetType: auditdomain.TargetContentItem, TargetID: "target-1"},
		},
		{
			name: "ContentItem update",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockContentItemRepo{
					getContentItemByID: existing,
					updateContentItem:  alwaysSucceedsUpdate,
				}
				return newTestServiceWithActions(repo, actions).UpdateContentItem(ctx, contentdomain.UpdateContentItemRequest{ID: "target-1"}, "admin-1")
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionUpdateItem, TargetType: auditdomain.TargetContentItem, TargetID: "target-1"},
		},
		{
			name: "ContentItem publish",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockContentItemRepo{
					getContentItemByID: validGetContentItemByID,
					publishContentItem: func(_ context.Context, _, _ string) error { return nil },
				}
				return newTestServiceWithActions(repo, actions).PublishContentItem(ctx, "target-1", "admin-1")
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionPublishItem, TargetType: auditdomain.TargetContentItem, TargetID: "target-1"},
		},
		{
			name: "ContentItem archive",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockContentItemRepo{archiveContentItem: func(_ context.Context, _, _ string) error { return nil }}
				return newTestServiceWithActions(repo, actions).ArchiveContentItem(ctx, "target-1", "admin-1")
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionArchiveItem, TargetType: auditdomain.TargetContentItem, TargetID: "target-1"},
		},
		{
			name: "ContentItem delete",
			run: func(actions *mockAdminActionRepo) error {
				repo := &mockContentItemRepo{deleteContentItem: func(_ context.Context, _, _ string) error { return nil }}
				return newTestServiceWithActions(repo, actions).DeleteContentItem(ctx, "target-1", "admin-1")
			},
			want: auditdomain.AdminAction{AdminUserID: "admin-1", ActionType: auditdomain.ActionDeleteItem, TargetType: auditdomain.TargetContentItem, TargetID: "target-1"},
		},
	}
}

func TestAdminActionWritten(t *testing.T) {
	Convey("Given each mutating ContentItem operation", t, func() {
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
	Convey("Given each mutating ContentItem operation", t, func() {
		for _, tc := range adminActionCases() {
			Convey(tc.name+" returns the audit error", func() {
				var got []*auditdomain.AdminAction
				err := tc.run(capturingAdminActions(&got, testutil.ErrDBUnexpected))
				So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			})
		}
	})
}
