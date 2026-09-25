package articleservice

import (
	"context"
	"fmt"
	auditdomain "learnflow_backend/internal/audit/domain"
)

// DeleteArticle soft-deletes a article.
func (s *Service) DeleteArticle(ctx context.Context, articleID, userID string) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		if err := s.articleRepo.DeleteArticle(ctx, articleID, userID); err != nil {
			return fmt.Errorf("service.DeleteArticle: %w", err)
		}

		return s.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
			AdminUserID: userID,
			ActionType:  auditdomain.ActionDeleteItem,
			TargetType:  auditdomain.TargetArticle,
			TargetID:    articleID,
		})
	})
}
