package config

import "os"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

func Load() Config {
	return Config{
		HTTPAddr:    env("HTTP_ADDR", ":8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://postgres:1234@127.0.0.1:5432/tn_svetofor_online?sslmode=disable"),
	}
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
