package contentservice

import (
	"context"
	"errors"
	"fmt"
	auditdomain "learnflow_backend/internal/audit/domain"
	contentdomain "learnflow_backend/internal/content/domain"
)

// CreateContentItem creates a new draft content item after checking the slug is not already in use.
func (s *Service) CreateContentItem(ctx context.Context, req contentdomain.CreateContentItemRequest) (string, error) {
	var contentItemID string
	err := s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		existing, err := s.contentRepo.GetContentItemBySlug(ctx, req.Slug)
		if err != nil && !errors.Is(err, contentdomain.ErrContentItemNotFound) {
			return fmt.Errorf("service.CreateContentItem: %w", err)
		}
		if existing != nil {
			return contentdomain.ErrInvalidSlug
		}

		isIndexable := true
		if req.IsIndexable != nil {
			isIndexable = *req.IsIndexable
		}

		contentItem := contentdomain.ContentItem{
			Slug:             req.Slug,
			Title:            req.Title,
			Description:      req.Description,
			ThumbnailURL:     req.ThumbnailURL,
			EstimatedMinutes: req.EstimatedMinutes,
			SeoTitle:         req.SeoTitle,
			SeoDescription:   req.SeoDescription,
			OgImageURL:       req.OgImageURL,
			CanonicalURL:     req.CanonicalURL,
			IsIndexable:      isIndexable,
			CreatedByUserID:  req.CreatedByUserID,
			ContentType:      req.ContentType,
			Body:             req.Body,
			VideoURL:         req.VideoURL,
			FileURL:          req.FileURL,
			EstimatedPages:   req.EstimatedPages,
		}

		createdContentItem, err := s.contentRepo.CreateContentItem(ctx, &contentItem)
		if err != nil {
			return fmt.Errorf("service.CreateContentItem: %w", err)
		}

		contentItemID = createdContentItem.ID
		return s.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
			AdminUserID: req.CreatedByUserID,
			ActionType:  auditdomain.ActionCreateItem,
			TargetType:  auditdomain.TargetContentItem,
			TargetID:    contentItemID,
		})
	})

	return contentItemID, err
}
