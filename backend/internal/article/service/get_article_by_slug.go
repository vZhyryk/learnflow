package articleservice

import (
	"context"
	"fmt"
	articledomain "learnflow_backend/internal/article/domain"
)

// GetArticleBySlug returns a published article by its slug.
func (s *Service) GetArticleBySlug(ctx context.Context, slug string) (*articledomain.Article, error) {
	article, err := s.articleRepo.GetArticleBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("service.GetArticleBySlug: %w", err)
	}
	if article.Status != articledomain.PublishedStatus {
		return nil, articledomain.ErrArticleNotFound
	}
	return article, nil
}
