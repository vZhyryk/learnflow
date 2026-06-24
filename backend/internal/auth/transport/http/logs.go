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

func (h *Handler) logAuthFailure(r *http.Request, event string, reason string, props map[string]any) {
	if props == nil {
		props = make(map[string]any)
	}
	props["ip"] = appcontext.IPAddressFromContext(r.Context())
	props["reason"] = reason
	h.jsonLogger.Error(fmt.Errorf("%s failure", event), props)
}
