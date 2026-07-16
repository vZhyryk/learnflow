package courseservice

import (
	"context"
	"fmt"
	coursedomain "learnflow_backend/internal/courses/domain"
)

// GetAllCourses returns courses filtered by status, or every course if getType is empty/unknown.
func (s *Service) GetAllCourses(ctx context.Context, getType coursedomain.CourseStatus) (courseList []*coursedomain.Course, err error) {
	switch getType {
	case coursedomain.ArchivedStatus:
		courseList, err = s.courseRepo.GetAllArchivedCourses(ctx)
	case coursedomain.PublishedStatus:
		courseList, err = s.courseRepo.GetAllPublishedCourses(ctx)
	case coursedomain.DraftStatus:
		courseList, err = s.courseRepo.GetAllDraftCourses(ctx)
	default:
		courseList, err = s.courseRepo.GetAllCourses(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("service.GetAllCourses: %w", err)
	}
	return courseList, nil
}
