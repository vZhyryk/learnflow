package adminhttp

import (
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/pagination"
	"net/http"
	"time"
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

	actions, count, err := h.svc.GetInstanceAdminActions(ctx, user.ID, req.TargetType, req.ItemID, pagination.ParsePaginationParams(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"actions": actions, "total": count}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

func (h *Handler) getAdminActions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	filter, err := parseAdminActionFilter(r)
	if err == nil {
		err = filter.Validate()
	}
	if err != nil {
		if respErr := helpers.BadRequestResponse(w, err); respErr != nil {
			h.jsonLogger.Error(respErr, map[string]any{"user_id": user.ID, "path": r.URL.Path})
		}
		return
	}

	actions, count, err := h.svc.GetAdminActions(ctx, user.ID, filter, pagination.ParsePaginationParams(r))
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"actions": actions, "total": count}, nil)
	if err != nil {
		h.jsonLogger.Error(err, map[string]any{"user_id": user.ID, "path": r.URL.Path})
	}
}

// parseAdminActionFilter reads admin_user_id, action_type and the RFC 3339 from/to query parameters.
func parseAdminActionFilter(r *http.Request) (auditdomain.AdminActionFilter, error) {
	query := r.URL.Query()
	filter := auditdomain.AdminActionFilter{
		AdminUserID: query.Get("admin_user_id"),
		ActionType:  auditdomain.AdminActionType(query.Get("action_type")),
	}

	var err error
	if filter.From, err = parseOptionalTime(query.Get("from")); err != nil {
		return filter, err
	}
	filter.To, err = parseOptionalTime(query.Get("to"))

	return filter, err
}

func parseOptionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, auditdomain.ErrInvalidDate
	}

	return &parsed, nil
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
