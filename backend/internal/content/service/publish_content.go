package contentservice

import (
	"context"
	"fmt"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/events"
)

// PublishContentItem publishes a draft contentItem, provided its content is ready, and emits a
// notification event in the same transaction.
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

		// TODO(notifications module, Phase 3+): recipient unset — nothing to wire to yet;
		// NotificationWorker must build "contentUrl" from Data["slug"] + base URL (same
		// pattern as internal/worker/email_verification.go's verificationUrl) before Send.
		// *contentItem.Description is safe to dereference only because ReadyToPublish above
		// already guarantees it is non-nil/non-empty (checkDescriptionReady).
		payload := events.NotificationSendPayload{
			Template: "content_published.html",
			Data: map[string]string{
				"title":       contentItem.Title,
				"description": *contentItem.Description,
				"slug":        contentItem.Slug,
			},
		}

		return s.outbox.Emit(ctx, events.AggregationTypeNotification, contentItemID, events.EventNotificationSend, payload)
	})
}
