package contentservice

import (
	"context"
	"fmt"
	contentdomain "learnflow_backend/internal/content/domain"
)

// GetContentItemBySlug returns a published content item by its slug.
func (s *Service) GetContentItemBySlug(ctx context.Context, slug string) (*contentdomain.ContentItem, error) {
	contentItem, err := s.contentRepo.GetContentItemBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("service.GetContentItemBySlug: %w", err)
	}
	if contentItem.Status != contentdomain.PublishedStatus {
		return nil, contentdomain.ErrContentItemNotFound
	}
	return contentItem, nil
}
