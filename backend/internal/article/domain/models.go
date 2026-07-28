package articledomain

import (
	"time"

	"learnflow_backend/internal/shared/validator"
)

// ArticleStatus represents the current lifecycle state of a Article.
type ArticleStatus string

// ArticleStatus values.
const (
	DraftStatus     ArticleStatus = "draft"
	PublishedStatus ArticleStatus = "published"
	ArchivedStatus  ArticleStatus = "archived"
)

// Valid reports whether r is one of the known ArticleStatus values.
func (r ArticleStatus) Valid() bool {
	switch r {
	case
		DraftStatus,
		PublishedStatus,
		ArchivedStatus:
		return true
	}
	return false
}

// Article represents a Article in draft, published, or archived state.
type Article struct {
	ID              string        `json:"id"`
	Slug            string        `json:"slug"`
	Title           string        `json:"title"`
	Excerpt         *string       `json:"excerpt"`
	Body            string        `json:"body"`
	SeoTitle        *string       `json:"seo_title"`
	SeoDescription  *string       `json:"seo_description"`
	OgImageURL      *string       `json:"og_image_url"`
	IsIndexable     bool          `json:"is_indexable"`
	Status          ArticleStatus `json:"status"`
	CreatedByUserID string        `json:"created_by_user_id"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	PublishedAt     *time.Time    `json:"published_at"`
	DeletedAt       *time.Time    `json:"deleted_at"`
}

// ReadyToPublish reports whether the Article has all fields required to go public.
func (c *Article) ReadyToPublish() error {
	checks := []func() error{
		c.checkTitleReady,
		c.checkBody,
		c.checkExcerpt,
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

func (c *Article) checkTitleReady() error {
	if c.Title == "" {
		return ErrInvalidTitle
	}
	return nil
}

func (c *Article) checkSeoTitleReady() error {
	if c.SeoTitle == nil || *c.SeoTitle == "" {
		return ErrInvalidSeoTitle
	}
	return nil
}

func (c *Article) checkSeoDescriptionReady() error {
	if c.SeoDescription == nil || *c.SeoDescription == "" {
		return ErrInvalidSeoDescription
	}
	return nil
}

func (c *Article) checkBody() error {
	if c.Body == "" {
		return ErrInvalidBody
	}

	return nil
}

func (c *Article) checkExcerpt() error {
	if c.Excerpt == nil || *c.Excerpt == "" {
		return ErrInvalidExcerpt
	}

	return nil
}

// CreateArticleRequest carries the fields needed to create a new draft Article.
type CreateArticleRequest struct {
	Slug            string  `json:"slug"`
	Title           string  `json:"title"`
	Excerpt         *string `json:"excerpt"`
	Body            string  `json:"body"`
	SeoTitle        *string `json:"seo_title"`
	SeoDescription  *string `json:"seo_description"`
	OgImageURL      *string `json:"og_image_url"`
	IsIndexable     *bool   `json:"is_indexable"`
	CreatedByUserID string  `json:"-"`
}

// Validate checks that the create Article request fields meet format requirements.
func (req *CreateArticleRequest) Validate() error {
	checks := []func() error{
		req.validateSlug,
		req.validateTitle,
		req.validateSeoTitle,
		req.validateSeoDescription,
		req.validateOgImageURL,
		req.validateBody,
		req.validateExcerpt,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *CreateArticleRequest) validateExcerpt() error {
	if req.Excerpt != nil && *req.Excerpt == "" {
		return ErrInvalidExcerpt
	}
	return nil
}

func (req *CreateArticleRequest) validateSlug() error {
	if req.Slug == "" || !validator.IsValidSlug(req.Slug) {
		return ErrInvalidSlug
	}
	return nil
}

func (req *CreateArticleRequest) validateTitle() error {
	if !validator.IsValidContentTitle(req.Title) {
		return ErrInvalidTitle
	}
	return nil
}

func (req *CreateArticleRequest) validateSeoTitle() error {
	return validator.RequireOptionalSeoTitle(req.SeoTitle, ErrInvalidSeoTitle)
}

func (req *CreateArticleRequest) validateSeoDescription() error {
	return validator.RequireOptionalSeoDescription(req.SeoDescription, ErrInvalidSeoDescription)
}

func (req *CreateArticleRequest) validateOgImageURL() error {
	return validator.RequireOptionalHTTPSURL(req.OgImageURL, ErrInvalidOgImageURL)
}

func (req *CreateArticleRequest) validateBody() error {
	if req.Body == "" || !validator.IsValidContentBody(req.Body) {
		return ErrInvalidBody
	}
	return nil
}

// UpdateArticleRequest carries the fields to patch onto an existing Article; nil fields
// are left unchanged.
type UpdateArticleRequest struct {
	ID             string  `json:"id"`
	Slug           *string `json:"slug"`
	Title          *string `json:"title"`
	SeoTitle       *string `json:"seo_title"`
	SeoDescription *string `json:"seo_description"`
	OgImageURL     *string `json:"og_image_url"`
	IsIndexable    *bool   `json:"is_indexable"`
	Body           *string `json:"body"`
	Excerpt        *string `json:"excerpt"`
}

// Validate checks that the update Article request fields meet format requirements.
func (req *UpdateArticleRequest) Validate() error {
	checks := []func() error{
		req.validateID,
		req.validateSlug,
		req.validateTitle,
		req.validateSeoTitle,
		req.validateSeoDescription,
		req.validateOgImageURL,
		req.validateBody,
		req.validateExcerpt,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *UpdateArticleRequest) validateID() error {
	if req.ID == "" || !validator.IsValidUUID(req.ID) {
		return ErrInvalidArticleID
	}
	return nil
}

func (req *UpdateArticleRequest) validateSlug() error {
	if req.Slug != nil && (*req.Slug == "" || !validator.IsValidSlug(*req.Slug)) {
		return ErrInvalidSlug
	}
	return nil
}

func (req *UpdateArticleRequest) validateExcerpt() error {
	if req.Excerpt != nil && *req.Excerpt == "" {
		return ErrInvalidExcerpt
	}
	return nil
}

func (req *UpdateArticleRequest) validateTitle() error {
	if req.Title != nil && !validator.IsValidContentTitle(*req.Title) {
		return ErrInvalidTitle
	}
	return nil
}

func (req *UpdateArticleRequest) validateSeoTitle() error {
	return validator.RequireOptionalSeoTitle(req.SeoTitle, ErrInvalidSeoTitle)
}

func (req *UpdateArticleRequest) validateSeoDescription() error {
	return validator.RequireOptionalSeoDescription(req.SeoDescription, ErrInvalidSeoDescription)
}

func (req *UpdateArticleRequest) validateOgImageURL() error {
	return validator.RequireOptionalHTTPSURL(req.OgImageURL, ErrInvalidOgImageURL)
}

func (req *UpdateArticleRequest) validateBody() error {
	return validator.RequireOptionalContentBody(req.Body, ErrInvalidBody)
}

// Apply copies every non-nil field from r onto p.
func (r UpdateArticleRequest) Apply(p *Article) {
	appliers := []func(*Article){
		r.applySlug,
		r.applyTitle,
		r.applySeoTitle,
		r.applySeoDescription,
		r.applyOgImageURL,
		r.applyIsIndexable,
		r.applyBody,
		r.applyExcerpt,
	}
	for _, apply := range appliers {
		apply(p)
	}
}

func (r UpdateArticleRequest) applySlug(p *Article) {
	if r.Slug != nil {
		p.Slug = *r.Slug
	}
}

func (r UpdateArticleRequest) applyTitle(p *Article) {
	if r.Title != nil {
		p.Title = *r.Title
	}
}

func (r UpdateArticleRequest) applySeoTitle(p *Article) {
	if r.SeoTitle != nil {
		p.SeoTitle = r.SeoTitle
	}
}

func (r UpdateArticleRequest) applySeoDescription(p *Article) {
	if r.SeoDescription != nil {
		p.SeoDescription = r.SeoDescription
	}
}

func (r UpdateArticleRequest) applyOgImageURL(p *Article) {
	if r.OgImageURL != nil {
		p.OgImageURL = r.OgImageURL
	}
}

func (r UpdateArticleRequest) applyIsIndexable(p *Article) {
	if r.IsIndexable != nil {
		p.IsIndexable = *r.IsIndexable
	}
}

func (r UpdateArticleRequest) applyBody(p *Article) {
	if r.Body != nil {
		p.Body = *r.Body
	}
}

func (r UpdateArticleRequest) applyExcerpt(p *Article) {
	if r.Excerpt != nil {
		p.Excerpt = r.Excerpt
	}
}
