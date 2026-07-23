package courseservice

import (
	"context"
	"fmt"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/pagination"
)

// GetAllCourses returns courses filtered by status, or every course if getType is empty/unknown.
func (s *Service) GetAllCourses(ctx context.Context, getType coursedomain.CourseStatus, params pagination.Params) (courseList []*coursedomain.Course, err error) {
	switch getType {
	case coursedomain.ArchivedStatus:
		courseList, err = s.courseRepo.GetAllArchivedCourses(ctx, params)
	case coursedomain.PublishedStatus:
		courseList, err = s.courseRepo.GetAllPublishedCourses(ctx, params)
	case coursedomain.DraftStatus:
		courseList, err = s.courseRepo.GetAllDraftCourses(ctx, params)
	default:
		courseList, err = s.courseRepo.GetAllCourses(ctx, params)
	}

	if err != nil {
		return nil, fmt.Errorf("service.GetAllCourses: %w", err)
	}
	return courseList, nil
}
