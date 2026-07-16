package courseservice

import (
	"context"
	"fmt"
)

// ArchiveCourse marks a course as archived.
func (s *Service) ArchiveCourse(ctx context.Context, courseID string) error {
	err := s.courseRepo.ArchiveCourse(ctx, courseID)
	if err != nil {
		return fmt.Errorf("service.ArchiveCourse: %w", err)
	}

	return nil
}
