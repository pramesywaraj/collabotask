package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigLoad(t *testing.T) {
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
	}()

	// Unset environment variables to test defaults
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("SERVER_HOST")
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
}

func TestConfigValidation(t *testing.T) {
	// Save original
	originalDBName := os.Getenv("DB_NAME")
	originalDBUser := os.Getenv("DB_USER")

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
}

func TestConfigDurationParsing(t *testing.T) {
	// Save original
	originalTimeout := os.Getenv("SERVER_TIMEOUT")

	// Clean up
	defer func() {
		if originalTimeout != "" {
			os.Setenv("SERVER_TIMEOUT", originalTimeout)
		} else {
			os.Unsetenv("SERVER_TIMEOUT")
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
}
