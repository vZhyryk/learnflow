// Package rediskeys holds Redis key formats shared by the auth middleware and the services.
package rediskeys

const (
	// UserBlockedPrefix prefixes the key that makes the middleware reject a user's access tokens.
	UserBlockedPrefix = "user_blocked:"
	// JTIBlocklistPrefix prefixes the key of a single blocklisted access token (by jti).
	JTIBlocklistPrefix = "blocklist:"
)

// UserBlocked returns the Redis key marking userID as blocked.
func UserBlocked(userID string) string {
	return UserBlockedPrefix + userID
}

// JTIBlocked returns the Redis key marking the access token with the given jti as blocklisted.
func JTIBlocked(jti string) string {
	return JTIBlocklistPrefix + jti
}
