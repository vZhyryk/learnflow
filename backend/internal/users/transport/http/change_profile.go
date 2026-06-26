package usershttp

import (
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	usersdomain "learnflow_backend/internal/users/domain"
	"net/http"
)

func (h *Handler) changeProfile(w http.ResponseWriter, r *http.Request) {
	var req usersdomain.ChangeUserProfileRequest
	if !h.decodeAndValidate(w, r, &req, nil) {
		return
	}

	ctx := r.Context()
	user, ok := appcontext.UserFromContext(ctx)
	if !ok {
		if err := helpers.InvalidCredentialsResponse(w); err != nil {
			h.jsonLogger.Error(err, map[string]any{"case": "invalid_credentials", "path": r.URL.Path})
		}

		return
	}
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
