package courseservice

import (
	"context"
	"fmt"
)

// DeleteCourse soft-deletes a course. Intentionally allowed from any current status, same
// rationale as ArchiveCourse — deletion is a takedown action, not a publish-state transition.
func (s *Service) DeleteCourse(ctx context.Context, courseID string) error {
	err := s.courseRepo.DeleteCourse(ctx, courseID)
	if err != nil {
		return fmt.Errorf("service.DeleteCourse: %w", err)
	}

	return nil
}
