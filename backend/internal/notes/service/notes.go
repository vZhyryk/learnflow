package notesservice

import (
	"context"
	"fmt"
	notesdomain "learnflow_backend/internal/notes/domain"
	"learnflow_backend/internal/shared/pagination"
)

// CreateUserNotes stores a note; a linked course or content item must exist and be published.
func (srv *Service) CreateUserNotes(ctx context.Context, req notesdomain.CreateNotesRequest) (*notesdomain.UserNotes, error) {
	if err := srv.checkResourceExists(ctx, req.ResourceType, req.ResourceID); err != nil {
		return nil, fmt.Errorf("service.CreateUserNotes: %w", err)
	}

	note, err := srv.notesRepo.CreateUserNotes(ctx, &notesdomain.UserNotes{
		UserID:       req.UserID,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Title:        req.Title,
		Description:  req.Description,
		Body:         req.Body,
	})
	if err != nil {
		return nil, fmt.Errorf("service.CreateUserNotes: %w", err)
	}

	return note, nil
}

// checkResourceExists accepts a note without a linked resource and rejects a link to a missing or unpublished one.
func (srv *Service) checkResourceExists(ctx context.Context, resourceType *notesdomain.ResourceType, resourceID *string) error {
	if resourceType == nil || resourceID == nil {
		return nil
	}

	var exists bool
	var err error
	switch *resourceType {
	case notesdomain.ContentResourceType:
		exists, err = srv.contentRepo.CheckIfContentItemExistsActiveByID(ctx, *resourceID)
	case notesdomain.CourseResourceType:
		exists, err = srv.courseRepo.CheckIfCourseExistsActiveByID(ctx, *resourceID)
	}
	if err != nil {
		return err
	}

	if !exists {
		return notesdomain.ErrInvalidResourceID
	}

	return nil
}

// GetUserNotesByID implements notesdomain.Service.
func (srv *Service) GetUserNotesByID(ctx context.Context, id, userID string) (*notesdomain.UserNotes, error) {
	notes, err := srv.notesRepo.GetUserNotesByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("service.GetUserNotesByID: %w", err)
	}
	return notes, nil
}

// GetUserAllNotesByUserID implements notesdomain.Service.
func (srv *Service) GetUserAllNotesByUserID(ctx context.Context, userID, search string, params pagination.Params) ([]*notesdomain.UserNotes, error) {
	notesList, err := srv.notesRepo.GetUserAllNotesByUserID(ctx, userID, search, params)
	if err != nil {
		return nil, fmt.Errorf("service.GetUserAllNotesByUserID: %w", err)
	}
	return notesList, nil
}

// UpdateUserNotes implements notesdomain.Service.
func (srv *Service) UpdateUserNotes(ctx context.Context, req notesdomain.UpdateNotesRequest) error {
	return srv.transactor.InTransaction(ctx, func(ctx context.Context) error {
		notes, err := srv.notesRepo.GetUserNotesByID(ctx, req.ID, req.UserID)
		if err != nil {
			return fmt.Errorf("service.UpdateUserNotes: %w", err)
		}

		req.Apply(notes)

		err = srv.notesRepo.UpdateUserNotes(ctx, notes)
		if err != nil {
			return fmt.Errorf("service.UpdateUserNotes: %w", err)
		}
		return nil
	})
}

// DeleteUserNotes implements notesdomain.Service.
func (srv *Service) DeleteUserNotes(ctx context.Context, id, userID string) error {
	err := srv.notesRepo.DeleteUserNotes(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("service.DeleteUserNotes: %w", err)
	}
	return nil
}
