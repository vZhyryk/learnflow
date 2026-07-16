package courseservice

import (
	"context"
	"fmt"
)

// DeleteCourse soft-deletes a course.
func (s *Service) DeleteCourse(ctx context.Context, courseID string) error {
	err := s.courseRepo.DeleteCourse(ctx, courseID)
	if err != nil {
		return fmt.Errorf("service.DeleteCourse: %w", err)
	}

	return nil
}
