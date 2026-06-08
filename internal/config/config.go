// Package config loads application configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DBUrlMigration string
	SecretJwt      string

	DBHost     string
	DBUser     string
	DBName     string
	DBPassword string
	DBPort     string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnvOrDefault("PORT", "8080"),
		DBUrlMigration: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		SecretJwt:      strings.TrimSpace(os.Getenv("SECRET_JWT")),
		DBHost:         strings.TrimSpace(os.Getenv("DB_HOST")),
		DBUser:         strings.TrimSpace(os.Getenv("DB_USER")),
		DBName:         strings.TrimSpace(os.Getenv("DB_NAME")),
		DBPassword:     strings.TrimSpace(os.Getenv("DB_PASSWORD")),
		DBPort:         getEnvOrDefault("DB_PORT", "3306"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) ServerAddress() string {
	return ":" + c.Port
}

func (c *Config) validate() error {
	missingFields := make([]string, 0, 5)

	if c.SecretJwt == "" {
		missingFields = append(missingFields, "SECRET_JWT")
	}
	if c.DBHost == "" {
		missingFields = append(missingFields, "DB_HOST")
	}
	if c.DBUser == "" {
		missingFields = append(missingFields, "DB_USER")
	}
	if c.DBName == "" {
		missingFields = append(missingFields, "DB_NAME")
	}
	if c.DBPassword == "" {
		missingFields = append(missingFields, "DB_PASSWORD")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingFields, ", "))
	}

	return nil
}

func getEnvOrDefault(key string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	return value
}
