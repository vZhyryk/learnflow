package contentrepository

import (
	"context"
	"errors"
	"fmt"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5"
)

// contentItemsSlugUniqueConstraint is the DB-level backstop for the check-then-insert slug race.
const contentItemsSlugUniqueConstraint = "content_items_slug_unique"

// CreateContentItem inserts a new draft ContentItem.
func (rep *Repository) CreateContentItem(ctx context.Context, contentItem *contentdomain.ContentItem) (*contentdomain.ContentItem, error) {
	contentItem, err := scanContentItem(rep.QueryRunner(ctx).QueryRow(ctx, createDraftContentItemSQL, contentItem.Slug, contentItem.Title, contentItem.ContentType, contentItem.Description, contentItem.Body, contentItem.VideoURL, contentItem.FileURL, contentItem.ThumbnailURL, contentItem.EstimatedMinutes, contentItem.EstimatedPages, contentItem.SeoTitle, contentItem.SeoDescription, contentItem.OgImageURL, contentItem.CanonicalURL, contentItem.IsIndexable, contentItem.CreatedByUserID))
	if db.IsUniqueViolation(err, contentItemsSlugUniqueConstraint) {
		return nil, contentdomain.ErrInvalidSlug
	}
	if err != nil {
		return nil, fmt.Errorf("repository.CreateContentItem: %w", err)
	}

	return contentItem, nil
}

// PublishContentItem marks a ContentItem as published.
func (rep *Repository) PublishContentItem(ctx context.Context, contentItemID string) error {
	return repository.ExecUpdateByID(ctx, &rep.BaseRepository, publishContentItemSQL, "PublishContentItem", contentItemID, contentdomain.ErrContentItemNotFound)
}

// ArchiveContentItem marks a ContentItem as archived.
func (rep *Repository) ArchiveContentItem(ctx context.Context, contentItemID string) error {
	return repository.ExecUpdateByID(ctx, &rep.BaseRepository, archiveContentItemSQL, "ArchiveContentItem", contentItemID, contentdomain.ErrContentItemNotFound)
}

// DeleteContentItem soft-deletes a ContentItem.
func (rep *Repository) DeleteContentItem(ctx context.Context, contentItemID string) error {
	return repository.ExecUpdateByID(ctx, &rep.BaseRepository, deleteContentItemSQL, "DeleteContentItem", contentItemID, contentdomain.ErrContentItemNotFound)
}

// UpdateContentItem persists changes to an existing ContentItem.
func (rep *Repository) UpdateContentItem(ctx context.Context, contentItem *contentdomain.ContentItem) error {
	tag, err := rep.QueryRunner(ctx).Exec(ctx, updateContentItemSQL, contentItem.ID, contentItem.Slug, contentItem.Title, contentItem.Description, contentItem.Body, contentItem.VideoURL, contentItem.FileURL, contentItem.EstimatedMinutes, contentItem.EstimatedPages, contentItem.ThumbnailURL, contentItem.SeoTitle, contentItem.SeoDescription, contentItem.OgImageURL, contentItem.CanonicalURL, contentItem.IsIndexable)
	if db.IsUniqueViolation(err, contentItemsSlugUniqueConstraint) {
		return contentdomain.ErrInvalidSlug
	}
	if err != nil {
		return fmt.Errorf("repository.UpdateContentItem: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return contentdomain.ErrContentItemNotFound
	}

	return nil
}

// GetAllPublishedContentItems returns every non-deleted published ContentItem.
func (rep *Repository) GetAllPublishedContentItems(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getAllPublishedContentItemsSQL, "GetAllPublishedContentItems", params, scanContentItem)
}

// GetAllDraftContentItems returns every non-deleted draft ContentItem.
func (rep *Repository) GetAllDraftContentItems(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getAllDraftContentItemsSQL, "GetAllDraftContentItems", params, scanContentItem)
}

// GetAllArchivedContentItems returns every archived ContentItem, including soft-deleted ones.
func (rep *Repository) GetAllArchivedContentItems(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getAllArchivedContentItemsSQL, "GetAllArchivedContentItems", params, scanContentItem)
}

// GetAllContentItems returns every ContentItem regardless of status, including soft-deleted ones.
func (rep *Repository) GetAllContentItems(ctx context.Context, params pagination.Params) ([]*contentdomain.ContentItem, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getAllContentItemsSQL, "GetAllContentItems", params, scanContentItem)
}

// GetContentItemByID retrieves a non-deleted ContentItem by ID.
func (rep *Repository) GetContentItemByID(ctx context.Context, contentItemID string) (*contentdomain.ContentItem, error) {
	contentItem, err := scanContentItem(rep.QueryRunner(ctx).QueryRow(ctx, getContentItemByIDSQL, contentItemID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, contentdomain.ErrContentItemNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("repository.GetContentItemByID: %w", err)
	}

	return contentItem, nil
}

// GetContentItemBySlug retrieves a non-deleted ContentItem by slug.
func (rep *Repository) GetContentItemBySlug(ctx context.Context, slug string) (*contentdomain.ContentItem, error) {
	contentItem, err := scanContentItem(rep.QueryRunner(ctx).QueryRow(ctx, getContentItemBySlugSQL, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, contentdomain.ErrContentItemNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("repository.GetContentItemBySlug: %w", err)
	}

	return contentItem, nil
}
