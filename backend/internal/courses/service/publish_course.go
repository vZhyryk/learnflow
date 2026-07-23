package courseservice

import (
	"context"
	"fmt"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/events"
)

// PublishCourse publishes a draft course, provided its content is ready, and emits a
// notification event in the same transaction.
func (s *Service) PublishCourse(ctx context.Context, courseID string) error {
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

		err = s.courseRepo.PublishCourse(ctx, courseID)
		if err != nil {
			return fmt.Errorf("service.PublishCourse: %w", err)
		}

		// TODO(notifications module, Phase 3+): Template/recipient unset — nothing to
		// wire to yet; fill in once the notifications module lands.
		payload := events.NotificationSendPayload{
			Template: "",
			Data: map[string]string{
				"title":       course.Title,
				"description": *course.Description,
			},
		}

		return s.outbox.Emit(ctx, events.AggregationTypeNotification, courseID, events.EventNotificationSend, payload)
	})
}
