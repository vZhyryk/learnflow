package router

import (
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
	"strconv"

	"golang.org/x/time/rate"
)

// RateLimitExceededResponse writes a 429 response with Retry-After header.
func (route *RouteHandler) RateLimitExceededResponse(w http.ResponseWriter, limiter *rate.Limiter) {
	reservation := limiter.Reserve()
	delay := reservation.Delay()
	reservation.Cancel()

	retryAfter := max(int(delay.Seconds())+1, 1)

	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	if err := helpers.ErrorResponse(w, http.StatusTooManyRequests, "rate limit exceeded"); err != nil {
		route.App.Logger.Error(err, nil)
	}
}
