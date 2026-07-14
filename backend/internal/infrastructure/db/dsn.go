package db

import (
	"errors"
	"fmt"
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
	dbSSLMode := env.GetStringEnv("DB_SSLMODE", "disable")

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

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode), nil
}
