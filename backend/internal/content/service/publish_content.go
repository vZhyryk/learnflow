package contentservice

import (
	"context"
	"fmt"
	contentdomain "learnflow_backend/internal/content/domain"
)

// PublishContentItem publishes a draft contentItem, provided its content is ready.
func (s *Service) PublishContentItem(ctx context.Context, contentItemID, userID string) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		contentItem, err := s.contentRepo.GetContentItemByID(ctx, contentItemID)
		if err != nil {
			return fmt.Errorf("service.PublishContentItem: get contentItem: %w", err)
		}

		if contentItem.Status != contentdomain.DraftStatus {
			return fmt.Errorf("service.PublishContentItem: %w", contentdomain.ErrInvalidContentItemStatus)
		}

		err = contentItem.ReadyToPublish()
		if err != nil {
			return fmt.Errorf("service.ReadyToPublish: %w", err)
		}

		err = s.contentRepo.PublishContentItem(ctx, contentItemID, userID)
		if err != nil {
			return fmt.Errorf("service.PublishContentItem: %w", err)
		}

		return nil
	})
}
