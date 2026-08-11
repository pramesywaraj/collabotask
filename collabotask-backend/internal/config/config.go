package config

import (
	"fmt"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Log      LogConfig
	CORS     CORSConfig
	Auth     AuthConfig
}

type AppConfig struct {
	Environment string
	AppName     string
	Version     string
	Debug       bool
}

type ServerConfig struct {
	Port    string
	Host    string
	Timeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type AuthConfig struct {
	JWTSecret     string
	JWTExpiration time.Duration
	BcryptCost    int
	CookieSecure  bool
	CookieName    string
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

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
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

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		App: AppConfig{
			Environment: getEnv("APP_ENV", "development"),
			AppName:     getEnv("APP_NAME", "collabotask"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
			Debug:       getEnvBool("APP_DEBUG", false),
		},
		Server: ServerConfig{
			Port:    getEnv("SERVER_PORT", "8080"),
			Host:    getEnv("SERVER_HOST", "localhost"),
			Timeout: getEnvDuration("SERVER_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", ""),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", ""),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxConns:        getEnvInt("DB_MAX_CONNS", 25),
			MinConns:        getEnvInt("DB_MIN_CONNS", 5),
			MaxConnLifetime: getEnvDuration("DB_MAX_CONN_LIFETIME", 0),
			MaxConnIdleTime: getEnvDuration("DB_MAX_CONN_IDLE_TIME", 0),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "console"),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getEnvStringSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			AllowedMethods:   getEnvStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders:   getEnvStringSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "X-CSRF-Protection"}),
			AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           getEnvInt("CORS_MAX_AGE", 3600),
		},
		Auth: AuthConfig{
			JWTSecret:     getEnv("AUTH_JWT_SECRET", ""),
			JWTExpiration: getEnvDuration("AUTH_JWT_EXPIRATION", 24*time.Hour),
			BcryptCost:    getEnvInt("AUTH_BCRYPT_COST", 12),
			CookieSecure:  getEnvBool("COOKIE_SECURE", true),
			CookieName:    getEnv("COOKIE_NAME", "__Host-token"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// WSOriginPatterns derives host-only patterns for coder/websocket's OriginPatterns
// from the CORS allowlist. CORS stores full origins (with scheme) for reflection;
// websocket.Accept's OriginPatterns matches only the host part.
// Called once at startup after Validate() ensures every origin is parseable.
func (c *Config) WSOriginPatterns() []string {
	patterns := make([]string, 0, len(c.CORS.AllowedOrigins))
	for _, origin := range c.CORS.AllowedOrigins {
		if origin == "*" {
			continue // "*" has no host; cookie auth forbids it anyway (guarded in Validate)
		}
		u, _ := url.Parse(origin) // validated in Validate(); no error possible here
		patterns = append(patterns, u.Host)
	}
	return patterns
}

func (c *Config) Validate() error {
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}

	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("AUTH_JWT_SECRET is required")
	}

	if c.CORS.AllowCredentials && slices.Contains(c.CORS.AllowedOrigins, "*") {
		return fmt.Errorf("CORS_ALLOWED_ORIGINS must not contain * when CORS_ALLOW_CREDENTIALS is true")
	}

	// Fail-fast: unparseable CORS origin → WS origin pattern derivation would silently
	// reject every cross-origin handshake. Mirrors the *+credentials guard above.
	for _, origin := range c.CORS.AllowedOrigins {
		if origin == "*" {
			continue // already guarded above
		}
		u, err := url.Parse(origin)
		if err != nil || u.Host == "" {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS: %q is not a valid origin URL (needs scheme + host)", origin)
		}
	}

	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}

	if !validEnvs[c.App.Environment] {
		return fmt.Errorf("APP_ENV must be one of: development, staging, production")
	}

	return nil
}
