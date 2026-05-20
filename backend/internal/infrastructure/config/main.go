// Package config loads and resolves runtime configuration from flags and environment variables.
package config

// Config holds the parsed configuration shared by all service entrypoints.
type Config struct {
	DSN       string
	Direction string
	Steps     int
}
