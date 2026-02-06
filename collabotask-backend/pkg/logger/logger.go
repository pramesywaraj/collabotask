package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	*zerolog.Logger
}

type Config struct {
	Level  string
	Format string
}

func New(config Config) *Logger {
	level := parseLogLevel(config.Level)
	zerolog.SetGlobalLevel(level)

	var logger zerolog.Logger
	if config.Format == "json" {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
		logger = zerolog.New(output).With().Timestamp().Stack().Logger()
	}

	return &Logger{
		Logger: &logger,
	}
}

func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "trace":
		return zerolog.TraceLevel
	default:
		return zerolog.InfoLevel
	}
}

func (l *Logger) Debug(msg string) {
	l.Logger.Debug().Msg(msg)
}

func (l *Logger) Info(msg string) {
	l.Logger.Info().Msg(msg)
}

func (l *Logger) Warn(msg string) {
	l.Logger.Warn().Msg(msg)
}

func (l *Logger) Error(msg string) {
	l.Logger.Error().Msg(msg)
}

func (l *Logger) ErrorWithErr(msg string, err error) {
	l.Logger.Error().Err(err).Msg(msg)
}

func (l *Logger) Fatal(msg string) {
	l.Logger.Fatal().Msg(msg)
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
	logger := l.Logger.With().Interface(key, value).Logger()
	return &Logger{Logger: &logger}
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.Logger.With()
	for key, value := range fields {
		ctx = ctx.Interface(key, value)
	}
	logger := ctx.Logger()
	return &Logger{Logger: &logger}
}

func (l *Logger) WithError(err error) *Logger {
	logger := l.Logger.With().Err(err).Logger()
	return &Logger{Logger: &logger}
}
