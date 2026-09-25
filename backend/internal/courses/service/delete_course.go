package courseservice

import (
	"context"
	"fmt"
	auditdomain "learnflow_backend/internal/audit/domain"
)

// DeleteCourse soft-deletes a course. Intentionally allowed from any current status, same
// rationale as ArchiveCourse — deletion is a takedown action, not a publish-state transition.
func (s *Service) DeleteCourse(ctx context.Context, courseID, userID string) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		if err := s.courseRepo.DeleteCourse(ctx, courseID, userID); err != nil {
			return fmt.Errorf("service.DeleteCourse: %w", err)
		}

		return s.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
			AdminUserID: userID,
			ActionType:  auditdomain.ActionDeleteItem,
			TargetType:  auditdomain.TargetCourse,
			TargetID:    courseID,
		})
	})
}
