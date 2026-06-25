package tokens

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenByteLength = 32
	jtibyteLength   = 32
)

// GenerateSecureToken generates a cryptographically random token and returns both the raw value and its SHA-256 hex hash.
func GenerateSecureToken() (raw, hash string, err error) {
	b := make([]byte, tokenByteLength)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generateSecureToken: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(sum[:])
	return raw, hash, nil
}

func generateJTI() (string, error) {
	b := make([]byte, jtibyteLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generateJTI: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// Tokens holds JWT signing configuration and exposes token generation and validation.
type Tokens struct {
	secret     string
	prevSecret string
	issuer     string
	audience   string
}

// NewTokens returns a Tokens instance configured with the provided JWT secret, optional previous secret for rotation, issuer, and audience.
func NewTokens(secret, prevSecret, issuer, audience string) *Tokens {
	return &Tokens{secret: secret, prevSecret: prevSecret, issuer: issuer, audience: audience}
}

type jwtClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a signed HS256 JWT for the given user with the specified TTL.
func (t *Tokens) GenerateAccessToken(user *authdomain.User, accessTokenTTL time.Duration) (string, error) {
	rClaimsID, err := generateJTI()
	if err != nil {
		return "", fmt.Errorf("generateAccessToken: %w", err)
	}

	claims := jwtClaims{
		Role: string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Audience:  jwt.ClaimStrings{t.audience},
			ID:        rClaimsID,
			Issuer:    t.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(t.secret))
	if err != nil {
		return "", fmt.Errorf("generateAccessToken: %w", err)
	}

	return signed, nil
}

// ValidateToken parses and validates a JWT, falling back to prevSecret on signature mismatch to support key rotation.
func (t *Tokens) ValidateToken(tokenString string) (*jwtClaims, error) {
	claims, err := t.validateWithSecret(tokenString, t.secret)
	if err == nil {
		return claims, nil
	}

	if t.prevSecret != "" && errors.Is(err, jwt.ErrSignatureInvalid) {
		return t.validateWithSecret(tokenString, t.prevSecret)
	}

	return nil, err
}

func (t *Tokens) validateWithSecret(tokenString, secret string) (*jwtClaims, error) {
	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, keyFunc, jwt.WithAudience(t.audience), jwt.WithIssuer(t.issuer), jwt.WithExpirationRequired())
	if err != nil {
		return nil, fmt.Errorf("validateWithSecret: parse error: %w", err)
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("validateWithSecret: invalid token claims")
	}

	if claims.IssuedAt == nil {
		return nil, fmt.Errorf("validateWithSecret: missing iat claim")
	}

	return claims, nil
}
