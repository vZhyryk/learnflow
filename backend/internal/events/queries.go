package events

const (
	queryInsertOutbox = `
		INSERT INTO event_outbox (aggregate_type, aggregate_id, event_type, payload_json)
		VALUES ($1, $2, $3, $4)
	`
)
