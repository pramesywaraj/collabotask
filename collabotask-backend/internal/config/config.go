package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port    string
	Host    string
	Timeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type LogConfig struct {
	Level  string
	Format string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}

	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}

	return defaultValue
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:    getEnv("SERVER_PORT", "8080"),
			Host:    getEnv("SERVER_HOST", "localhost"),
			Timeout: getEnvDuration("SERVER_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "console"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	return nil
}
