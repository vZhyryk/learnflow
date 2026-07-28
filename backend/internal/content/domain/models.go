package contentdomain

import (
	"time"

	"learnflow_backend/internal/shared/validator"
)

// ContentItemStatus represents the current lifecycle state of a ContentItem.
type ContentItemStatus string

// ContentItemStatus values.
const (
	DraftStatus     ContentItemStatus = "draft"
	PublishedStatus ContentItemStatus = "published"
	ArchivedStatus  ContentItemStatus = "archived"
)

// Valid reports whether r is one of the known ContentItemStatus values.
func (r ContentItemStatus) Valid() bool {
	switch r {
	case
		DraftStatus,
		PublishedStatus,
		ArchivedStatus:
		return true
	}
	return false
}

// ContentType represents the kind of media a ContentItem holds.
type ContentType string

// ContentType values.
const (
	VideoContent        ContentType = "video"
	BookContent         ContentType = "book"
	PresentationContent ContentType = "presentation"
)

// Valid reports whether r is one of the known ContentType values.
func (r ContentType) Valid() bool {
	switch r {
	case
		VideoContent,
		BookContent,
		PresentationContent:
		return true
	}
	return false
}

// ContentItem represents a ContentItem in draft, published, or archived state.
type ContentItem struct {
	ID               string            `json:"id"`
	Slug             string            `json:"slug"`
	Title            string            `json:"title"`
	ContentType      ContentType       `json:"content_type"`
	Description      *string           `json:"description"`
	Body             *string           `json:"body"`
	VideoURL         *string           `json:"video_url"`
	FileURL          *string           `json:"file_url"`
	EstimatedMinutes *int              `json:"estimated_minutes"`
	EstimatedPages   *int              `json:"estimated_pages"`
	ThumbnailURL     *string           `json:"thumbnail_url"`
	SeoTitle         *string           `json:"seo_title"`
	SeoDescription   *string           `json:"seo_description"`
	OgImageURL       *string           `json:"og_image_url"`
	CanonicalURL     *string           `json:"canonical_url"`
	IsIndexable      bool              `json:"is_indexable"`
	Status           ContentItemStatus `json:"status"`
	CreatedByUserID  string            `json:"created_by_user_id"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	PublishedAt      *time.Time        `json:"published_at"`
	DeletedAt        *time.Time        `json:"deleted_at"`
}

// ReadyToPublish reports whether the ContentItem has all fields required to go public.
func (c *ContentItem) ReadyToPublish() error {
	checks := []func() error{
		c.checkTitleReady,
		c.checkDescriptionReady,
		c.checkThumbnailURL,
		c.checkMediaReady,
		c.checkSeoTitleReady,
		c.checkSeoDescriptionReady,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

func (c *ContentItem) checkTitleReady() error {
	if c.Title == "" {
		return ErrInvalidTitle
	}
	return nil
}

func (c *ContentItem) checkDescriptionReady() error {
	if c.Description == nil || *c.Description == "" {
		return ErrInvalidDescription
	}
	return nil
}

func (c *ContentItem) checkThumbnailURL() error {
	if c.ThumbnailURL == nil || *c.ThumbnailURL == "" {
		return ErrInvalidThumbnailURL
	}
	return nil
}

func (c *ContentItem) checkSeoTitleReady() error {
	if c.SeoTitle == nil || *c.SeoTitle == "" {
		return ErrInvalidSeoTitle
	}
	return nil
}

func (c *ContentItem) checkSeoDescriptionReady() error {
	if c.SeoDescription == nil || *c.SeoDescription == "" {
		return ErrInvalidSeoDescription
	}
	return nil
}

func (c *ContentItem) checkMediaReady() error {
	switch c.ContentType {
	case VideoContent:
		if c.VideoURL == nil || *c.VideoURL == "" {
			return ErrInvalidMedia
		}
	case BookContent:
		if c.Body == nil || *c.Body == "" {
			return ErrInvalidMedia
		}
	case PresentationContent:
		if c.FileURL == nil || *c.FileURL == "" {
			return ErrInvalidMedia
		}
	default:
		return ErrInvalidMedia
	}

	return nil
}

// CreateContentItemRequest carries the fields needed to create a new draft ContentItem.
type CreateContentItemRequest struct {
	Slug             string      `json:"slug"`
	Title            string      `json:"title"`
	ContentType      ContentType `json:"content_type"`
	Description      *string     `json:"description"`
	Body             *string     `json:"body"`
	ThumbnailURL     *string     `json:"thumbnail_url"`
	VideoURL         *string     `json:"video_url"`
	FileURL          *string     `json:"file_url"`
	EstimatedMinutes *int        `json:"estimated_minutes"`
	EstimatedPages   *int        `json:"estimated_pages"`
	SeoTitle         *string     `json:"seo_title"`
	SeoDescription   *string     `json:"seo_description"`
	OgImageURL       *string     `json:"og_image_url"`
	CanonicalURL     *string     `json:"canonical_url"`
	IsIndexable      *bool       `json:"is_indexable"`
	CreatedByUserID  string      `json:"-"`
}

// Validate checks that the create ContentItem request fields meet format requirements.
func (req *CreateContentItemRequest) Validate() error {
	checks := []func() error{
		req.validateSlug,
		req.validateTitle,
		req.validateDescription,
		req.validateEstimatedMinutes,
		req.validateEstimatedPages,
		req.validateThumbnailURL,
		req.validateSeoTitle,
		req.validateSeoDescription,
		req.validateOgImageURL,
		req.validateCanonicalURL,
		req.validateFileURL,
		req.validateVideoURL,
		req.validateBody,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *CreateContentItemRequest) validateSlug() error {
	if req.Slug == "" || !validator.IsValidSlug(req.Slug) {
		return ErrInvalidSlug
	}
	return nil
}

func (req *CreateContentItemRequest) validateTitle() error {
	if !validator.IsValidContentTitle(req.Title) {
		return ErrInvalidTitle
	}
	return nil
}

func (req *CreateContentItemRequest) validateDescription() error {
	return validator.RequireOptionalContentDescription(req.Description, ErrInvalidDescription)
}

func (req *CreateContentItemRequest) validateSeoTitle() error {
	return validator.RequireOptionalSeoTitle(req.SeoTitle, ErrInvalidSeoTitle)
}

func (req *CreateContentItemRequest) validateSeoDescription() error {
	return validator.RequireOptionalSeoDescription(req.SeoDescription, ErrInvalidSeoDescription)
}

func (req *CreateContentItemRequest) validateEstimatedMinutes() error {
	return validator.RequireOptionalPositiveInt(req.EstimatedMinutes, ErrInvalidEstimatedMinutes)
}

func (req *CreateContentItemRequest) validateEstimatedPages() error {
	return validator.RequireOptionalPositiveInt(req.EstimatedPages, ErrInvalidEstimatedPages)
}

func (req *CreateContentItemRequest) validateThumbnailURL() error {
	return validator.RequireOptionalHTTPSURL(req.ThumbnailURL, ErrInvalidThumbnailURL)
}

func (req *CreateContentItemRequest) validateOgImageURL() error {
	return validator.RequireOptionalHTTPSURL(req.OgImageURL, ErrInvalidOgImageURL)
}

func (req *CreateContentItemRequest) validateCanonicalURL() error {
	return validator.RequireOptionalHTTPSURL(req.CanonicalURL, ErrInvalidCanonicalURL)
}

func (req *CreateContentItemRequest) validateFileURL() error {
	return validator.RequireOptionalHTTPSURL(req.FileURL, ErrInvalidMedia)
}

func (req *CreateContentItemRequest) validateVideoURL() error {
	return validator.RequireOptionalHTTPSURL(req.VideoURL, ErrInvalidMedia)
}

func (req *CreateContentItemRequest) validateBody() error {
	return validator.RequireOptionalContentBody(req.Body, ErrInvalidMedia)
}

// UpdateContentItemRequest carries the fields to patch onto an existing ContentItem; nil fields
// are left unchanged.
type UpdateContentItemRequest struct {
	ID               string  `json:"id"`
	Slug             *string `json:"slug"`
	Title            *string `json:"title"`
	Description      *string `json:"description"`
	ThumbnailURL     *string `json:"thumbnail_url"`
	EstimatedMinutes *int    `json:"estimated_minutes"`
	EstimatedPages   *int    `json:"estimated_pages"`
	SeoTitle         *string `json:"seo_title"`
	SeoDescription   *string `json:"seo_description"`
	OgImageURL       *string `json:"og_image_url"`
	CanonicalURL     *string `json:"canonical_url"`
	IsIndexable      *bool   `json:"is_indexable"`
	Body             *string `json:"body"`
	VideoURL         *string `json:"video_url"`
	FileURL          *string `json:"file_url"`
}

// Validate checks that the update ContentItem request fields meet format requirements.
func (req *UpdateContentItemRequest) Validate() error {
	checks := []func() error{
		req.validateID,
		req.validateSlug,
		req.validateTitle,
		req.validateDescription,
		req.validateThumbnailURL,
		req.validateSeoTitle,
		req.validateSeoDescription,
		req.validateOgImageURL,
		req.validateCanonicalURL,
		req.validateEstimatedMinutes,
		req.validateEstimatedPages,
		req.validateFileURL,
		req.validateVideoURL,
		req.validateBody,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *UpdateContentItemRequest) validateID() error {
	if req.ID == "" || !validator.IsValidUUID(req.ID) {
		return ErrInvalidContentItemID
	}
	return nil
}

func (req *UpdateContentItemRequest) validateSlug() error {
	if req.Slug != nil && (*req.Slug == "" || !validator.IsValidSlug(*req.Slug)) {
		return ErrInvalidSlug
	}
	return nil
}

func (req *UpdateContentItemRequest) validateTitle() error {
	if req.Title != nil && !validator.IsValidContentTitle(*req.Title) {
		return ErrInvalidTitle
	}
	return nil
}

func (req *UpdateContentItemRequest) validateDescription() error {
	return validator.RequireOptionalContentDescription(req.Description, ErrInvalidDescription)
}

func (req *UpdateContentItemRequest) validateThumbnailURL() error {
	return validator.RequireOptionalHTTPSURL(req.ThumbnailURL, ErrInvalidThumbnailURL)
}

func (req *UpdateContentItemRequest) validateSeoTitle() error {
	return validator.RequireOptionalSeoTitle(req.SeoTitle, ErrInvalidSeoTitle)
}

func (req *UpdateContentItemRequest) validateSeoDescription() error {
	return validator.RequireOptionalSeoDescription(req.SeoDescription, ErrInvalidSeoDescription)
}

func (req *UpdateContentItemRequest) validateOgImageURL() error {
	return validator.RequireOptionalHTTPSURL(req.OgImageURL, ErrInvalidOgImageURL)
}

func (req *UpdateContentItemRequest) validateCanonicalURL() error {
	return validator.RequireOptionalHTTPSURL(req.CanonicalURL, ErrInvalidCanonicalURL)
}

func (req *UpdateContentItemRequest) validateEstimatedMinutes() error {
	return validator.RequireOptionalPositiveInt(req.EstimatedMinutes, ErrInvalidEstimatedMinutes)
}

func (req *UpdateContentItemRequest) validateEstimatedPages() error {
	return validator.RequireOptionalPositiveInt(req.EstimatedPages, ErrInvalidEstimatedPages)
}

func (req *UpdateContentItemRequest) validateFileURL() error {
	return validator.RequireOptionalHTTPSURL(req.FileURL, ErrInvalidMedia)
}

func (req *UpdateContentItemRequest) validateVideoURL() error {
	return validator.RequireOptionalHTTPSURL(req.VideoURL, ErrInvalidMedia)
}

func (req *UpdateContentItemRequest) validateBody() error {
	return validator.RequireOptionalContentBody(req.Body, ErrInvalidMedia)
}

// Apply copies every non-nil field from r onto p.
func (r UpdateContentItemRequest) Apply(p *ContentItem) {
	appliers := []func(*ContentItem){
		r.applySlug,
		r.applyTitle,
		r.applyDescription,
		r.applyThumbnailURL,
		r.applyEstimatedMinutes,
		r.applySeoTitle,
		r.applySeoDescription,
		r.applyOgImageURL,
		r.applyCanonicalURL,
		r.applyIsIndexable,
		r.applyBody,
		r.applyVideoURL,
		r.applyFileURL,
		r.applyEstimatedPages,
	}
	for _, apply := range appliers {
		apply(p)
	}
}

func (r UpdateContentItemRequest) applySlug(p *ContentItem) {
	if r.Slug != nil {
		p.Slug = *r.Slug
	}
}

func (r UpdateContentItemRequest) applyTitle(p *ContentItem) {
	if r.Title != nil {
		p.Title = *r.Title
	}
}

func (r UpdateContentItemRequest) applyDescription(p *ContentItem) {
	if r.Description != nil {
		p.Description = r.Description
	}
}

func (r UpdateContentItemRequest) applyThumbnailURL(p *ContentItem) {
	if r.ThumbnailURL != nil {
		p.ThumbnailURL = r.ThumbnailURL
	}
}

func (r UpdateContentItemRequest) applyEstimatedMinutes(p *ContentItem) {
	if r.EstimatedMinutes != nil {
		p.EstimatedMinutes = r.EstimatedMinutes
	}
}

func (r UpdateContentItemRequest) applyEstimatedPages(p *ContentItem) {
	if r.EstimatedPages != nil {
		p.EstimatedPages = r.EstimatedPages
	}
}

func (r UpdateContentItemRequest) applySeoTitle(p *ContentItem) {
	if r.SeoTitle != nil {
		p.SeoTitle = r.SeoTitle
	}
}

func (r UpdateContentItemRequest) applySeoDescription(p *ContentItem) {
	if r.SeoDescription != nil {
		p.SeoDescription = r.SeoDescription
	}
}

func (r UpdateContentItemRequest) applyOgImageURL(p *ContentItem) {
	if r.OgImageURL != nil {
		p.OgImageURL = r.OgImageURL
	}
}

func (r UpdateContentItemRequest) applyCanonicalURL(p *ContentItem) {
	if r.CanonicalURL != nil {
		p.CanonicalURL = r.CanonicalURL
	}
}

func (r UpdateContentItemRequest) applyIsIndexable(p *ContentItem) {
	if r.IsIndexable != nil {
		p.IsIndexable = *r.IsIndexable
	}
}

func (r UpdateContentItemRequest) applyBody(p *ContentItem) {
	if r.Body != nil {
		p.Body = r.Body
	}
}

func (r UpdateContentItemRequest) applyVideoURL(p *ContentItem) {
	if r.VideoURL != nil {
		p.VideoURL = r.VideoURL
	}
}

func (r UpdateContentItemRequest) applyFileURL(p *ContentItem) {
	if r.FileURL != nil {
		p.FileURL = r.FileURL
	}
}
