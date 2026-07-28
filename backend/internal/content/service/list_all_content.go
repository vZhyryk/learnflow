package contentservice

import (
	"context"
	"fmt"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/shared/pagination"
)

// GetAllContentItems returns content items filtered by status (archived, published, draft, or all).
func (s *Service) GetAllContentItems(ctx context.Context, getType contentdomain.ContentItemStatus, params pagination.Params) (contentItemList []*contentdomain.ContentItem, err error) {
	switch getType {
	case contentdomain.ArchivedStatus:
		contentItemList, err = s.contentRepo.GetAllArchivedContentItems(ctx, params)
	case contentdomain.PublishedStatus:
		contentItemList, err = s.contentRepo.GetAllPublishedContentItems(ctx, params)
	case contentdomain.DraftStatus:
		contentItemList, err = s.contentRepo.GetAllDraftContentItems(ctx, params)
	default:
		contentItemList, err = s.contentRepo.GetAllContentItems(ctx, params)
	}

	if err != nil {
		return nil, fmt.Errorf("service.GetAllContentItems: %w", err)
	}
	return contentItemList, nil
}
