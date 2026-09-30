package worker

import (
	"fmt"
	"learnflow_backend/internal/events"
)

const expirationTimeLayout = "2 Jan 2006, 15:04 UTC"

// tokenFields returns the template variables shared by every token-link email:
// the user's name, the link built from path and the raw token, and the expiry time.
func tokenFields(urlKey, path string) []SchemaField[events.TokenPayload] {
	return []SchemaField[events.TokenPayload]{
		{Name: "name", Value: func(p events.TokenPayload, _ string) string { return p.UserName }},
		{Name: urlKey, Value: func(p events.TokenPayload, baseURL string) string {
			return fmt.Sprintf("%s%s?token=%s", baseURL, path, p.RawToken)
		}},
		{Name: "expirationTime", Value: func(p events.TokenPayload, _ string) string {
			return p.ExpiresAt.UTC().Format(expirationTimeLayout)
		}},
	}
}

// userNotificationFields is the template variable set of the user block/unblock/delete/restore emails.
func userNotificationFields() []SchemaField[events.UserNotificationPayload] {
	return []SchemaField[events.UserNotificationPayload]{
		{Name: "name", Value: func(p events.UserNotificationPayload, _ string) string { return p.UserName }},
	}
}
