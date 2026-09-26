package adminservice

import (
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/events"
)

// Service implements admindomain.Service.
type Service struct {
	announRepo  admindomain.AnnouncementRepository
	userRepo    admindomain.UserRepository
	actionRepo  admindomain.AdminActionRepository
	sessionRepo admindomain.SessionRepository
	transactor  admindomain.Transactor
	outbox      *events.OutboxWriter
	blocklist   admindomain.UserBlocklist
}

var _ admindomain.Service = (*Service)(nil)

// New returns a new Service wired to the given repository.
func New(
	announRepo admindomain.AnnouncementRepository,
	userRepo admindomain.UserRepository,
	actionRepo admindomain.AdminActionRepository,
	sessionRepo admindomain.SessionRepository,
	transactor admindomain.Transactor,
	outbox *events.OutboxWriter,
	blocklist admindomain.UserBlocklist,
) *Service {
	return &Service{announRepo: announRepo, userRepo: userRepo, actionRepo: actionRepo, sessionRepo: sessionRepo, transactor: transactor, outbox: outbox, blocklist: blocklist}
}
