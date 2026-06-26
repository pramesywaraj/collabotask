package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigLoad(t *testing.T) {
	// Save original environment
	originalPort := os.Getenv("SERVER_PORT")
	originalDBName := os.Getenv("DB_NAME")
	originalDBUser := os.Getenv("DB_USER")

	defer func() {
		if originalPort != "" {
			os.Setenv("SERVER_PORT", originalPort)
		} else {
			os.Unsetenv("SERVER_PORT")
		}
		if originalDBName != "" {
			os.Setenv("DB_NAME", originalDBName)
		} else {
			os.Unsetenv("DB_NAME")
		}
		if originalDBUser != "" {
			os.Setenv("DB_USER", originalDBUser)
		} else {
			os.Unsetenv("DB_USER")
		}
	}()

	os.Setenv("SERVER_PORT", "3000")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_USER", "testuser")

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test that custom values are loaded
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
	// Save original environment
	originalPort := os.Getenv("SERVER_PORT")
	originalHost := os.Getenv("SERVER_HOST")
	originalAppEnv := os.Getenv("APP_ENV")
	originalAppName := os.Getenv("APP_NAME")

	// Clean up
	defer func() {
		if originalPort != "" {
			os.Setenv("SERVER_PORT", originalPort)
		} else {
			os.Unsetenv("SERVER_PORT")
		}
		if originalHost != "" {
			os.Setenv("SERVER_HOST", originalHost)
		} else {
			os.Unsetenv("SERVER_HOST")
		}
		if originalAppEnv != "" {
			os.Setenv("APP_ENV", originalAppEnv)
		} else {
			os.Unsetenv("APP_ENV")
		}
		if originalAppName != "" {
			os.Setenv("APP_NAME", originalAppName)
		} else {
			os.Unsetenv("APP_NAME")
		}
	}()

	// Unset environment variables to test defaults
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("SERVER_HOST")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("APP_NAME")
	os.Setenv("DB_NAME", "testdb")   // Required field
	os.Setenv("DB_USER", "testuser") // Required field

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test defaults
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
	// Save original
	originalDBName := os.Getenv("DB_NAME")
	originalDBUser := os.Getenv("DB_USER")
	originalAppEnv := os.Getenv("APP_ENV")

	// Clean up
	defer func() {
		if originalDBName != "" {
			os.Setenv("DB_NAME", originalDBName)
		} else {
			os.Unsetenv("DB_NAME")
		}
		if originalDBUser != "" {
			os.Setenv("DB_USER", originalDBUser)
		} else {
			os.Unsetenv("DB_USER")
		}
		if originalAppEnv != "" {
			os.Setenv("APP_ENV", originalAppEnv)
		} else {
			os.Unsetenv("APP_ENV")
		}
	}()

	// Test missing DB_NAME
	os.Unsetenv("DB_NAME")
	os.Setenv("DB_USER", "testuser")

	_, err := Load()
	if err == nil {
		t.Error("Expected error when DB_NAME is missing, got nil")
	}
	if err != nil && err.Error() != "config validation failed: DB_NAME is required" {
		t.Errorf("Expected 'DB_NAME is required' error, got: %v", err)
	}

	// Test missing DB_USER
	os.Setenv("DB_NAME", "testdb")
	os.Unsetenv("DB_USER")

	_, err = Load()
	if err == nil {
		t.Error("Expected error when DB_USER is missing, got nil")
	}
	if err != nil && err.Error() != "config validation failed: DB_USER is required" {
		t.Errorf("Expected 'DB_USER is required' error, got: %v", err)
	}

	// Test invalid APP_ENV
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

func TestConfigDurationParsing(t *testing.T) {
	// Save original
	originalTimeout := os.Getenv("SERVER_TIMEOUT")
	originalMaxConnLifetime := os.Getenv("DB_MAX_CONN_LIFETIME")

	// Clean up
	defer func() {
		if originalTimeout != "" {
			os.Setenv("SERVER_TIMEOUT", originalTimeout)
		} else {
			os.Unsetenv("SERVER_TIMEOUT")
		}
		if originalMaxConnLifetime != "" {
			os.Setenv("DB_MAX_CONN_LIFETIME", originalMaxConnLifetime)
		} else {
			os.Unsetenv("DB_MAX_CONN_LIFETIME")
		}
	}()

	// Set required fields
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_USER", "testuser")

	// Test custom timeout
	os.Setenv("SERVER_TIMEOUT", "60s")

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	expectedTimeout := 60 * time.Second
	if config.Server.Timeout != expectedTimeout {
		t.Errorf("Expected timeout 60s, got %v", config.Server.Timeout)
	}

	// Test default timeout
	os.Unsetenv("SERVER_TIMEOUT")

	config, err = Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	expectedTimeout = 30 * time.Second
	if config.Server.Timeout != expectedTimeout {
		t.Errorf("Expected default timeout 30s, got %v", config.Server.Timeout)
	}

	// Test database connection lifetime
	os.Setenv("DB_MAX_CONN_LIFETIME", "1h")

	config, err = Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	expectedLifetime := 1 * time.Hour
	if config.Database.MaxConnLifetime != expectedLifetime {
		t.Errorf("Expected max conn lifetime 1h, got %v", config.Database.MaxConnLifetime)
	}
}

func TestAppConfig(t *testing.T) {
	// Save original
	originalAppEnv := os.Getenv("APP_ENV")
	originalAppName := os.Getenv("APP_NAME")
	originalAppVersion := os.Getenv("APP_VERSION")
	originalAppDebug := os.Getenv("APP_DEBUG")

	defer func() {
		if originalAppEnv != "" {
			os.Setenv("APP_ENV", originalAppEnv)
		} else {
			os.Unsetenv("APP_ENV")
		}
		if originalAppName != "" {
			os.Setenv("APP_NAME", originalAppName)
		} else {
			os.Unsetenv("APP_NAME")
		}
		if originalAppVersion != "" {
			os.Setenv("APP_VERSION", originalAppVersion)
		} else {
			os.Unsetenv("APP_VERSION")
		}
		if originalAppDebug != "" {
			os.Setenv("APP_DEBUG", originalAppDebug)
		} else {
			os.Unsetenv("APP_DEBUG")
		}
	}()

	// Set required fields
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_USER", "testuser")

	// Test custom app config
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

	if config.App.Debug != true {
		t.Errorf("Expected debug true, got %v", config.App.Debug)
	}
}

func TestDatabaseConfig(t *testing.T) {
	// Save original
	originalDBHost := os.Getenv("DB_HOST")
	originalDBPort := os.Getenv("DB_PORT")
	originalDBPassword := os.Getenv("DB_PASSWORD")
	originalDBSSLMode := os.Getenv("DB_SSLMODE")
	originalMaxConns := os.Getenv("DB_MAX_CONNS")
	originalMinConns := os.Getenv("DB_MIN_CONNS")

	defer func() {
		if originalDBHost != "" {
			os.Setenv("DB_HOST", originalDBHost)
		} else {
			os.Unsetenv("DB_HOST")
		}
		if originalDBPort != "" {
			os.Setenv("DB_PORT", originalDBPort)
		} else {
			os.Unsetenv("DB_PORT")
		}
		if originalDBPassword != "" {
			os.Setenv("DB_PASSWORD", originalDBPassword)
		} else {
			os.Unsetenv("DB_PASSWORD")
		}
		if originalDBSSLMode != "" {
			os.Setenv("DB_SSLMODE", originalDBSSLMode)
		} else {
			os.Unsetenv("DB_SSLMODE")
		}
		if originalMaxConns != "" {
			os.Setenv("DB_MAX_CONNS", originalMaxConns)
		} else {
			os.Unsetenv("DB_MAX_CONNS")
		}
		if originalMinConns != "" {
			os.Setenv("DB_MIN_CONNS", originalMinConns)
		} else {
			os.Unsetenv("DB_MIN_CONNS")
		}
	}()

	// Set required fields
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_USER", "testuser")

	// Test custom database config
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

	// Test defaults
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
	// Save original
	originalLogLevel := os.Getenv("LOG_LEVEL")
	originalLogFormat := os.Getenv("LOG_FORMAT")

	defer func() {
		if originalLogLevel != "" {
			os.Setenv("LOG_LEVEL", originalLogLevel)
		} else {
			os.Unsetenv("LOG_LEVEL")
		}
		if originalLogFormat != "" {
			os.Setenv("LOG_FORMAT", originalLogFormat)
		} else {
			os.Unsetenv("LOG_FORMAT")
		}
	}()

	// Set required fields
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_USER", "testuser")

	// Test custom log config
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

	// Test defaults
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
	// Save original
	originalCORSOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	originalCORSMethods := os.Getenv("CORS_ALLOWED_METHODS")
	originalCORSHeaders := os.Getenv("CORS_ALLOWED_HEADERS")
	originalCORSAllowCreds := os.Getenv("CORS_ALLOW_CREDENTIALS")
	originalCORSMaxAge := os.Getenv("CORS_MAX_AGE")

	defer func() {
		if originalCORSOrigins != "" {
			os.Setenv("CORS_ALLOWED_ORIGINS", originalCORSOrigins)
		} else {
			os.Unsetenv("CORS_ALLOWED_ORIGINS")
		}
		if originalCORSMethods != "" {
			os.Setenv("CORS_ALLOWED_METHODS", originalCORSMethods)
		} else {
			os.Unsetenv("CORS_ALLOWED_METHODS")
		}
		if originalCORSHeaders != "" {
			os.Setenv("CORS_ALLOWED_HEADERS", originalCORSHeaders)
		} else {
			os.Unsetenv("CORS_ALLOWED_HEADERS")
		}
		if originalCORSAllowCreds != "" {
			os.Setenv("CORS_ALLOW_CREDENTIALS", originalCORSAllowCreds)
		} else {
			os.Unsetenv("CORS_ALLOW_CREDENTIALS")
		}
		if originalCORSMaxAge != "" {
			os.Setenv("CORS_MAX_AGE", originalCORSMaxAge)
		} else {
			os.Unsetenv("CORS_MAX_AGE")
		}
	}()

	// Set required fields
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_USER", "testuser")

	// Test custom CORS config
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

	expectedMethods := []string{"GET", "POST"}
	if len(config.CORS.AllowedMethods) != len(expectedMethods) {
		t.Errorf("Expected %d methods, got %d", len(expectedMethods), len(config.CORS.AllowedMethods))
	}

	if config.CORS.AllowCredentials != false {
		t.Errorf("Expected allow credentials false, got %v", config.CORS.AllowCredentials)
	}

	if config.CORS.MaxAge != 7200 {
		t.Errorf("Expected max age 7200, got %d", config.CORS.MaxAge)
	}

	// Test defaults
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	os.Unsetenv("CORS_ALLOWED_METHODS")
	os.Unsetenv("CORS_ALLOWED_HEADERS")
	os.Unsetenv("CORS_ALLOW_CREDENTIALS")
	os.Unsetenv("CORS_MAX_AGE")

	config, err = Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(config.CORS.AllowedOrigins) != 1 || config.CORS.AllowedOrigins[0] != "*" {
		t.Errorf("Expected default origins [*], got %v", config.CORS.AllowedOrigins)
	}

	if config.CORS.AllowCredentials != true {
		t.Errorf("Expected default allow credentials true, got %v", config.CORS.AllowCredentials)
	}

	if config.CORS.MaxAge != 3600 {
		t.Errorf("Expected default max age 3600, got %d", config.CORS.MaxAge)
	}
}
