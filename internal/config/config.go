package config

import (
	"fmt"
	"os"
)

type Config struct {
	Postgres PostgresConfig
	Port     string
	LogLevel string
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
	SSLMode  string
}

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DB, c.SSLMode,
	)
}

func Load() (*Config, error) {
	return &Config{
		Postgres: PostgresConfig{
			Host:     env("POSTGRES_HOST", "localhost"),
			Port:     env("POSTGRES_PORT", "5432"),
			User:     required("POSTGRES_USER"),
			Password: required("POSTGRES_PASSWORD"),
			DB:       required("POSTGRES_DB"),
			SSLMode:  env("POSTGRES_SSLMODE", "disable"),
		},
		Port:     env("PORT", "8080"),
		LogLevel: env("LOG_LEVEL", "info"),
	}, nil
}

func env(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func required(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}
