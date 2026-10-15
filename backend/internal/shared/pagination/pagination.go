package pagination

import (
	"net/http"
	"strconv"
)

// Params carries normalized, 1-indexed page/page-size values for a LIMIT/OFFSET query.
type Params struct {
	Page     int
	PageSize int
}

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// NewParams normalizes page/pageSize query values: page < 1 becomes 1, pageSize <= 0
// becomes defaultPageSize, and pageSize > maxPageSize is capped at maxPageSize.
func NewParams(page, pageSize int) Params {
	if page < 1 {
		page = 1
	}
	switch {
	case pageSize <= 0:
		pageSize = defaultPageSize
	case pageSize > maxPageSize:
		pageSize = maxPageSize
	}
	return Params{Page: page, PageSize: pageSize}
}

// Limit returns the SQL LIMIT value.
func (p Params) Limit() int {
	return p.PageSize
}

// Offset returns the SQL OFFSET value.
func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// ParsePaginationParams reads ?page=/?page_size=; invalid or missing values fall back to
// NewParams' defaults instead of 400ing.
func ParsePaginationParams(r *http.Request) Params {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 0
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil {
		pageSize = 0
	}
	return NewParams(page, pageSize)
}
