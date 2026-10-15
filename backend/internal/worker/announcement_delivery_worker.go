package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"

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
		Name:            "announcement_delivery",
		Schema:          "announcement_delivery.html",
		SchemaFields: []SchemaField[AnnouncementDeliver]{
			{Name: "name", Value: func(p AnnouncementDeliver, _ string) string { return p.FirstName }},
			{Name: "title", Value: func(p AnnouncementDeliver, _ string) string { return p.Title }},
			{Name: "body", Value: func(p AnnouncementDeliver, _ string) string { return p.Body }},
		},
	})
}
