package courseservice

import (
	"context"
	"fmt"
	coursedomain "learnflow_backend/internal/courses/domain"
)

// GetCourseBySlug retrieves a published course by its slug (public course page).
// Slug-uniqueness checks need any status — use courseRepo.GetCourseBySlug directly for that.
func (s *Service) GetCourseBySlug(ctx context.Context, slug string) (*coursedomain.Course, error) {
	course, err := s.courseRepo.GetCourseBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("service.GetCourseBySlug: %w", err)
	}
	if course.Status != coursedomain.PublishedStatus {
		return nil, coursedomain.ErrCourseNotFound
	}
	return course, nil
}
