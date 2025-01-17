package utils

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// NewLogger creates and returns a Logrus logger instance
func NewLogger() *logrus.Logger {
	logger := logrus.New()

	// Set the output to standard output (console)
	logger.SetOutput(os.Stdout)

	// Set the log format to JSON (for structured logging)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	// Set the logging level (e.g., Debug, Info, Warn, Error)
	logger.SetLevel(logrus.DebugLevel)

	return logger.WithField("service", "text-to-speech-converter").Logger
}
