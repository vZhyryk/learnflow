package courseservice

import (
	"context"
	"fmt"
	coursedomain "learnflow_backend/internal/courses/domain"
)

// GetCourseBySlug retrieves a course by its slug.
func (s *Service) GetCourseBySlug(ctx context.Context, slug string) (*coursedomain.Course, error) {
	course, err := s.courseRepo.GetCourseBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("service.GetCourseBySlug: %w", err)
	}
	return course, nil
}
