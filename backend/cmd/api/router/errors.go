package router

import (
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
)

// RateLimitExceededResponse writes a 429 response with Retry-After header.
func (route *RouteHandler) RateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Retry-After", "60")
	if err := helpers.ErrorResponse(w, http.StatusTooManyRequests, "rate limit exceeded"); err != nil {
		route.App.Logger.Error(err, map[string]any{
			"method": r.Method,
			"path":   r.URL.Path,
		})
	}
}
