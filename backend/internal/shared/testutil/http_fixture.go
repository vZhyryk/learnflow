package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

// HTTPFixture wires a mux and a request builder for a single route, shared by every
// per-handler fixture across the *_test.go packages that need one (loginFixture,
// registerFixture, ...). Embed it and add the handler-specific svcResult/svcErr fields
// on top.
type HTTPFixture struct {
	Mux    *http.ServeMux
	NewReq func(body string, urlParams map[string]string) *http.Request
}

// NewHTTPFixture returns an HTTPFixture for method/path against the given mux.
func NewHTTPFixture(mux *http.ServeMux, method, path string) *HTTPFixture {
	return &HTTPFixture{
		Mux: mux,
		NewReq: func(body string, urlParams map[string]string) *http.Request {
			query := url.Values{}
			for key, value := range urlParams {
				if value != "" && key != "" {
					query.Set(key, value)
				}
			}

			reqPath := path
			if len(query) > 0 {
				reqPath += "?" + query.Encode()
			}

			return httptest.NewRequestWithContext(context.Background(), method, reqPath, strings.NewReader(body))
		},
	}
}
