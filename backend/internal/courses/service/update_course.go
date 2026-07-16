package courseservice

import (
	"context"
	"errors"
	"fmt"
	coursedomain "learnflow_backend/internal/courses/domain"
)

// UpdateCourse patches an existing course, checking any new slug is not already in use.
func (s *Service) UpdateCourse(ctx context.Context, req coursedomain.UpdateCourseRequest) error {
	course, err := s.courseRepo.GetCourseByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("service.UpdateCourse: %w", err)
	}

	if req.Slug != nil && *req.Slug != course.Slug {
		existing, err := s.courseRepo.GetCourseBySlug(ctx, *req.Slug)
		if err != nil && !errors.Is(err, coursedomain.ErrCourseNotFound) {
			return fmt.Errorf("service.UpdateCourse: %w", err)
		}
		if existing != nil && existing.ID != course.ID {
			return coursedomain.ErrInvalidSlug
		}
	}

	req.Apply(course)

	return s.courseRepo.UpdateCourse(ctx, course)
}
