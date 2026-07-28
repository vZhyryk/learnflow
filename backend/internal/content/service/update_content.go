package contentservice

import (
	"context"
	"errors"
	"fmt"
	contentdomain "learnflow_backend/internal/content/domain"
)

// UpdateContentItem applies a partial update to an existing content item, re-checking slug uniqueness if it changed.
func (s *Service) UpdateContentItem(ctx context.Context, req contentdomain.UpdateContentItemRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		contentItem, err := s.contentRepo.GetContentItemByID(ctx, req.ID)
		if err != nil {
			return fmt.Errorf("service.UpdateContentItem: %w", err)
		}

		if req.Slug != nil && *req.Slug != contentItem.Slug {
			existing, err := s.contentRepo.GetContentItemBySlug(ctx, *req.Slug)
			if err != nil && !errors.Is(err, contentdomain.ErrContentItemNotFound) {
				return fmt.Errorf("service.UpdateContentItem: %w", err)
			}
			if existing != nil && existing.ID != contentItem.ID {
				return contentdomain.ErrInvalidSlug
			}
		}
		req.Apply(contentItem)
		if err := s.contentRepo.UpdateContentItem(ctx, contentItem); err != nil {
			return fmt.Errorf("service.UpdateContentItem: update: %w", err)
		}
		return nil
	})
}
