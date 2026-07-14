package authhttp

import (
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ua := r.UserAgent()

	var req authdomain.RefreshRequest
	if !h.decodeAndValidate(w, r, &req, func() {
		req.IPAddress = appcontext.IPAddressFromContext(r.Context())
		req.UserAgent = ua
	}) {
		return
	}

	token, err := h.svc.Refresh(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"auth": token}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"event": refreshEvent, "user_id": token.UserID, "path": r.URL.Path})
	}

	h.logAuthEvent(r, refreshEvent, map[string]any{"user_id": token.UserID})
}
