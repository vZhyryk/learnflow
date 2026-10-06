package adminservice

import (
	admindomain "learnflow_backend/internal/admin/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/events"
)

// Service implements admindomain.Service.
type Service struct {
	announRepo      admindomain.AnnouncementRepository
	userRepo        admindomain.UserRepository
	actionRepo      admindomain.AdminActionRepository
	sessionRepo     admindomain.SessionRepository
	courseRepo      admindomain.CourseRepository
	contentItemRepo admindomain.ContentItemRepository
	transactor      admindomain.Transactor
	outbox          *events.OutboxWriter
	blocklist       admindomain.UserBlockList
	operations      map[admindomain.UserAdminOperation]userOperation
}

// Repos groups the repositories the admin Service depends on.
type Repos struct {
	AnnounRepo      admindomain.AnnouncementRepository
	UserRepo        admindomain.UserRepository
	ActionRepo      admindomain.AdminActionRepository
	SessionRepo     admindomain.SessionRepository
	CourseRepo      admindomain.CourseRepository
	ContentItemRepo admindomain.ContentItemRepository
}

// Utils groups the non-repository dependencies of the admin Service.
type Utils struct {
	Transactor admindomain.Transactor
	Outbox     *events.OutboxWriter
	Blocklist  admindomain.UserBlockList
}

var _ admindomain.Service = (*Service)(nil)

// New returns a new Service wired to its repositories, transactor, outbox and blocklist.
func New(
	repos Repos,
	utils Utils,
) *Service {
	srv := &Service{
		announRepo:      repos.AnnounRepo,
		userRepo:        repos.UserRepo,
		actionRepo:      repos.ActionRepo,
		sessionRepo:     repos.SessionRepo,
		courseRepo:      repos.CourseRepo,
		contentItemRepo: repos.ContentItemRepo,
		transactor:      utils.Transactor,
		outbox:          utils.Outbox,
		blocklist:       utils.Blocklist,
	}
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
