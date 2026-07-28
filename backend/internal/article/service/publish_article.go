package articleservice

import (
	"context"
	"fmt"
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/events"
)

// PublishArticle publishes a draft Article, provided its article is ready, and emits a
// notification event in the same transaction.
func (s *Service) PublishArticle(ctx context.Context, articleID string) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		article, err := s.articleRepo.GetArticleByID(ctx, articleID)
		if err != nil {
			return fmt.Errorf("service.PublishArticle: get article: %w", err)
		}

		if article.Status != articledomain.DraftStatus {
			return fmt.Errorf("service.PublishArticle: %w", articledomain.ErrInvalidArticleStatus)
		}

		err = article.ReadyToPublish()
		if err != nil {
			return fmt.Errorf("service.PublishArticle: %w", err)
		}

		err = s.articleRepo.PublishArticle(ctx, articleID)
		if err != nil {
			return fmt.Errorf("service.PublishArticle: %w", err)
		}

		// TODO(notifications module, Phase 3+): recipient unset — nothing to wire to yet;
		// NotificationWorker must build "articleUrl" from Data["slug"] + base URL (same
		// pattern as internal/worker/email_verification.go's verificationUrl) before Send.
		// *article.Excerpt is safe to dereference only because ReadyToPublish above
		// already guarantees it is non-nil/non-empty (checkExcerpt).
		payload := events.NotificationSendPayload{
			Template: "article_published.html",
			Data: map[string]string{
				"title":   article.Title,
				"excerpt": *article.Excerpt,
				"slug":    article.Slug,
			},
		}

		return s.outbox.Emit(ctx, events.AggregationTypeNotification, articleID, events.EventNotificationSend, payload)
	})
}
