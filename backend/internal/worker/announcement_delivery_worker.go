package worker

import (
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/mailer"

	"github.com/redis/go-redis/v9"
)

// NewAnnouncementDeliveryWorker returns an EmailWorker configured to handle announcement
// delivery events — Worker Б (step 5) of the announcement notification flow.
func NewAnnouncementDeliveryWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m Mailer,
	baseURL string,
) *EmailWorker[AnnouncementDeliver] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseURL, Config[AnnouncementDeliver]{
		EventType:       string(events.EventAnnouncementDeliver),
		AggregationType: string(events.AggregationTypeEmail),
		IdempotencyKey:  GenerateAnnouncementIdempotencyKey,
		Validate:        ValidateAnnouncementPayload,
		Process:         HandleAnnouncementProcess,
	})
}

// ValidateAnnouncementPayload checks that all required fields are present in the payload.
func ValidateAnnouncementPayload(p AnnouncementDeliver) error {
	var missing []string
	if p.UserID == "" {
		missing = append(missing, "UserID")
	}
	if p.Email == "" {
		missing = append(missing, "Email")
	}
	if p.AnnouncementID == "" {
		missing = append(missing, "AnnouncementID")
	}

	if p.FirstName == "" {
		missing = append(missing, "FirstName")
	}

	if p.Title == "" {
		missing = append(missing, "Title")
	}

	if p.Body == "" {
		missing = append(missing, "Body")
	}

	if len(missing) > 0 {
		return fmt.Errorf("announcement_delivery: invalid payload: missing fields: %v", missing)
	}
	return nil
}

// HandleAnnouncementProcess sends the announcement email for the given payload.
func HandleAnnouncementProcess(p AnnouncementDeliver, _ string, m Mailer) error {
	data := map[string]string{
		"name":  p.FirstName,
		"title": p.Title,
		"body":  p.Body,
	}
	return m.Send("announcement_delivery.html", data, mailer.CCUser{Mail: p.Email}, nil)
}

// GenerateAnnouncementIdempotencyKey returns a Redis key used to deduplicate
// announcement delivery processing.
func GenerateAnnouncementIdempotencyKey(p AnnouncementDeliver) string {
	return fmt.Sprintf("processed:announcement_delivery:%s:%s", p.UserID, p.AnnouncementID)
}
