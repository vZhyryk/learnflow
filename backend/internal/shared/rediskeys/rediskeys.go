// Package rediskeys holds Redis key formats shared by the auth middleware and the services.
package rediskeys

const (
	// UserBlockedPrefix prefixes the key that makes the middleware reject a user's access tokens.
	UserBlockedPrefix = "user_blocked:"
	// JTIBlockedPrefix prefixes the key of a single blocklisted access token (by jti).
	JTIBlockedPrefix = "blocklist:"
	// UserRoleRevokedPrefix prefixes the key that makes RequireRole reject a user whose role was just revoked.
	UserRoleRevokedPrefix = "role_revoked:"
)

// UserBlocked returns the Redis key marking userID as blocked.
func UserBlocked(userID string) string {
	return UserBlockedPrefix + userID
}

// JTIBlocked returns the Redis key marking the access token with the given jti as blocklisted.
func JTIBlocked(jti string) string {
	return JTIBlockedPrefix + jti
}

// UserRoleRevoked returns the Redis key marking userID's role as revoked.
func UserRoleRevoked(userID string) string {
	return UserRoleRevokedPrefix + userID
}
