package adminservice

import (
	admindomain "learnflow_backend/internal/admin/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
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
	blocklist   admindomain.UserBlockList
	operations  map[admindomain.UserAdminOperation]userOperation
}

var _ admindomain.Service = (*Service)(nil)

// New returns a new Service wired to its repositories, transactor, outbox and blocklist.
func New(
	announRepo admindomain.AnnouncementRepository,
	userRepo admindomain.UserRepository,
	actionRepo admindomain.AdminActionRepository,
	sessionRepo admindomain.SessionRepository,
	transactor admindomain.Transactor,
	outbox *events.OutboxWriter,
	blocklist admindomain.UserBlockList,
) *Service {
	srv := &Service{announRepo: announRepo, userRepo: userRepo, actionRepo: actionRepo, sessionRepo: sessionRepo, transactor: transactor, outbox: outbox, blocklist: blocklist}
	srv.operations = map[admindomain.UserAdminOperation]userOperation{
		admindomain.RevokeUserRole: {srv.userRepo.RevokeUserRole, auditdomain.ActionRevokeSubadmin},
		admindomain.AssignUserRole: {srv.userRepo.AssignUserRole, auditdomain.ActionAssignSubadmin},
		admindomain.DeleteUser:     {srv.userRepo.DeleteUser, auditdomain.ActionDeleteUser},
		admindomain.RestoreUser:    {srv.userRepo.RestoreUser, auditdomain.ActionRestoreUser},
		admindomain.BlockUser:      {srv.userRepo.BlockUser, auditdomain.ActionBlockUser},
		admindomain.UnBlockUser:    {srv.userRepo.UnBlockUser, auditdomain.ActionUnblockUser},
	}

	return srv
}
