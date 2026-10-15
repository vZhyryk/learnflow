package courseservice

import (
	"context"
	"fmt"
	auditdomain "learnflow_backend/internal/audit/domain"
)

// ArchiveCourse marks a course as archived. Unlike PublishCourse, intentionally allowed
// from any status — archiving is a takedown action, not a state-machine step.
func (s *Service) ArchiveCourse(ctx context.Context, courseID, userID string) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		if err := s.courseRepo.ArchiveCourse(ctx, courseID, userID); err != nil {
			return fmt.Errorf("service.ArchiveCourse: %w", err)
		}

		return s.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
			AdminUserID: userID,
			ActionType:  auditdomain.ActionArchiveItem,
			TargetType:  auditdomain.TargetCourse,
			TargetID:    courseID,
		})
	})
}
