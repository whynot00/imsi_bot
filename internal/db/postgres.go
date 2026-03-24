// Package config loads application configuration from environment variables.
// Values are read from the process environment; use a .env file with
// github.com/joho/godotenv to populate it before calling Load.
package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration for the application.
type Config struct {
	Postgres PostgresConfig
	LogLevel string
}

// PostgresConfig holds connection parameters for PostgreSQL.
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
	SSLMode  string
}

// DSN returns a lib/pq compatible connection string.
func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DB, c.SSLMode,
	)
}

// Load reads configuration from environment variables.
// Call godotenv.Load() before this if using a .env file.
func Load() (*Config, error) {
	cfg := &Config{
		Postgres: PostgresConfig{
			Host:     env("POSTGRES_HOST", "localhost"),
			Port:     env("POSTGRES_PORT", "5432"),
			User:     required("POSTGRES_USER"),
			Password: required("POSTGRES_PASSWORD"),
			DB:       required("POSTGRES_DB"),
			SSLMode:  env("POSTGRES_SSLMODE", "disable"),
		},
		LogLevel: env("LOG_LEVEL", "info"),
	}
	return cfg, nil
}

// env returns the value of the environment variable or the default.
func env(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// required returns the value of the environment variable or panics.
func required(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}
