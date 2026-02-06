package logger

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestLoggerCreation(t *testing.T) {
	log := New(Config{
		Level:  "debug",
		Format: "console",
	})

	if log == nil {
		t.Fatal("Logger should not be nil")
	}

	log.Info("Test info message")
	log.Debug("Test debug message")
	log.Warn("Test warning message")
}

func TestLoggerWithFields(t *testing.T) {
	log := New(Config{
		Level:  "info",
		Format: "json",
	})

	log.WithField("test_key", "test_value").Info("Test message with field")
	log.WithFields(map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}).Info("Test message with fields")
}

func TestLogLevelParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected zerolog.Level
	}{
		{"debug", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"invalid", zerolog.InfoLevel},
	}

	for _, tt := range tests {
		result := parseLogLevel(tt.input)
		if result != tt.expected {
			t.Errorf("parseLogLevel(%s) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
