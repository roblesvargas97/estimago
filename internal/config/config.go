package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	Port          string
	AppEnv        string
	AuthJWTSecret string
	AuthJWTTTLHrs int
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Port:          os.Getenv("PORT"),
		AppEnv:        os.Getenv("APP_ENV"),
		AuthJWTSecret: os.Getenv("AUTH_JWT_SECRET"),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL no configurado")
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	if cfg.AppEnv == "" {
		cfg.AppEnv = "dev"
	}

	ttlStr := os.Getenv("AUTH_JWT_TTL_HOURS")
	if ttlStr == "" {
		cfg.AuthJWTTTLHrs = 72
	} else {
		ttl, err := strconv.Atoi(ttlStr)
		if err != nil || ttl <= 0 {
			cfg.AuthJWTTTLHrs = 72
		} else {
			cfg.AuthJWTTTLHrs = ttl
		}
	}

	if cfg.AuthJWTSecret == "" {
		log.Fatal("AUTH_JWT_SECRET no configurado")
	}

	return cfg
}
