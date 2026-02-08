package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

// Init initializes the global logger with the specified level and environment
func Init(logLevel, environment string) {
	// Set log level
	level := parseLogLevel(logLevel)
	zerolog.SetGlobalLevel(level)

	// Configure based on environment
	if environment == "development" {
		// Pretty console output for development
		Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Caller().Logger()
	} else {
		// JSON output for production
		Logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	}

	// Set global logger
	log.Logger = Logger
}

// parseLogLevel converts string log level to zerolog.Level
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Info returns a logger for info level
func Info() *zerolog.Event {
	return Logger.Info()
}

// Debug returns a logger for debug level
func Debug() *zerolog.Event {
	return Logger.Debug()
}

// Warn returns a logger for warn level
func Warn() *zerolog.Event {
	return Logger.Warn()
}

// Error returns a logger for error level
func Error() *zerolog.Event {
	return Logger.Error()
}

// Fatal returns a logger for fatal level
func Fatal() *zerolog.Event {
	return Logger.Fatal()
}

// MaskEmail masks email for privacy (keeps first char and domain)
func MaskEmail(email string) string {
	if email == "" {
		return "***"
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}
	if len(parts[0]) == 0 {
		return "***@" + parts[1]
	}
	return parts[0][:1] + "***@" + parts[1]
}

// SanitizeUserID returns a safe user ID for logging (first 8 chars)
func SanitizeUserID(userID string) string {
	if len(userID) > 8 {
		return userID[:8] + "..."
	}
	return userID
}
