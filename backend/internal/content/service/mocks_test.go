package contentservice

import (
	"context"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
)

type mockContentItemRepo struct {
	createContentItem           func(ctx context.Context, contentItem *contentdomain.ContentItem) (*contentdomain.ContentItem, error)
	publishContentItem          func(ctx context.Context, contentItemID string) error
	archiveContentItem          func(ctx context.Context, contentItemID string) error
	deleteContentItem           func(ctx context.Context, contentItemID string) error
	updateContentItem           func(ctx context.Context, contentItem *contentdomain.ContentItem) error
	getAllPublishedContentItems func(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error)
	getAllDraftContentItems     func(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error)
	getAllArchivedContentItems  func(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error)
	getAllContentItems          func(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error)
	getContentItemByID          func(ctx context.Context, contentItemID string) (*contentdomain.ContentItem, error)
	getContentItemBySlug        func(ctx context.Context, slug string) (*contentdomain.ContentItem, error)
}

func (m *mockContentItemRepo) CreateContentItem(ctx context.Context, contentItem *contentdomain.ContentItem) (*contentdomain.ContentItem, error) {
	if m.createContentItem == nil {
		panic("mockContentItemRepo.CreateContentItem not set")
	}

	return m.createContentItem(ctx, contentItem)
}
func (m *mockContentItemRepo) PublishContentItem(ctx context.Context, contentItemID string) error {
	if m.publishContentItem == nil {
		panic("mockContentItemRepo.PublishContentItem not set")
	}

	return m.publishContentItem(ctx, contentItemID)
}
func (m *mockContentItemRepo) ArchiveContentItem(ctx context.Context, contentItemID string) error {
	if m.archiveContentItem == nil {
		panic("mockContentItemRepo.ArchiveContentItem not set")
	}

	return m.archiveContentItem(ctx, contentItemID)
}
func (m *mockContentItemRepo) DeleteContentItem(ctx context.Context, contentItemID string) error {
	if m.deleteContentItem == nil {
		panic("mockContentItemRepo.DeleteContentItem not set")
	}

	return m.deleteContentItem(ctx, contentItemID)
}
func (m *mockContentItemRepo) UpdateContentItem(ctx context.Context, contentItem *contentdomain.ContentItem) error {
	if m.updateContentItem == nil {
		panic("mockContentItemRepo.UpdateContentItem not set")
	}

	return m.updateContentItem(ctx, contentItem)
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

func newTestService(repo *mockContentItemRepo, outbox *events.OutboxWriter) *Service {
	return New(repo, &testutil.NoopTransactor{}, outbox)
}

// alwaysError is a getContentItemByID/getContentItemBySlug stub that always fails.
func alwaysError(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
	return nil, testutil.ErrDBUnexpected
}

// alwaysFailsErr is an updateContentItem stub that always fails.
func alwaysFailsErr(_ context.Context, _ *contentdomain.ContentItem) error {
	return testutil.ErrDBUnexpected
}

func alwaysSucceedsUpdate(_ context.Context, _ *contentdomain.ContentItem) error {
	return nil
}
