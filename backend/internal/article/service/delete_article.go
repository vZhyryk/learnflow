package articleservice

import (
	"context"
	"fmt"
)

// DeleteArticle soft-deletes a article.
func (s *Service) DeleteArticle(ctx context.Context, articleID, userID string) error {
	err := s.articleRepo.DeleteArticle(ctx, articleID, userID)
	if err != nil {
		return fmt.Errorf("service.DeleteArticle: %w", err)
	}

	return nil
}
