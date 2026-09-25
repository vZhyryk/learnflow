package contentdomain

import (
	"context"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/pagination"
)

// Transactor executes a function within a database transaction.
type Transactor interface {
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// ContentRepository defines persistence operations for ContentItem.
type ContentRepository interface {
	CreateContentItem(ctx context.Context, contentItem *ContentItem) (*ContentItem, error)
	PublishContentItem(ctx context.Context, contentItemID, userID string) error
	ArchiveContentItem(ctx context.Context, contentItemID, userID string) error
	DeleteContentItem(ctx context.Context, contentItemID, userID string) error
	UpdateContentItem(ctx context.Context, contentItem *ContentItem, userID string) error
	GetAllPublishedContentItems(ctx context.Context, params pagination.Params) ([]*ContentItem, error)
	GetAllDraftContentItems(ctx context.Context, params pagination.Params) ([]*ContentItem, error)
	GetAllArchivedContentItems(ctx context.Context, params pagination.Params) ([]*ContentItem, error)
	GetAllContentItems(ctx context.Context, params pagination.Params) ([]*ContentItem, error)
	GetContentItemByID(ctx context.Context, contentItemID string) (*ContentItem, error)
	GetContentItemBySlug(ctx context.Context, slug string) (*ContentItem, error)
	CheckIfContentItemExistsByID(ctx context.Context, contentItemID string) (bool, error)
}

// AdminActionRepository persists the admin audit trail.
type AdminActionRepository interface {
	CreateAdminAction(ctx context.Context, action *auditdomain.AdminAction) error
}

// Service defines the content module's business logic operations.
type Service interface {
	ArchiveContentItem(ctx context.Context, contentItemID, userID string) error
	CreateContentItem(ctx context.Context, req CreateContentItemRequest) (string, error)
	DeleteContentItem(ctx context.Context, contentItemID, userID string) error
	GetContentItemBySlug(ctx context.Context, slug string) (*ContentItem, error)
	PublishContentItem(ctx context.Context, contentItemID, userID string) error
	UpdateContentItem(ctx context.Context, req UpdateContentItemRequest, userID string) error
	GetAllContentItems(ctx context.Context, getType ContentItemStatus, params pagination.Params) (contentItemList []*ContentItem, err error)
}
