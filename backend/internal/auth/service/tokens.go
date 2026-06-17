package authservice

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenByteLength = 32
	jtibyteLength   = 16
)

func generateSecureToken() (raw, hash string, err error) {
	b := make([]byte, tokenByteLength)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generateSecureToken: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(sum[:])
	return raw, hash, nil
}

type jwtClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (s *Service) generateAccessToken(user *authdomain.User) (string, error) {
	rClaimsID, err := generateJTI()
	if err != nil {
		return "", fmt.Errorf("generateAccessToken: %w", err)
	}

	claims := jwtClaims{
		Role: string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Audience:  jwt.ClaimStrings{"learnflow-api"},
			ID:        rClaimsID,
			Issuer:    "learnflow-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("generateAccessToken: %w", err)
	}

	return signed, nil
}

func generateJTI() (string, error) {
	b := make([]byte, jtibyteLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generateJTI: %w", err)
	}
	return hex.EncodeToString(b), nil
}
