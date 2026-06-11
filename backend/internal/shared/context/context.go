package appcontext

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type requestIDKey struct{}
type ipAddressKey struct{}

// NewRequestID generates a random RFC 4122 v4 UUID.
func NewRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Errorf("newRequestID: %w", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant is 10
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	)
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

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
