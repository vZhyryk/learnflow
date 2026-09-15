package articleservice

import (
	"context"
	"errors"
	"fmt"
	articledomain "learnflow_backend/internal/article/domain"
)

// UpdateArticle applies a partial update to an existing article, re-checking slug uniqueness if it changed.
func (s *Service) UpdateArticle(ctx context.Context, req articledomain.UpdateArticleRequest, userID string) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		article, err := s.articleRepo.GetArticleByID(ctx, req.ID)
		if err != nil {
			return fmt.Errorf("service.UpdateArticle: %w", err)
		}

		if req.Slug != nil && *req.Slug != article.Slug {
			existing, err := s.articleRepo.GetArticleBySlug(ctx, *req.Slug)
			if err != nil && !errors.Is(err, articledomain.ErrArticleNotFound) {
				return fmt.Errorf("service.UpdateArticle: %w", err)
			}
			if existing != nil && existing.ID != article.ID {
				return articledomain.ErrInvalidSlug
			}
		}
		req.Apply(article)
		if err := s.articleRepo.UpdateArticle(ctx, article, userID); err != nil {
			return fmt.Errorf("service.UpdateArticle: update: %w", err)
		}
		return nil
	})
}
