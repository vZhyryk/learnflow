package courseservice

import (
	"context"
	"fmt"
	auditdomain "learnflow_backend/internal/audit/domain"
	coursedomain "learnflow_backend/internal/courses/domain"
)

// PublishCourse publishes a draft course, provided its content is ready.
func (s *Service) PublishCourse(ctx context.Context, courseID, userID string) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		course, err := s.courseRepo.GetCourseByID(ctx, courseID)
		if err != nil {
			return fmt.Errorf("service.PublishCourse: get course: %w", err)
		}

		if course.Status != coursedomain.DraftStatus {
			return fmt.Errorf("service.PublishCourse: %w", coursedomain.ErrInvalidCourseStatus)
		}

		err = course.ReadyToPublish()
		if err != nil {
			return fmt.Errorf("service.ReadyToPublish: %w", err)
		}

		err = s.courseRepo.PublishCourse(ctx, courseID, userID)
		if err != nil {
			return fmt.Errorf("service.PublishCourse: %w", err)
		}

		return s.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
			AdminUserID: userID,
			ActionType:  auditdomain.ActionPublishItem,
			TargetType:  auditdomain.TargetCourse,
			TargetID:    courseID,
		})
	})
}
