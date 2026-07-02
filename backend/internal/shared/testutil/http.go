package testutil

import (
	"net/http"
	"net/http/httptest"
)

// ServeHTTP fires req through mux and returns the recorded response, saving
// handler tests the boilerplate of constructing an httptest.ResponseRecorder.
func ServeHTTP(mux *http.ServeMux, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}
