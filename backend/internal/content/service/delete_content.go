package contentservice

import (
	"context"
	"fmt"
	auditdomain "learnflow_backend/internal/audit/domain"
)

// DeleteContentItem soft-deletes a content item.
func (s *Service) DeleteContentItem(ctx context.Context, contentItemID, userID string) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		if err := s.contentRepo.DeleteContentItem(ctx, contentItemID, userID); err != nil {
			return fmt.Errorf("service.DeleteContentItem: %w", err)
		}

		return s.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
			AdminUserID: userID,
			ActionType:  auditdomain.ActionDeleteItem,
			TargetType:  auditdomain.TargetContentItem,
			TargetID:    contentItemID,
		})
	})
}
