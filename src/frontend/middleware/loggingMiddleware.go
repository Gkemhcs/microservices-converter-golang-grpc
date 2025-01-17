package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func LoggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log the start of the request
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		logger.Infof("Incoming request: %s %s", method, path)

		// Process the request
		c.Next()

		// Calculate the time taken
		duration := time.Since(startTime)

		// Get the status code
		statusCode := c.Writer.Status()

		// Determine success or failure
		if statusCode >= 200 && statusCode < 300 {
			logger.Infof("Request succeeded: %s %s [Status: %d, Duration: %v]", method, path, statusCode, duration)
		} else {
			logger.Errorf("Request failed: %s %s [Status: %d, Duration: %v]", method, path, statusCode, duration)
		}
	}
}
