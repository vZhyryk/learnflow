package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	admindomain "learnflow_backend/internal/admin/domain"
	articledomain "learnflow_backend/internal/article/domain"
	contentdomain "learnflow_backend/internal/content/domain"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/infrastructure/retry"
	"slices"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// recipientBaseSelectSQL is the shared SELECT/FROM/JOIN shape every recipient query
	// builds on — only the entity-access JOIN and the WHERE clause differ between them.
	recipientBaseSelectSQL = `
		SELECT u.id FROM users u
		LEFT JOIN notification_preferences np ON np.user_id = u.id
	`

	getAnnouncementUserListSQL = recipientBaseSelectSQL + `
		WHERE np.email_on_announcement = true AND u.deleted_at IS NULL AND u.status = 'active'
	`

	getContentAnnouncementUserListSQL = recipientBaseSelectSQL + `
		LEFT JOIN user_content_access uca ON uca.user_id = u.id
		WHERE np.email_on_announcement = true AND uca.content_item_id = $1 AND uca.status = 'active'
		AND (uca.expires_at IS NULL OR uca.expires_at > now()) AND u.deleted_at IS NULL AND u.status = 'active'

		UNION

	` + recipientBaseSelectSQL + `
		LEFT JOIN user_course_access uca ON uca.user_id = u.id
		JOIN course_content_items cci ON cci.course_id = uca.course_id
		WHERE np.email_on_announcement = true AND cci.content_item_id = $1 AND uca.status = 'active'
		AND (uca.expires_at IS NULL OR uca.expires_at > now()) AND u.deleted_at IS NULL AND u.status = 'active'
	`

	getCourseAnnouncementUserListSQL = recipientBaseSelectSQL + `
		LEFT JOIN user_course_access uca ON uca.user_id = u.id
		WHERE np.email_on_announcement = true AND uca.course_id = $1 AND uca.status = 'active'
		AND (uca.expires_at IS NULL OR uca.expires_at > now()) AND u.deleted_at IS NULL AND u.status = 'active'
	`

	insertAnnouncementEmailDeliverQuery = `
		INSERT INTO announcement_email_deliveries (user_id, announcement_id)
		VALUES {setter}
		ON CONFLICT (user_id, announcement_id) DO NOTHING
	`
)

// entityConfig bundles, per EntityType, everything both buildRecipientQuery and
// checkEntityExists need — one lookup instead of two parallel switches over the same enum.
type entityConfig struct {
	checkExists      func(ctx context.Context, id string) (bool, error)
	recipientSQL     string
	takesEntityIDArg bool
}

// AnnouncementFanOutWorker is Worker A of the announcement notification flow — it BLPops
// "announcement.approved", resolves the recipient list for the announcement's EntityType,
// and bulk-inserts announcement_email_deliveries rows (status='pending'). See
// NOTIFICATION_FLOW.md for the full flow.
type AnnouncementFanOutWorker struct {
	db              db.QueryRunner
	redisClient     *redis.Client
	logger          *logger.Logger
	eventType       string
	aggregationType string
	dlq             *DLQWriter
	announcementRep admindomain.AnnouncementRepository
	entityConfigs   map[admindomain.EntityType]entityConfig
}

// NewAnnouncementFanOutWorker returns an AnnouncementFanOutWorker wired with the given
// dependencies, ready to Run.
func NewAnnouncementFanOutWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	annRep admindomain.AnnouncementRepository,
	contentRep contentdomain.ContentRepository,
	courseRep coursedomain.CourseRepository,
	articleRep articledomain.ArticleRepository,
) *AnnouncementFanOutWorker {
	return &AnnouncementFanOutWorker{
		db:              queryRunner,
		redisClient:     redisClient,
		logger:          jsonLogger,
		eventType:       string(events.EventAnnouncementApprove),
		aggregationType: string(events.AggregationTypeAnnouncement),
		dlq:             NewDLQ(queryRunner, jsonLogger),
		announcementRep: annRep,
		entityConfigs: map[admindomain.EntityType]entityConfig{
			admindomain.ArticleEntityType: {
				checkExists:  articleRep.CheckIfArticleExistsByID,
				recipientSQL: getAnnouncementUserListSQL,
			},
			admindomain.ContentEntityType: {
				checkExists:      contentRep.CheckIfContentItemExistsByID,
				recipientSQL:     getContentAnnouncementUserListSQL,
				takesEntityIDArg: true,
			},
			admindomain.CourseEntityType: {
				checkExists:      courseRep.CheckIfCourseExistsByID,
				recipientSQL:     getCourseAnnouncementUserListSQL,
				takesEntityIDArg: true,
			},
		},
	}
}

func (w *AnnouncementFanOutWorker) queryRunner(ctx context.Context) db.QueryRunner {
	return db.FallbackQueryRunner(ctx, w.db)
}

// Run starts the BLPop event loop, processing messages until ctx is cancelled.
func (w *AnnouncementFanOutWorker) Run(ctx context.Context) {
	for {
		result, err := w.redisClient.BLPop(ctx, 5*time.Second, w.eventType).Result()
		isCont, isRet := handleRunBLPopErrors(err, w.eventType, w.logger)
		if isCont {
			continue
		}

		if isRet {
			return
		}

		if len(result) < 2 {
			w.logger.Error(fmt.Errorf("%s: BLPop: unexpected result length %d, want 2", w.eventType, len(result)), nil)
			continue
		}

		payload, announcement, key, msgErr := w.parseMessage(ctx, result[1])
		if errors.Is(msgErr, errAlreadyProcessed) {
			continue
		}
		if msgErr != nil {
			w.logger.Error(msgErr, nil)
			continue
		}

		w.processAndHandleFailure(ctx, payload, announcement, key)
	}
}

func (w *AnnouncementFanOutWorker) parseMessage(ctx context.Context, message string) (*events.AnnouncementPayload, *admindomain.Announcement, string, error) {
	var payload events.AnnouncementPayload
	if err := json.Unmarshal([]byte(message), &payload); err != nil {
		return nil, nil, "", fmt.Errorf("%s: unmarshal: %w", w.eventType, err)
	}
	announcement, err := w.validatePayload(ctx, payload)
	if err != nil {
		return nil, nil, "", err
	}

	key := w.generateIdempotencyKey(payload)
	ok, err := w.redisClient.SetNX(ctx, key, 1, 24*time.Hour).Result()
	if err != nil {
		return nil, nil, "", fmt.Errorf("%s: idempotency check: %w", w.eventType, err)
	}
	if !ok {
		return nil, nil, "", errAlreadyProcessed
	}
	return &payload, announcement, key, nil
}

func (w *AnnouncementFanOutWorker) validatePayload(ctx context.Context, p events.AnnouncementPayload) (*admindomain.Announcement, error) {
	if p.AnnouncementID == "" {
		return nil, fmt.Errorf("announcement: invalid payload: missing fields: AnnouncementID")
	}

	announcement, err := w.fetchAnnouncement(ctx, p.AnnouncementID)
	if err != nil {
		return nil, err
	}

	if err := w.checkEntityExists(ctx, announcement); err != nil {
		return nil, err
	}

	return announcement, nil
}

// fetchAnnouncement loads the announcement by ID, mapping "not found" to an error.
func (w *AnnouncementFanOutWorker) fetchAnnouncement(ctx context.Context, id string) (*admindomain.Announcement, error) {
	announcement, err := w.announcementRep.GetAnnouncementByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("announcement: worker.fetchAnnouncement: %w", err)
	}

	if announcement == nil {
		return nil, fmt.Errorf("announcement: invalid payload: announcement with id %s doesn't exists", id)
	}

	return announcement, nil
}

// checkEntityExists verifies the entity a scoped announcement links to (course/content
// item/article) still exists. Platform-wide announcements (EntityType == nil) have
// nothing to check.
func (w *AnnouncementFanOutWorker) checkEntityExists(ctx context.Context, announcement *admindomain.Announcement) error {
	if announcement.EntityType == nil {
		return nil
	}

	cfg, ok := w.entityConfigs[*announcement.EntityType]
	if !ok {
		return fmt.Errorf("announcement: invalid payload: invalid announcement entity type %s", *announcement.EntityType)
	}

	exists, err := cfg.checkExists(ctx, *announcement.EntityID)
	if err != nil {
		return fmt.Errorf("announcement: invalid payload: %w", err)
	}

	if !exists {
		return fmt.Errorf("announcement: invalid payload: such %s with id %s doesn't exists", *announcement.EntityType, *announcement.EntityID)
	}

	return nil
}

func (w *AnnouncementFanOutWorker) processAndHandleFailure(ctx context.Context, payload *events.AnnouncementPayload, announcement *admindomain.Announcement, key string) {
	if err := retry.Do(ctx, attemptsCount, func() error {
		return w.fanOut(ctx, announcement)
	}); err != nil {
		w.logger.Error(err, nil)
		w.dlq.Write(ctx, w.eventType, w.aggregationType, payload, err, attemptsCount)
		if delErr := w.redisClient.Del(ctx, key).Err(); delErr != nil {
			w.logger.Error(fmt.Errorf("%s: idempotency key cleanup: %w", w.eventType, delErr), nil)
		}
	}
}

func (w *AnnouncementFanOutWorker) generateIdempotencyKey(p events.AnnouncementPayload) string {
	return fmt.Sprintf("announcement:approved:%s", p.AnnouncementID)
}

func (w *AnnouncementFanOutWorker) fanOut(ctx context.Context, p *admindomain.Announcement) error {
	query, args, err := w.buildRecipientQuery(p)
	if err != nil {
		return fmt.Errorf("AnnouncementFanOutWorker.fanOut: %w", err)
	}

	recipients, err := w.queryRecipients(ctx, query, args)
	if err != nil {
		return fmt.Errorf("AnnouncementFanOutWorker.fanOut: %w", err)
	}

	for batch := range slices.Chunk(recipients, 500) {
		query, args := w.buildInsertBatch(batch, p)
		if err := w.insertBatch(ctx, query, args); err != nil {
			return fmt.Errorf("AnnouncementFanOutWorker.fanOut: %w", err)
		}
	}

	return nil
}

// rowPlaceholders returns "($start, $start+1, ..., $start+cols-1)".
func rowPlaceholders(start, cols int) string {
	ph := make([]string, cols)
	for i := range cols {
		ph[i] = fmt.Sprintf("$%d", start+i)
	}
	return "(" + strings.Join(ph, ", ") + ")"
}

func (w *AnnouncementFanOutWorker) buildInsertBatch(recipients []string, p *admindomain.Announcement) (query string, args []any) {
	const cols = 2

	rows := make([]string, len(recipients))
	args = make([]any, 0, len(recipients)*cols)

	for idx, r := range recipients {
		rows[idx] = rowPlaceholders(idx*cols+1, cols)
		args = append(args, r, p.ID)
	}

	query = strings.ReplaceAll(insertAnnouncementEmailDeliverQuery, "{setter}", strings.Join(rows, ",\n"))
	return query, args
}

// insertBatch inserts the batch, relying on ON CONFLICT (user_id, announcement_id) DO
// NOTHING to make a retried fanOut (after a partial batch failure) safe to re-run — a
// row count below len(args)/2 is expected on retry (already-inserted rows are skipped),
// not an error condition, so it isn't checked here.
func (w *AnnouncementFanOutWorker) insertBatch(ctx context.Context, query string, args []any) error {
	if _, err := w.queryRunner(ctx).Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("AnnouncementFanOutWorker.insertBatch: %w", err)
	}

	return nil
}

// buildRecipientQuery picks the recipient SQL for the announcement's EntityType and the
// args it needs — platform-wide/article queries take none, course/content queries take
// EntityID as $1. One place decides both, so the query and its arg list can never
// disagree about whether $1 is expected.
func (w *AnnouncementFanOutWorker) buildRecipientQuery(p *admindomain.Announcement) (query string, args []any, err error) {
	if p.EntityType == nil {
		return getAnnouncementUserListSQL, nil, nil
	}

	cfg, ok := w.entityConfigs[*p.EntityType]
	if !ok {
		return "", nil, fmt.Errorf("entity type is not valid")
	}

	if cfg.takesEntityIDArg {
		return cfg.recipientSQL, []any{p.EntityID}, nil
	}
	return cfg.recipientSQL, nil, nil
}

func (w *AnnouncementFanOutWorker) queryRecipients(ctx context.Context, query string, args []any) ([]string, error) {
	rows, err := w.queryRunner(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("AnnouncementFanOutWorker.queryRecipients: %w", err)
	}

	defer rows.Close()

	var recipients []string
	for rows.Next() {
		var val string
		if scanErr := rows.Scan(&val); scanErr != nil {
			return nil, fmt.Errorf("AnnouncementFanOutWorker.queryRecipients: scan: %w", scanErr)
		}
		recipients = append(recipients, val)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("AnnouncementFanOutWorker.queryRecipients: %w", err)
	}

	return recipients, nil
}
