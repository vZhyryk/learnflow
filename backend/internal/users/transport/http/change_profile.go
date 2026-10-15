package usershttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	usersdomain "learnflow_backend/internal/users/domain"
	"net/http"
)

func (h *Handler) changeProfile(w http.ResponseWriter, r *http.Request) {
	var req usersdomain.ChangeUserProfileRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, nil) {
		return
	}

	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	req.UserID = &user.ID

	err := h.svc.ChangeUserProfile(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "user profile was successfully updated"}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
