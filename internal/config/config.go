package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Postgers Postgres
	Telegram Telegram
}

type Postgres struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
	SSLMode  string
}

type Telegram struct {
	Token string
}

func Load() *Config {

	godotenv.Load()

	config := Config{
		Postgers: Postgres{
			Host:     "localhost",
			Port:     "5436",
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			DB:       "datapump_db",
			SSLMode:  "disable",
		},
		Telegram: Telegram{
			Token: os.Getenv("TELEGRAM_TOKEN"),
		},
	}

	return &config
}
