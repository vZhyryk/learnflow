// Package env reads typed values from environment variables with safe fallbacks.
package env

import (
	"os"
	"strconv"
	"strings"
)

// GetStringEnv returns the string value of key, trimming whitespace.
// Returns fallback if the variable is unset or blank.
func GetStringEnv(key, fallback string) string {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	return raw
}

// GetIntEnv returns the integer value of key.
// Returns fallback if the variable is unset, blank, or not a valid integer.
func GetIntEnv(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return v
}

// GetFloat64Env returns the float64 value of key.
// Returns fallback if the variable is unset, blank, or not a valid float.
func GetFloat64Env(key string, fallback float64) float64 {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return fallback
	}
	return v
}

// GetBoolEnv returns the boolean value of key, accepting "true/1/yes" and "false/0/no".
// Returns fallback if the variable is unset, blank, or unrecognised.
func GetBoolEnv(key string, fallback bool) bool {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return fallback
	}
}
