package consultationsdomain

import (
	"context"
	auditdomain "learnflow_backend/internal/audit/domain"
	"time"
)

// Transactor executes a function within a database transaction.
type Transactor interface {
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// AdminActionRepository persists the admin audit trail.
type AdminActionRepository interface {
	CreateAdminAction(ctx context.Context, action *auditdomain.AdminAction) error
}

// SlotRepository persists consultation slots.
type SlotRepository interface {
	// GetAvailableSlots and GetAllSlots list the range; a non-nil consultantID narrows it to one consultant.
	GetAvailableSlots(ctx context.Context, from, to time.Time, consultantID *string) ([]*PublicSlot, int, error)
	GetAllSlots(ctx context.Context, from, to time.Time, consultantID *string) ([]*Slot, int, error)
	UpdateSlotsStatus(ctx context.Context, ids []string, status SlotStatus, updatedBy string) error
	GetSlotsByIDsForUpdate(ctx context.Context, ids []string) ([]*Slot, error)
	// CreateSlotsIfMissing inserts one slot per (dates[i], starts[i]) pair for the consultant and skips existing ones.
	CreateSlotsIfMissing(ctx context.Context, consultantID string, dates []time.Time, starts []string, durationMinutes int) error
	// GetConsultants returns the active admins, who conduct consultations.
	GetConsultants(ctx context.Context) ([]*Consultant, error)
}

// Service is the consultations use-case layer.
type Service interface {
	GetAvailableSlots(ctx context.Context, slotRange SlotRange, consultantID *string) ([]*PublicSlot, int, error)
	GetAllSlots(ctx context.Context, slotRange SlotRange, consultantID *string) ([]*Slot, int, error)
	UpdateSlots(ctx context.Context, req UpdateSlotsRequest) (*UpdateSlotsResult, error)
	GenerateSlots(ctx context.Context) error
	GetConsultants(ctx context.Context) ([]*Consultant, error)
}
