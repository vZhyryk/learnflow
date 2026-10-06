package adminhttp

import (
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/pagination"
	"net/http"
)

func (h *Handler) getInstanceAdminActions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	req := auditdomain.GetInstanceAdminActionsRequest{
		ItemID:     query.Get("target_id"),
		TargetType: auditdomain.AdminTargetType(query.Get("target_type")),
	}

	user := appcontext.MustUserFromContext(ctx)
	if err := req.Validate(); err != nil {
		if respErr := helpers.BadRequestResponse(w, err); respErr != nil {
			h.jsonLogger.Error(respErr, map[string]any{"user_id": user.ID, "path": r.URL.Path})
		}
		return
	}

	actions, count, err := h.svc.GetInstanceAdminActions(ctx, req.TargetType, req.ItemID, pagination.ParsePaginationParams(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"actions": actions, "total": count}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) getFailedJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)
	failedJobs, count, err := h.svc.GetFailedJobs(ctx, pagination.ParsePaginationParams(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"failed_jobs": failedJobs, "total": count}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}
