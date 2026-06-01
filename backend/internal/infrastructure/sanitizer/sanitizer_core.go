package sanitizer

// Sanitizer redacts sensitive values from arbitrary Go values before they are logged or serialised.
type Sanitizer struct {
	redactedValue           string
	maxStringLen            int
	sensitiveKeys           map[string]struct{}
	normalizedSensitiveKeys map[string]struct{}
}

// NewSanitizer creates a Sanitizer. Zero/nil values for any parameter use safe defaults:
// redactedValue → "[REDACTED]", maxStringLength → 2000, sensitiveKeys → DefaultSensitiveKeys().
func NewSanitizer(redactedValue string, maxStringLength int, sensitiveKeys map[string]struct{}) *Sanitizer {
	if redactedValue == "" {
		redactedValue = "[REDACTED]"
	}
	if maxStringLength <= 0 {
		maxStringLength = 2000
	}
	if sensitiveKeys == nil {
		sensitiveKeys = DefaultSensitiveKeys()
	}

	return &Sanitizer{
		redactedValue:           redactedValue,
		maxStringLen:            maxStringLength,
		sensitiveKeys:           sensitiveKeys,
		normalizedSensitiveKeys: normalizeSensitiveKeys(sensitiveKeys),
	}
}

// DefaultSensitiveKeys returns the built-in set of key names treated as sensitive
// (auth tokens, passwords, API keys, cookies, secrets, etc.).
func DefaultSensitiveKeys() map[string]struct{} {
	return map[string]struct{}{
		// Auth schemes and headers.
		"basic":               {},
		"bearer":              {},
		"authorization":       {},
		"proxy-authorization": {},

		// Cookies and session data.
		"cookie":        {},
		"set-cookie":    {},
		"session_token": {},

		// API keys and access keys.
		"x-api-key":  {},
		"api-key":    {},
		"apikey":     {},
		"api_secret": {},
		"access_key": {},

		// Tokens.
		"token":              {},
		"access_token":       {},
		"refresh_token":      {},
		"id_token":           {},
		"jwt":                {},
		"x-amz-access-token": {},

		// Client/application identifiers and secrets.
		"client_id":           {},
		"client_name":         {},
		"client_secret":       {},
		"platform_token_data": {},

		// Passwords and generic secrets.
		"password":    {},
		"passwd":      {},
		"pwd":         {},
		"secret":      {},
		"secret_key":  {},
		"private_key": {},
		"signature":   {},

		// Common combined forms from Go struct fields.
		"authorization_basic": {},

		"dsn":          {},
		"database_url": {},
	}
}
