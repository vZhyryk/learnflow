package consultationshttp

import (
	"net/http"

	consultationsdomain "learnflow_backend/internal/consultations/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/validator"
)

// parseSlotQuery reads the from/to range and the optional consultant_id of a slot listing.
func parseSlotQuery(r *http.Request) (consultationsdomain.SlotRange, *string, error) {
	query := r.URL.Query()

	slotRange, err := consultationsdomain.ParseSlotRange(query.Get("from"), query.Get("to"))
	if err != nil {
		return consultationsdomain.SlotRange{}, nil, err
	}

	consultantID := query.Get("consultant_id")
	if consultantID == "" {
		return slotRange, nil, nil
	}

	if !validator.IsValidUUID(consultantID) {
		return consultationsdomain.SlotRange{}, nil, consultationsdomain.ErrInvalidConsultant
	}

	return slotRange, &consultantID, nil
}

func (h *Handler) getAvailableSlots(w http.ResponseWriter, r *http.Request) {
	slotRange, consultantID, err := parseSlotQuery(r)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	slots, total, err := h.svc.GetAvailableSlots(r.Context(), slotRange, consultantID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	h.writeJSON(w, r, helpers.Envelope{"slots": slots, "total": total})
}

func (h *Handler) getAllSlots(w http.ResponseWriter, r *http.Request) {
	slotRange, consultantID, err := parseSlotQuery(r)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	slots, total, err := h.svc.GetAllSlots(r.Context(), slotRange, consultantID)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	h.writeJSON(w, r, helpers.Envelope{"slots": slots, "total": total})
}

func (h *Handler) updateSlots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := appcontext.MustUserFromContext(ctx)

	var req consultationsdomain.UpdateSlotsRequest
	if !helpers.DecodeAndValidate(w, r, h.jsonLogger, &req, func() { req.AdminUserID = user.ID }) {
		return
	}

	result, err := h.svc.UpdateSlots(ctx, req)
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	h.writeJSON(w, r, helpers.Envelope{"result": result})
}

func (h *Handler) getConsultants(w http.ResponseWriter, r *http.Request) {
	consultants, err := h.svc.GetConsultants(r.Context())
	if err != nil {
		h.handleErrorResponse(w, r, err)
		return
	}

	h.writeJSON(w, r, helpers.Envelope{"consultants": consultants})
}

func (h *Handler) writeJSON(w http.ResponseWriter, r *http.Request, body helpers.Envelope) {
	if err := helpers.WriteJSON(w, http.StatusOK, body, nil); err != nil {
		h.jsonLogger.Error(err, map[string]any{"path": r.URL.Path})
	}
}
