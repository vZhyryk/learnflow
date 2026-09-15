package reviewdomain

import (
	"learnflow_backend/internal/shared/validator"
	"strings"
	"time"
	"unicode/utf8"
)

// CourseReview is a user's rating and comment on a course.
type CourseReview struct {
	ID              string     `json:"id"`
	CourseID        string     `json:"course_id"`
	UserID          string     `json:"user_id"`
	Rating          int        `json:"rating"`
	Comment         *string    `json:"comment"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
	DeletedByUserID *string    `json:"deleted_by_user_id"`
}

// CreateCourseReviewRequest is the input for creating a CourseReview.
type CreateCourseReviewRequest struct {
	CourseID string  `json:"course_id"`
	UserID   string  `json:"user_id"`
	Rating   int     `json:"rating"`
	Comment  *string `json:"comment"`
}

// Validate checks that all fields of the request are valid.
func (req *CreateCourseReviewRequest) Validate() error {
	checks := []func() error{
		req.validateCourseID,
		req.validateRating,
		req.validateComment,
		req.validateUserID,
	}

	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *CreateCourseReviewRequest) validateCourseID() error {
	if req.CourseID == "" || !validator.IsValidUUID(req.CourseID) {
		return ErrInvalidCourseID
	}
	return nil
}
func (req *CreateCourseReviewRequest) validateRating() error {
	if req.Rating < 1 || req.Rating > 5 {
		return ErrInvalidRating
	}
	return nil
}

func (req *CreateCourseReviewRequest) validateComment() error {
	if req.Comment == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*req.Comment)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > 2000 {
		return ErrInvalidComment
	}
	return nil
}

func (req *CreateCourseReviewRequest) validateUserID() error {
	if req.UserID == "" || !validator.IsValidUUID(req.UserID) {
		return ErrInvalidUserID
	}
	return nil
}

// UpdateCourseReviewRequest is the input for updating a CourseReview.
type UpdateCourseReviewRequest struct {
	ReviewID string  `json:"review_id"`
	UserID   string  `json:"user_id"`
	Rating   *int    `json:"rating"`
	Comment  *string `json:"comment"`
}

// Validate checks that all fields of the request are valid.
func (req *UpdateCourseReviewRequest) Validate() error {
	checks := []func() error{
		req.validateReviewID,
		req.validateRating,
		req.validateComment,
	}

	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *UpdateCourseReviewRequest) validateReviewID() error {
	if req.ReviewID == "" || !validator.IsValidUUID(req.ReviewID) {
		return ErrInvalidReviewID
	}
	return nil
}

func (req *UpdateCourseReviewRequest) validateRating() error {
	if req.Rating != nil && (*req.Rating < 1 || *req.Rating > 5) {
		return ErrInvalidRating
	}
	return nil
}

func (req *UpdateCourseReviewRequest) validateComment() error {
	if req.Comment == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*req.Comment)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > 2000 {
		return ErrInvalidComment
	}
	return nil
}

func (r UpdateCourseReviewRequest) Apply(p *CourseReview) {
	appliers := []func(*CourseReview){
		r.applyRating,
		r.applyComment,
	}
	for _, apply := range appliers {
		apply(p)
	}
}

func (r UpdateCourseReviewRequest) applyRating(p *CourseReview) {
	if r.Rating != nil {
		p.Rating = *r.Rating
	}
}

func (r UpdateCourseReviewRequest) applyComment(p *CourseReview) {
	if r.Comment != nil {
		p.Comment = r.Comment
	}
}

// ContentReview is a user's rating and comment on a content item.
type ContentReview struct {
	ID              string     `json:"id"`
	ContentID       string     `json:"content_id"`
	UserID          string     `json:"user_id"`
	Rating          int        `json:"rating"`
	Comment         *string    `json:"comment"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
	DeletedByUserID *string    `json:"deleted_by_user_id"`
}

// CreateContentReviewRequest is the input for creating a ContentReview.
type CreateContentReviewRequest struct {
	ContentID string  `json:"content_id"`
	UserID    string  `json:"user_id"`
	Rating    int     `json:"rating"`
	Comment   *string `json:"comment"`
}

// Validate checks that all fields of the request are valid.
func (req *CreateContentReviewRequest) Validate() error {
	checks := []func() error{
		req.validateContentID,
		req.validateRating,
		req.validateComment,
		req.validateUserID,
	}

	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *CreateContentReviewRequest) validateContentID() error {
	if req.ContentID == "" || !validator.IsValidUUID(req.ContentID) {
		return ErrInvalidContentItemID
	}
	return nil
}
func (req *CreateContentReviewRequest) validateRating() error {
	if req.Rating < 1 || req.Rating > 5 {
		return ErrInvalidRating
	}
	return nil
}

func (req *CreateContentReviewRequest) validateComment() error {
	if req.Comment == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*req.Comment)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > 2000 {
		return ErrInvalidComment
	}
	return nil
}

func (req *CreateContentReviewRequest) validateUserID() error {
	if req.UserID == "" || !validator.IsValidUUID(req.UserID) {
		return ErrInvalidUserID
	}
	return nil
}

// UpdateContentReviewRequest is the input for updating a ContentReview.
type UpdateContentReviewRequest struct {
	ReviewID string  `json:"review_id"`
	UserID   string  `json:"user_id"`
	Rating   *int    `json:"rating"`
	Comment  *string `json:"comment"`
}

// Validate checks that all fields of the request are valid.
func (req *UpdateContentReviewRequest) Validate() error {
	checks := []func() error{
		req.validateReviewID,
		req.validateRating,
		req.validateComment,
	}

	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *UpdateContentReviewRequest) validateReviewID() error {
	if req.ReviewID == "" || !validator.IsValidUUID(req.ReviewID) {
		return ErrInvalidReviewID
	}
	return nil
}

func (req *UpdateContentReviewRequest) validateRating() error {
	if req.Rating != nil && (*req.Rating < 1 || *req.Rating > 5) {
		return ErrInvalidRating
	}
	return nil
}

func (req *UpdateContentReviewRequest) validateComment() error {
	if req.Comment == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*req.Comment)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > 2000 {
		return ErrInvalidComment
	}
	return nil
}

func (r UpdateContentReviewRequest) Apply(p *ContentReview) {
	appliers := []func(*ContentReview){
		r.applyRating,
		r.applyComment,
	}
	for _, apply := range appliers {
		apply(p)
	}
}

func (r UpdateContentReviewRequest) applyRating(p *ContentReview) {
	if r.Rating != nil {
		p.Rating = *r.Rating
	}
}

func (r UpdateContentReviewRequest) applyComment(p *ContentReview) {
	if r.Comment != nil {
		p.Comment = r.Comment
	}
}

type ReviewFilter struct {
	Rating int
	Op     string
}

func (f *ReviewFilter) IsUsed() bool {
	validOps := map[string]bool{
		"gt":  true,
		"lt":  true,
		"eq":  true,
		"gte": true,
		"lte": true,
	}
	if !validOps[f.Op] {
		return false
	}

	if f.Rating < 1 || f.Rating > 5 {
		return false
	}

	return true
}
