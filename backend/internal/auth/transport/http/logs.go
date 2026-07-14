package authhttp

import (
	"fmt"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) logAuthEvent(r *http.Request, event string, props map[string]any) {
	if props == nil {
		props = make(map[string]any)
	}
	props["ip"] = appcontext.IPAddressFromContext(r.Context())
	props["user_agent"] = r.UserAgent()
	h.jsonLogger.Info(event, props)
}

func (h *Handler) logAuthFailure(r *http.Request, event, reason string, props map[string]any) {
	if props == nil {
		props = make(map[string]any)
	}
	props["ip"] = appcontext.IPAddressFromContext(r.Context())
	props["user_agent"] = r.UserAgent()
	props["request_id"] = appcontext.RequestIDFromContext(r.Context())
	props["reason"] = reason
	h.jsonLogger.Error(fmt.Errorf("%s failure", event), props)
}

func userIDProps(r *http.Request) map[string]any {
	if user, ok := appcontext.UserFromContext(r.Context()); ok {
		return map[string]any{"user_id": user.ID}
	}
	return nil
}
