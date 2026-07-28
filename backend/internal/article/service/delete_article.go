package articleservice

import (
	"context"
	"fmt"
)

// DeleteArticle soft-deletes a article.
func (s *Service) DeleteArticle(ctx context.Context, articleID string) error {
	err := s.articleRepo.DeleteArticle(ctx, articleID)
	if err != nil {
		return fmt.Errorf("service.DeleteArticle: %w", err)
	}

	return nil
}
