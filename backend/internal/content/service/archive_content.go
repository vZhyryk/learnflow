package contentservice

import (
	"context"
	"fmt"
)

// ArchiveContentItem marks a content item as archived. Unlike PublishContentItem, intentionally
// allowed from any status — archiving is a takedown action, not a state-machine step.
func (s *Service) ArchiveContentItem(ctx context.Context, contentItemID, userID string) error {
	err := s.contentRepo.ArchiveContentItem(ctx, contentItemID, userID)
	if err != nil {
		return fmt.Errorf("service.ArchiveContentItem: %w", err)
	}

	return nil
}
