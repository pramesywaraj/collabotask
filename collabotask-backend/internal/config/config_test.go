package config

import (
	"os"
	"testing"
	"time"
)

// setRequiredEnvVars sets the minimum env vars needed to pass Validate().
func setRequiredEnvVars() {
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("AUTH_JWT_SECRET", "test-secret")
}

func TestConfigLoad(t *testing.T) {
	originalPort := os.Getenv("SERVER_PORT")
	originalDBName := os.Getenv("DB_NAME")
	originalDBUser := os.Getenv("DB_USER")
	originalJWTSecret := os.Getenv("AUTH_JWT_SECRET")

	defer func() {
		restoreEnv("SERVER_PORT", originalPort)
		restoreEnv("DB_NAME", originalDBName)
		restoreEnv("DB_USER", originalDBUser)
		restoreEnv("AUTH_JWT_SECRET", originalJWTSecret)
	}()

	setRequiredEnvVars()
	os.Setenv("SERVER_PORT", "3000")

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Server.Port != "3000" {
		t.Errorf("Expected port 3000, got %s", config.Server.Port)
	}
	if config.Database.Name != "testdb" {
		t.Errorf("Expected db name testdb, got %s", config.Database.Name)
	}
	if config.Database.User != "testuser" {
		t.Errorf("Expected db user testuser, got %s", config.Database.User)
	}
}

func TestConfigDefaults(t *testing.T) {
	originalPort := os.Getenv("SERVER_PORT")
	originalHost := os.Getenv("SERVER_HOST")
	originalAppEnv := os.Getenv("APP_ENV")
	originalAppName := os.Getenv("APP_NAME")
	originalJWTSecret := os.Getenv("AUTH_JWT_SECRET")

	defer func() {
		restoreEnv("SERVER_PORT", originalPort)
		restoreEnv("SERVER_HOST", originalHost)
		restoreEnv("APP_ENV", originalAppEnv)
		restoreEnv("APP_NAME", originalAppName)
		restoreEnv("AUTH_JWT_SECRET", originalJWTSecret)
	}()

	setRequiredEnvVars()
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("SERVER_HOST")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("APP_NAME")

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Server.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", config.Server.Port)
	}
	if config.Server.Host != "localhost" {
		t.Errorf("Expected default host localhost, got %s", config.Server.Host)
	}
	if config.App.Environment != "development" {
		t.Errorf("Expected default environment development, got %s", config.App.Environment)
	}
	if config.App.AppName != "collabotask" {
		t.Errorf("Expected default app name collabotask, got %s", config.App.AppName)
	}
}

func TestConfigValidation(t *testing.T) {
	originalDBName := os.Getenv("DB_NAME")
	originalDBUser := os.Getenv("DB_USER")
	originalAppEnv := os.Getenv("APP_ENV")
	originalJWTSecret := os.Getenv("AUTH_JWT_SECRET")

	defer func() {
		restoreEnv("DB_NAME", originalDBName)
		restoreEnv("DB_USER", originalDBUser)
		restoreEnv("APP_ENV", originalAppEnv)
		restoreEnv("AUTH_JWT_SECRET", originalJWTSecret)
	}()

	os.Setenv("AUTH_JWT_SECRET", "test-secret")

	// Missing DB_NAME
	os.Unsetenv("DB_NAME")
	os.Setenv("DB_USER", "testuser")
	_, err := Load()
	if err == nil {
		t.Error("Expected error when DB_NAME is missing, got nil")
	}
	if err != nil && err.Error() != "config validation failed: DB_NAME is required" {
		t.Errorf("Expected 'DB_NAME is required' error, got: %v", err)
	}

	// Missing DB_USER
	os.Setenv("DB_NAME", "testdb")
	os.Unsetenv("DB_USER")
	_, err = Load()
	if err == nil {
		t.Error("Expected error when DB_USER is missing, got nil")
	}
	if err != nil && err.Error() != "config validation failed: DB_USER is required" {
		t.Errorf("Expected 'DB_USER is required' error, got: %v", err)
	}

	// Invalid APP_ENV
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("APP_ENV", "invalid")
	_, err = Load()
	if err == nil {
		t.Error("Expected error when APP_ENV is invalid, got nil")
	}
	if err != nil && err.Error() != "config validation failed: APP_ENV must be one of: development, staging, production" {
		t.Errorf("Expected 'APP_ENV must be one of...' error, got: %v", err)
	}
}

func TestValidate_WildcardWithCredentials(t *testing.T) {
	originalOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	originalCreds := os.Getenv("CORS_ALLOW_CREDENTIALS")
	originalJWTSecret := os.Getenv("AUTH_JWT_SECRET")

	defer func() {
		restoreEnv("CORS_ALLOWED_ORIGINS", originalOrigins)
		restoreEnv("CORS_ALLOW_CREDENTIALS", originalCreds)
		restoreEnv("AUTH_JWT_SECRET", originalJWTSecret)
	}()

	setRequiredEnvVars()
	os.Setenv("CORS_ALLOWED_ORIGINS", "*")
	os.Setenv("CORS_ALLOW_CREDENTIALS", "true")

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error when CORS uses * with AllowCredentials=true, got nil")
	}
}

func TestAuthCookieConfig(t *testing.T) {
	originalJWTSecret := os.Getenv("AUTH_JWT_SECRET")
	originalCookieSecure := os.Getenv("COOKIE_SECURE")
	originalCookieName := os.Getenv("COOKIE_NAME")

	defer func() {
		restoreEnv("AUTH_JWT_SECRET", originalJWTSecret)
		restoreEnv("COOKIE_SECURE", originalCookieSecure)
		restoreEnv("COOKIE_NAME", originalCookieName)
	}()

	setRequiredEnvVars()

	// Defaults
	os.Unsetenv("COOKIE_SECURE")
	os.Unsetenv("COOKIE_NAME")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if !cfg.Auth.CookieSecure {
		t.Error("Expected CookieSecure default true")
	}
	if cfg.Auth.CookieName != "__Host-token" {
		t.Errorf("Expected CookieName default __Host-token, got %s", cfg.Auth.CookieName)
	}

	// Custom values
	os.Setenv("COOKIE_SECURE", "false")
	os.Setenv("COOKIE_NAME", "token")

	cfg, err = Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg.Auth.CookieSecure {
		t.Error("Expected CookieSecure false")
	}
	if cfg.Auth.CookieName != "token" {
		t.Errorf("Expected CookieName token, got %s", cfg.Auth.CookieName)
	}
}

func TestConfigDurationParsing(t *testing.T) {
	originalTimeout := os.Getenv("SERVER_TIMEOUT")
	originalMaxConnLifetime := os.Getenv("DB_MAX_CONN_LIFETIME")
	originalJWTSecret := os.Getenv("AUTH_JWT_SECRET")

	defer func() {
		restoreEnv("SERVER_TIMEOUT", originalTimeout)
		restoreEnv("DB_MAX_CONN_LIFETIME", originalMaxConnLifetime)
		restoreEnv("AUTH_JWT_SECRET", originalJWTSecret)
	}()

	setRequiredEnvVars()

	os.Setenv("SERVER_TIMEOUT", "60s")
	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if config.Server.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", config.Server.Timeout)
	}

	os.Unsetenv("SERVER_TIMEOUT")
	config, err = Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if config.Server.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", config.Server.Timeout)
	}

	os.Setenv("DB_MAX_CONN_LIFETIME", "1h")
	config, err = Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if config.Database.MaxConnLifetime != time.Hour {
		t.Errorf("Expected max conn lifetime 1h, got %v", config.Database.MaxConnLifetime)
	}
}

func TestAppConfig(t *testing.T) {
	originalAppEnv := os.Getenv("APP_ENV")
	originalAppName := os.Getenv("APP_NAME")
	originalAppVersion := os.Getenv("APP_VERSION")
	originalAppDebug := os.Getenv("APP_DEBUG")
	originalJWTSecret := os.Getenv("AUTH_JWT_SECRET")

	defer func() {
		restoreEnv("APP_ENV", originalAppEnv)
		restoreEnv("APP_NAME", originalAppName)
		restoreEnv("APP_VERSION", originalAppVersion)
		restoreEnv("APP_DEBUG", originalAppDebug)
		restoreEnv("AUTH_JWT_SECRET", originalJWTSecret)
	}()

	setRequiredEnvVars()
	os.Setenv("APP_ENV", "staging")
	os.Setenv("APP_NAME", "testapp")
	os.Setenv("APP_VERSION", "2.0.0")
	os.Setenv("APP_DEBUG", "true")

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.App.Environment != "staging" {
		t.Errorf("Expected environment staging, got %s", config.App.Environment)
	}
	if config.App.AppName != "testapp" {
		t.Errorf("Expected app name testapp, got %s", config.App.AppName)
	}
	if config.App.Version != "2.0.0" {
		t.Errorf("Expected version 2.0.0, got %s", config.App.Version)
	}
	if !config.App.Debug {
		t.Errorf("Expected debug true, got %v", config.App.Debug)
	}
}

func TestDatabaseConfig(t *testing.T) {
	originalDBHost := os.Getenv("DB_HOST")
	originalDBPort := os.Getenv("DB_PORT")
	originalDBPassword := os.Getenv("DB_PASSWORD")
	originalDBSSLMode := os.Getenv("DB_SSLMODE")
	originalMaxConns := os.Getenv("DB_MAX_CONNS")
	originalMinConns := os.Getenv("DB_MIN_CONNS")
	originalJWTSecret := os.Getenv("AUTH_JWT_SECRET")

	defer func() {
		restoreEnv("DB_HOST", originalDBHost)
		restoreEnv("DB_PORT", originalDBPort)
		restoreEnv("DB_PASSWORD", originalDBPassword)
		restoreEnv("DB_SSLMODE", originalDBSSLMode)
		restoreEnv("DB_MAX_CONNS", originalMaxConns)
		restoreEnv("DB_MIN_CONNS", originalMinConns)
		restoreEnv("AUTH_JWT_SECRET", originalJWTSecret)
	}()

	setRequiredEnvVars()
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_PASSWORD", "secret123")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("DB_MAX_CONNS", "50")
	os.Setenv("DB_MIN_CONNS", "10")

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Database.Host != "db.example.com" {
		t.Errorf("Expected host db.example.com, got %s", config.Database.Host)
	}
	if config.Database.Port != "5433" {
		t.Errorf("Expected port 5433, got %s", config.Database.Port)
	}
	if config.Database.Password != "secret123" {
		t.Errorf("Expected password secret123, got %s", config.Database.Password)
	}
	if config.Database.SSLMode != "require" {
		t.Errorf("Expected SSL mode require, got %s", config.Database.SSLMode)
	}
	if config.Database.MaxConns != 50 {
		t.Errorf("Expected max conns 50, got %d", config.Database.MaxConns)
	}
	if config.Database.MinConns != 10 {
		t.Errorf("Expected min conns 10, got %d", config.Database.MinConns)
	}

	// Defaults
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_SSLMODE")
	os.Unsetenv("DB_MAX_CONNS")
	os.Unsetenv("DB_MIN_CONNS")

	config, err = Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Database.Host != "localhost" {
		t.Errorf("Expected default host localhost, got %s", config.Database.Host)
	}
	if config.Database.Port != "5432" {
		t.Errorf("Expected default port 5432, got %s", config.Database.Port)
	}
	if config.Database.SSLMode != "disable" {
		t.Errorf("Expected default SSL mode disable, got %s", config.Database.SSLMode)
	}
	if config.Database.MaxConns != 25 {
		t.Errorf("Expected default max conns 25, got %d", config.Database.MaxConns)
	}
	if config.Database.MinConns != 5 {
		t.Errorf("Expected default min conns 5, got %d", config.Database.MinConns)
	}
}

func TestLogConfig(t *testing.T) {
	originalLogLevel := os.Getenv("LOG_LEVEL")
	originalLogFormat := os.Getenv("LOG_FORMAT")
	originalJWTSecret := os.Getenv("AUTH_JWT_SECRET")

	defer func() {
		restoreEnv("LOG_LEVEL", originalLogLevel)
		restoreEnv("LOG_FORMAT", originalLogFormat)
		restoreEnv("AUTH_JWT_SECRET", originalJWTSecret)
	}()

	setRequiredEnvVars()
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "json")

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if config.Log.Level != "debug" {
		t.Errorf("Expected log level debug, got %s", config.Log.Level)
	}
	if config.Log.Format != "json" {
		t.Errorf("Expected log format json, got %s", config.Log.Format)
	}

	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LOG_FORMAT")

	config, err = Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if config.Log.Level != "info" {
		t.Errorf("Expected default log level info, got %s", config.Log.Level)
	}
	if config.Log.Format != "console" {
		t.Errorf("Expected default log format console, got %s", config.Log.Format)
	}
}

func TestCORSConfig(t *testing.T) {
	originalCORSOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	originalCORSMethods := os.Getenv("CORS_ALLOWED_METHODS")
	originalCORSHeaders := os.Getenv("CORS_ALLOWED_HEADERS")
	originalCORSAllowCreds := os.Getenv("CORS_ALLOW_CREDENTIALS")
	originalCORSMaxAge := os.Getenv("CORS_MAX_AGE")
	originalJWTSecret := os.Getenv("AUTH_JWT_SECRET")

	defer func() {
		restoreEnv("CORS_ALLOWED_ORIGINS", originalCORSOrigins)
		restoreEnv("CORS_ALLOWED_METHODS", originalCORSMethods)
		restoreEnv("CORS_ALLOWED_HEADERS", originalCORSHeaders)
		restoreEnv("CORS_ALLOW_CREDENTIALS", originalCORSAllowCreds)
		restoreEnv("CORS_MAX_AGE", originalCORSMaxAge)
		restoreEnv("AUTH_JWT_SECRET", originalJWTSecret)
	}()

	setRequiredEnvVars()

	// Custom values
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
	os.Setenv("CORS_ALLOWED_METHODS", "GET,POST")
	os.Setenv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization")
	os.Setenv("CORS_ALLOW_CREDENTIALS", "false")
	os.Setenv("CORS_MAX_AGE", "7200")

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	expectedOrigins := []string{"http://localhost:3000", "http://localhost:5173"}
	if len(config.CORS.AllowedOrigins) != len(expectedOrigins) {
		t.Errorf("Expected %d origins, got %d", len(expectedOrigins), len(config.CORS.AllowedOrigins))
	}
	for i, origin := range expectedOrigins {
		if config.CORS.AllowedOrigins[i] != origin {
			t.Errorf("Expected origin %s, got %s", origin, config.CORS.AllowedOrigins[i])
		}
	}
	if len(config.CORS.AllowedMethods) != 2 {
		t.Errorf("Expected 2 methods, got %d", len(config.CORS.AllowedMethods))
	}
	if config.CORS.AllowCredentials != false {
		t.Errorf("Expected allow credentials false, got %v", config.CORS.AllowCredentials)
	}
	if config.CORS.MaxAge != 7200 {
		t.Errorf("Expected max age 7200, got %d", config.CORS.MaxAge)
	}

	// Defaults (no * → safe with credentials)
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	os.Unsetenv("CORS_ALLOWED_METHODS")
	os.Unsetenv("CORS_ALLOWED_HEADERS")
	os.Unsetenv("CORS_ALLOW_CREDENTIALS")
	os.Unsetenv("CORS_MAX_AGE")

	config, err = Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(config.CORS.AllowedOrigins) == 0 {
		t.Error("Expected at least one default CORS origin")
	}
	for _, o := range config.CORS.AllowedOrigins {
		if o == "*" {
			t.Error("Default CORS origins must not contain * (invalid with AllowCredentials=true)")
		}
	}
	if config.CORS.AllowCredentials != true {
		t.Errorf("Expected default allow credentials true, got %v", config.CORS.AllowCredentials)
	}
	if config.CORS.MaxAge != 3600 {
		t.Errorf("Expected default max age 3600, got %d", config.CORS.MaxAge)
	}
}

// restoreEnv sets or unsets an env var based on whether the saved value is non-empty.
func restoreEnv(key, saved string) {
	if saved != "" {
		os.Setenv(key, saved)
	} else {
		os.Unsetenv(key)
	}
}
