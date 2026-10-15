package contentservice

import (
	"context"
	"fmt"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/shared/pagination"
)

// GetAllContentItems returns content items filtered by status (archived, published, draft, or all).
func (s *Service) GetAllContentItems(ctx context.Context, getType contentdomain.ContentItemStatus, params pagination.Params) (contentItemList []*contentdomain.ContentItem, total int, err error) {
	switch getType {
	case contentdomain.ArchivedStatus:
		contentItemList, total, err = s.contentRepo.GetAllArchivedContentItems(ctx, params)
	case contentdomain.PublishedStatus:
		contentItemList, total, err = s.contentRepo.GetAllPublishedContentItems(ctx, params)
	case contentdomain.DraftStatus:
		contentItemList, total, err = s.contentRepo.GetAllDraftContentItems(ctx, params)
	default:
		contentItemList, total, err = s.contentRepo.GetAllContentItems(ctx, params)
	}

	if err != nil {
		return nil, 0, fmt.Errorf("service.GetAllContentItems: %w", err)
	}
	return contentItemList, total, nil
}
