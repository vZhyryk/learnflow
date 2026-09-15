package contentservice

import (
	"context"
	"fmt"
)

// DeleteContentItem soft-deletes a content item.
func (s *Service) DeleteContentItem(ctx context.Context, contentItemID, userID string) error {
	err := s.contentRepo.DeleteContentItem(ctx, contentItemID, userID)
	if err != nil {
		return fmt.Errorf("service.DeleteContentItem: %w", err)
	}

	return nil
}
