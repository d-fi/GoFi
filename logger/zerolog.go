package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Initialize(logLevel string) {
	level := zerolog.DebugLevel

	switch logLevel {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	default:
		log.Warn().Msgf("Invalid log level '%s', defaulting to Debug level", logLevel)
	}

	zerolog.SetGlobalLevel(level)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
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
