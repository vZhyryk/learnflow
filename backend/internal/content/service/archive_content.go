package contentservice

import (
	"context"
	"fmt"
	auditdomain "learnflow_backend/internal/audit/domain"
)

// ArchiveContentItem marks a content item as archived. Unlike PublishContentItem, intentionally
// allowed from any status — archiving is a takedown action, not a state-machine step.
func (s *Service) ArchiveContentItem(ctx context.Context, contentItemID, userID string) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		if err := s.contentRepo.ArchiveContentItem(ctx, contentItemID, userID); err != nil {
			return fmt.Errorf("service.ArchiveContentItem: %w", err)
		}

		return s.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
			AdminUserID: userID,
			ActionType:  auditdomain.ActionArchiveItem,
			TargetType:  auditdomain.TargetContentItem,
			TargetID:    contentItemID,
		})
	})
}
