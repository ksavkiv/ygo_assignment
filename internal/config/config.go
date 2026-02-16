package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddr  string
	DatabaseURL string
	RedisAddr   string
	APIToken    string
	LogLevel    string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		ServerAddr:  getEnv("SERVER_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/destinations?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		APIToken:    getEnv("API_TOKEN", ""),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
