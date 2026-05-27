package router

import (
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
	"strconv"

	"golang.org/x/time/rate"
)

// RateLimitExceededResponse writes a 429 response with Retry-After header.
func (route *RouteHandler) RateLimitExceededResponse(w http.ResponseWriter, _ *http.Request, limiter *rate.Limiter) {
	reservation := limiter.Reserve()
	delay := reservation.Delay()
	reservation.Cancel()

	retryAfter := max(int(delay.Seconds())+1, 1)

	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	route.ErrorResponse(w, http.StatusTooManyRequests, "rate limit exceeded")
}

// NotFoundResponse writes a 404 response.
func (route *RouteHandler) NotFoundResponse(w http.ResponseWriter, _ *http.Request) {
	route.ErrorResponse(w, http.StatusNotFound, "the requested resource could not be found")
}

// ErrorResponse writes a JSON error envelope with the given status code and message.
func (route *RouteHandler) ErrorResponse(w http.ResponseWriter, status int, message string) {
	env := helpers.Envelope{
		"error": message,
	}

	err := helpers.WriteJSON(w, status, env, nil)
	if err != nil {
		route.App.Logger.Error(err, nil)
	}
}
