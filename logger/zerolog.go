package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	// Read the DEBUG environment variable to determine the log level
	debugEnv := strings.ToLower(os.Getenv("DEBUG"))
	level := zerolog.InfoLevel

	if debugEnv == "true" {
		level = zerolog.DebugLevel
	}

	// Create a new logger with the desired level
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    false,
	}).Level(level).With().Timestamp().Logger()

	// Assign the new logger to the package-level logger
	log.Logger = logger
}

func Debug(msg string, args ...interface{}) {
	log.Debug().Msgf(msg, args...)
}

func Info(msg string, args ...interface{}) {
	log.Info().Msgf(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	log.Warn().Msgf(msg, args...)
}

func Error(msg string, args ...interface{}) {
	log.Error().Msgf(msg, args...)
}

func Fatal(msg string, args ...interface{}) {
	log.Fatal().Msgf(msg, args...)
}
