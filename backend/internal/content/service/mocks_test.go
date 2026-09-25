package contentservice

import (
	"context"
	auditdomain "learnflow_backend/internal/audit/domain"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
)

type mockContentItemRepo struct {
	createContentItem            func(ctx context.Context, contentItem *contentdomain.ContentItem) (*contentdomain.ContentItem, error)
	publishContentItem           func(ctx context.Context, contentItemID, userID string) error
	archiveContentItem           func(ctx context.Context, contentItemID, userID string) error
	deleteContentItem            func(ctx context.Context, contentItemID, userID string) error
	updateContentItem            func(ctx context.Context, contentItem *contentdomain.ContentItem, userID string) error
	getAllPublishedContentItems  func(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error)
	getAllDraftContentItems      func(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error)
	getAllArchivedContentItems   func(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error)
	getAllContentItems           func(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error)
	getContentItemByID           func(ctx context.Context, contentItemID string) (*contentdomain.ContentItem, error)
	getContentItemBySlug         func(ctx context.Context, slug string) (*contentdomain.ContentItem, error)
	checkIfContentItemExistsByID func(ctx context.Context, contentItemID string) (bool, error)
}

func (m *mockContentItemRepo) CreateContentItem(ctx context.Context, contentItem *contentdomain.ContentItem) (*contentdomain.ContentItem, error) {
	if m.createContentItem == nil {
		panic("mockContentItemRepo.CreateContentItem not set")
	}

	return m.createContentItem(ctx, contentItem)
}
func (m *mockContentItemRepo) PublishContentItem(ctx context.Context, contentItemID, userID string) error {
	if m.publishContentItem == nil {
		panic("mockContentItemRepo.PublishContentItem not set")
	}

	return m.publishContentItem(ctx, contentItemID, userID)
}
func (m *mockContentItemRepo) ArchiveContentItem(ctx context.Context, contentItemID, userID string) error {
	if m.archiveContentItem == nil {
		panic("mockContentItemRepo.ArchiveContentItem not set")
	}

	return m.archiveContentItem(ctx, contentItemID, userID)
}
func (m *mockContentItemRepo) DeleteContentItem(ctx context.Context, contentItemID, userID string) error {
	if m.deleteContentItem == nil {
		panic("mockContentItemRepo.DeleteContentItem not set")
	}

	return m.deleteContentItem(ctx, contentItemID, userID)
}
func (m *mockContentItemRepo) UpdateContentItem(ctx context.Context, contentItem *contentdomain.ContentItem, userID string) error {
	if m.updateContentItem == nil {
		panic("mockContentItemRepo.UpdateContentItem not set")
	}

	return m.updateContentItem(ctx, contentItem, userID)
}
func (m *mockContentItemRepo) GetAllPublishedContentItems(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error) {
	if m.getAllPublishedContentItems == nil {
		panic("mockContentItemRepo.GetAllPublishedContentItems not set")
	}

	return m.getAllPublishedContentItems(ctx, params)
}
func (m *mockContentItemRepo) GetAllDraftContentItems(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error) {
	if m.getAllDraftContentItems == nil {
		panic("mockContentItemRepo.GetAllDraftContentItems not set")
	}

	return m.getAllDraftContentItems(ctx, params)
}
func (m *mockContentItemRepo) GetAllArchivedContentItems(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error) {
	if m.getAllArchivedContentItems == nil {
		panic("mockContentItemRepo.GetAllArchivedContentItems not set")
	}

	return m.getAllArchivedContentItems(ctx, params)
}
func (m *mockContentItemRepo) GetAllContentItems(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error) {
	if m.getAllContentItems == nil {
		panic("mockContentItemRepo.GetAllContentItems not set")
	}

	return m.getAllContentItems(ctx, params)
}
func (m *mockContentItemRepo) GetContentItemByID(ctx context.Context, contentItemID string) (*contentdomain.ContentItem, error) {
	if m.getContentItemByID == nil {
		panic("mockContentItemRepo.GetContentItemByID not set")
	}

	return m.getContentItemByID(ctx, contentItemID)
}
func (m *mockContentItemRepo) GetContentItemBySlug(ctx context.Context, slug string) (*contentdomain.ContentItem, error) {
	if m.getContentItemBySlug == nil {
		panic("mockContentItemRepo.GetContentItemBySlug not set")
	}

	return m.getContentItemBySlug(ctx, slug)
}

func (m *mockContentItemRepo) CheckIfContentItemExistsByID(ctx context.Context, contentItemID string) (bool, error) {
	if m.checkIfContentItemExistsByID == nil {
		panic("mockContentItemRepo.checkIfContentItemExistsByID not set")
	}

	return m.checkIfContentItemExistsByID(ctx, contentItemID)
}

func newTestService(repo *mockContentItemRepo) *Service {
	return newTestServiceWithActions(repo, noopAdminActions())
}

func newTestServiceWithActions(repo *mockContentItemRepo, actions *mockAdminActionRepo) *Service {
	return New(repo, actions, &testutil.NoopTransactor{})
}

type mockAdminActionRepo struct {
	createAdminAction func(ctx context.Context, action *auditdomain.AdminAction) error
}

func (m *mockAdminActionRepo) CreateAdminAction(ctx context.Context, action *auditdomain.AdminAction) error {
	if m.createAdminAction == nil {
		panic("mockAdminActionRepo.CreateAdminAction not set")
	}

	return m.createAdminAction(ctx, action)
}

func noopAdminActions() *mockAdminActionRepo {
	return &mockAdminActionRepo{createAdminAction: func(_ context.Context, _ *auditdomain.AdminAction) error { return nil }}
}

func capturingAdminActions(got *[]*auditdomain.AdminAction, err error) *mockAdminActionRepo {
	return &mockAdminActionRepo{createAdminAction: func(_ context.Context, action *auditdomain.AdminAction) error {
		*got = append(*got, action)
		return err
	}}
}

// alwaysError is a getContentItemByID/getContentItemBySlug stub that always fails.
func alwaysError(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
	return nil, testutil.ErrDBUnexpected
}

// alwaysFailsErr is an updateContentItem stub that always fails.
func alwaysFailsErr(_ context.Context, _ *contentdomain.ContentItem, _ string) error {
	return testutil.ErrDBUnexpected
}

func alwaysSucceedsUpdate(_ context.Context, _ *contentdomain.ContentItem, _ string) error {
	return nil
}
