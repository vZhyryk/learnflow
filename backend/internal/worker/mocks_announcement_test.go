package worker

import (
	"context"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"

	"github.com/redis/go-redis/v9"
)

// mockAnnouncementRepo implements admindomain.AnnouncementRepository via function
// fields — only GetAnnouncementByID is exercised by AnnouncementFanOutWorker, the rest
// panic if ever called from a test that didn't set them up.
type mockAnnouncementRepo struct {
	createAnnouncement         func(ctx context.Context, a *admindomain.Announcement) (*admindomain.Announcement, error)
	updateAnnouncement         func(ctx context.Context, a *admindomain.Announcement) error
	approveAnnouncement        func(ctx context.Context, announcementID, userID string) error
	getAnnouncements           func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getUnApprovedAnnouncements func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getApprovedAnnouncements   func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getExpiredAnnouncements    func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getAnnouncementByID        func(ctx context.Context, id string) (*admindomain.Announcement, error)
	getPublicAnnouncements     func(ctx context.Context, params pagination.Params, userID string) ([]*admindomain.AnnouncementPublic, error)
}

func (m *mockAnnouncementRepo) CreateAnnouncement(ctx context.Context, a *admindomain.Announcement) (*admindomain.Announcement, error) {
	if m.createAnnouncement == nil {
		panic("mockAnnouncementRepo.CreateAnnouncement not set")
	}
	return m.createAnnouncement(ctx, a)
}

func (m *mockAnnouncementRepo) UpdateAnnouncement(ctx context.Context, a *admindomain.Announcement) error {
	if m.updateAnnouncement == nil {
		panic("mockAnnouncementRepo.UpdateAnnouncement not set")
	}
	return m.updateAnnouncement(ctx, a)
}

func (m *mockAnnouncementRepo) ApproveAnnouncement(ctx context.Context, announcementID, userID string) error {
	if m.approveAnnouncement == nil {
		panic("mockAnnouncementRepo.ApproveAnnouncement not set")
	}
	return m.approveAnnouncement(ctx, announcementID, userID)
}

func (m *mockAnnouncementRepo) GetAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	if m.getAnnouncements == nil {
		panic("mockAnnouncementRepo.GetAnnouncements not set")
	}
	return m.getAnnouncements(ctx, params)
}

func (m *mockAnnouncementRepo) GetUnApprovedAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	if m.getUnApprovedAnnouncements == nil {
		panic("mockAnnouncementRepo.GetUnApprovedAnnouncements not set")
	}
	return m.getUnApprovedAnnouncements(ctx, params)
}

func (m *mockAnnouncementRepo) GetApprovedAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	if m.getApprovedAnnouncements == nil {
		panic("mockAnnouncementRepo.GetApprovedAnnouncements not set")
	}
	return m.getApprovedAnnouncements(ctx, params)
}

func (m *mockAnnouncementRepo) GetPublicAnnouncements(ctx context.Context, params pagination.Params, userID string) ([]*admindomain.AnnouncementPublic, error) {
	if m.getPublicAnnouncements == nil {
		panic("mockAnnouncementRepo.GetPublicAnnouncements not set")
	}
	return m.getPublicAnnouncements(ctx, params, userID)
}

func (m *mockAnnouncementRepo) GetExpiredAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	if m.getExpiredAnnouncements == nil {
		panic("mockAnnouncementRepo.GetExpiredAnnouncements not set")
	}
	return m.getExpiredAnnouncements(ctx, params)
}

func (m *mockAnnouncementRepo) GetAnnouncementByID(ctx context.Context, id string) (*admindomain.Announcement, error) {
	if m.getAnnouncementByID == nil {
		panic("mockAnnouncementRepo.GetAnnouncementByID not set")
	}
	return m.getAnnouncementByID(ctx, id)
}

// newTestAnnouncementFanOutWorker builds an AnnouncementFanOutWorker directly (bypassing
// NewAnnouncementFanOutWorker) so tests can inject checkExists as a plain func instead of
// full ContentRepository/CourseRepository/ArticleRepository mocks.
func newTestAnnouncementFanOutWorker(runner db.QueryRunner, redisClient *redis.Client, annRep admindomain.AnnouncementRepository, entityConfigs map[admindomain.EntityType]entityConfig) *AnnouncementFanOutWorker {
	testLogger := testutil.NewTestLogger()
	return &AnnouncementFanOutWorker{
		db:              runner,
		redisClient:     redisClient,
		logger:          testLogger,
		eventType:       string(events.EventAnnouncementApprove),
		aggregationType: string(events.AggregationTypeAnnouncement),
		dlq:             NewDLQ(runner, testLogger),
		announcementRep: annRep,
		entityConfigs:   entityConfigs,
	}
}

func defaultTestEntityConfigs() map[admindomain.EntityType]entityConfig {
	return map[admindomain.EntityType]entityConfig{
		admindomain.ArticleEntityType: {
			checkExists:  func(_ context.Context, _ string) (bool, error) { return true, nil },
			recipientSQL: getAnnouncementUserListSQL,
		},
		admindomain.ContentEntityType: {
			checkExists:      func(_ context.Context, _ string) (bool, error) { return true, nil },
			recipientSQL:     getContentAnnouncementUserListSQL,
			takesEntityIDArg: true,
		},
		admindomain.CourseEntityType: {
			checkExists:      func(_ context.Context, _ string) (bool, error) { return true, nil },
			recipientSQL:     getCourseAnnouncementUserListSQL,
			takesEntityIDArg: true,
		},
	}
}
