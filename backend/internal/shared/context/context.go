package appcontext

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	authdomain "learnflow_backend/internal/auth/domain"
)

type (
	requestIDKey            struct{}
	ipAddressKey            struct{}
	userKey                 struct{}
	jtiKey                  struct{}
	accessTokenExpiresAtKey struct{}
)

// NewRequestID generates a random RFC 4122 v4 UUID.
func NewRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Errorf("newRequestID: %w", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant is 10
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	)
}

// WithRequestID stores a request ID in the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// WithIPAddress stores the client IP address in the context.
func WithIPAddress(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ipAddressKey{}, ip)
}

// RequestIDFromContext extracts the request ID from the given context.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// IPAddressFromContext extracts the client IP address from the given context.
func IPAddressFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(ipAddressKey{}).(string); ok {
		return id
	}
	return ""
}

// WithUser stores the authenticated user in ctx.
func WithUser(ctx context.Context, user *authdomain.User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

// UserFromContext retrieves the authenticated user from ctx.
func UserFromContext(ctx context.Context) (*authdomain.User, bool) {
	user, ok := ctx.Value(userKey{}).(*authdomain.User)
	return user, ok && user != nil
}

// WithJTI stores the JWT ID in ctx.
func WithJTI(ctx context.Context, jti string) context.Context {
	return context.WithValue(ctx, jtiKey{}, jti)
}

// JTIFromContext retrieves the JWT ID from ctx.
func JTIFromContext(ctx context.Context) string {
	if jti, ok := ctx.Value(jtiKey{}).(string); ok {
		return jti
	}
	return ""
}

// WithAccessTokenExpiresAt stores the access token expiry time in ctx.
func WithAccessTokenExpiresAt(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, accessTokenExpiresAtKey{}, t)
}

// AccessTokenExpiresAtFromContext retrieves the access token expiry time from ctx.
func AccessTokenExpiresAtFromContext(ctx context.Context) time.Time {
	if t, ok := ctx.Value(accessTokenExpiresAtKey{}).(time.Time); ok {
		return t
	}
	return time.Time{}
}
