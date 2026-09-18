package articleservice

import (
	"context"
	"fmt"
	articledomain "learnflow_backend/internal/article/domain"
)

// PublishArticle publishes a draft Article, provided its article is ready.
func (s *Service) PublishArticle(ctx context.Context, articleID, userID string) error {
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

		err = s.articleRepo.PublishArticle(ctx, articleID, userID)
		if err != nil {
			return fmt.Errorf("service.PublishArticle: %w", err)
		}
		return nil
	})
}
