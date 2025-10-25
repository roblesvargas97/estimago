package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)


type Config struct {
	DatabaseURL string
	Port        string
	AppEnv      string
}


func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
		AppEnv:      os.Getenv("APP_ENV"),
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

	return cfg
}
