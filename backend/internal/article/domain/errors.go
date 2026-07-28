package articledomain

import "errors"

// Domain error sentinels for the Articles module — mapped to HTTP status codes in
// transport/http/errors.go.
var (
	ErrArticleNotFound = errors.New("Article not found")

	ErrInvalidSlug           = errors.New("invalid slug")
	ErrInvalidTitle          = errors.New("invalid title")
	ErrInvalidSeoTitle       = errors.New("invalid seo title")
	ErrInvalidSeoDescription = errors.New("invalid seo description")
	ErrInvalidOgImageURL     = errors.New("invalid og image url")
	ErrInvalidBody           = errors.New("invalid body")
	ErrInvalidExcerpt        = errors.New("invalid Excerpt")

	ErrInvalidArticleID = errors.New("invalid Article ID")

	ErrInvalidArticleStatus = errors.New("invalid Article status for this operation")

	ErrInvalidGetType = errors.New("invalid get type")
)
