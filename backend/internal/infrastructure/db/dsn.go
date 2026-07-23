package db

import (
	"errors"
	"fmt"
	"net/url"

	"learnflow_backend/internal/infrastructure/env"
)

// BuildDSNFromEnv constructs a PostgreSQL DSN from individual DB_* environment variables.
// Returns an error if any required variable (DB_NAME, DB_USER, DB_HOST, DB_PASSWORD) is empty.
func BuildDSNFromEnv() (string, error) {
	dbPort := env.GetIntEnv("DB_PORT", 5432)
	dbName := env.GetStringEnv("DB_NAME", "")
	dbUser := env.GetStringEnv("DB_USER", "")
	dbHost := env.GetStringEnv("DB_HOST", "")
	dbPassword := env.GetStringEnv("DB_PASSWORD", "")
	dbSSLMode := env.GetStringEnv("DB_SSLMODE", "require")

	switch {
	case dbName == "":
		return "", errors.New("db: missing DB_NAME")
	case dbUser == "":
		return "", errors.New("db: missing DB_USER")
	case dbHost == "":
		return "", errors.New("db: missing DB_HOST")
	case dbPassword == "":
		return "", errors.New("db: missing DB_PASSWORD")
	}

	return fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=%s", url.UserPassword(dbUser, dbPassword), dbHost, dbPort, dbName, dbSSLMode), nil
}

// MaskDSN redacts dsn's password for logging (CWE-532: the field sanitizer misses
// credentials inline in a URL). Returns dsn unchanged if unparseable or password-less.
func MaskDSN(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil || u.User == nil {
		return dsn
	}
	if _, hasPassword := u.User.Password(); hasPassword {
		u.User = url.UserPassword(u.User.Username(), "***")
	}
	return u.String()
}
