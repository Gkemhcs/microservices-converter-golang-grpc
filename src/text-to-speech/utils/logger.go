package utils

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// NewLogger creates and returns a Logrus logger instance
func NewLogger() *logrus.Logger {
	logger := logrus.New()
	isDocker := os.Getenv("DOCKER_ENVIRONMENT") == "true"
	if isDocker {
		// Open or create a log file
		logFile, err := os.OpenFile(os.Getenv("LOG_DIR")+"/file-uploader.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.Fatal("Failed to open log file:", err)
		}
		logger.SetOutput(logFile)
	} else {
		logger.SetOutput(os.Stdout)
	}

	// Set the log format to JSON (for structured logging)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	// Set the logging level (e.g., Debug, Info, Warn, Error)
	logger.SetLevel(logrus.DebugLevel)

	return logger
}
