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

		// TODO(notifications module, Phase 3+): recipient unset — nothing to wire to yet;
		// NotificationWorker must build "courseUrl" from Data["slug"] + base URL (same
		// pattern as internal/worker/email_verification.go's verificationUrl) before Send.
		// *course.Description is safe to dereference only because ReadyToPublish above
		// already guarantees it is non-nil/non-empty (checkDescriptionReady).
		payload := events.NotificationSendPayload{
			Template: "course_published.html",
			Data: map[string]string{
				"title":       course.Title,
				"description": *course.Description,
				"slug":        course.Slug,
			},
		}

		return s.outbox.Emit(ctx, events.AggregationTypeNotification, courseID, events.EventNotificationSend, payload)
	})
}
