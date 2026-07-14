package testutil

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authdomain "learnflow_backend/internal/auth/domain"
	appcontext "learnflow_backend/internal/shared/context"
)

// ServeHTTP fires req through mux and returns the recorded response, saving
// handler tests the boilerplate of constructing an httptest.ResponseRecorder.
func ServeHTTP(mux *http.ServeMux, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

// ErrWriter is an http.ResponseWriter whose Write always fails
type ErrWriter struct {
	httptest.ResponseRecorder
}

func (e *ErrWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

// WithUser attaches a fixed test user to req's context.
func WithUser(r *http.Request) *http.Request {
	user := &authdomain.User{ID: "user-123"}
	return r.WithContext(appcontext.WithUser(r.Context(), user))
}

// DecodeBody unmarshals a JSON response body into a map, failing the test on error.
func DecodeBody(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return m
}

// OkHandler returns a handler that always responds 200 OK, for tests that only need a pass-through next handler.
func OkHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
