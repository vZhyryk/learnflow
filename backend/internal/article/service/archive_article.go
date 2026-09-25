package articleservice

import (
	"context"
	"fmt"
	auditdomain "learnflow_backend/internal/audit/domain"
)

// ArchiveArticle marks a article as archived.
func (s *Service) ArchiveArticle(ctx context.Context, articleID, userID string) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		if err := s.articleRepo.ArchiveArticle(ctx, articleID, userID); err != nil {
			return fmt.Errorf("service.ArchiveArticle: %w", err)
		}

		return s.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
			AdminUserID: userID,
			ActionType:  auditdomain.ActionArchiveItem,
			TargetType:  auditdomain.TargetArticle,
			TargetID:    articleID,
		})
	})
}
