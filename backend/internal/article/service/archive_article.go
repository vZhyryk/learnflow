package articleservice

import (
	"context"
	"fmt"
)

// ArchiveArticle marks a article as archived.
func (s *Service) ArchiveArticle(ctx context.Context, articleID string) error {
	err := s.articleRepo.ArchiveArticle(ctx, articleID)
	if err != nil {
		return fmt.Errorf("service.ArchiveArticle: %w", err)
	}

	return nil
}
