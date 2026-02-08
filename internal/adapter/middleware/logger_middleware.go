package middleware

import (
	"time"

	"github.com/dhanuprys/infrantery-backend-go/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LoggerMiddleware logs HTTP requests with structured logging
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Get user ID if authenticated
		userID, _ := c.Get("user_id")
		userIDStr := ""
		if userID != nil {
			userIDStr = userID.(string)
		}

		// Log request
		event := logger.Logger.Info()
		if statusCode >= 500 {
			event = logger.Logger.Error()
		} else if statusCode >= 400 {
			event = logger.Logger.Warn()
		}

		logEvent := event.
			Str("request_id", requestID).
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Dur("duration", duration).
			Str("ip", c.ClientIP())

		if userIDStr != "" {
			logEvent = logEvent.Str("user_id", logger.SanitizeUserID(userIDStr))
		}

		// Add user agent for non-GET requests
		if method != "GET" {
			logEvent = logEvent.Str("user_agent", c.Request.UserAgent())
		}

		logEvent.Msg("HTTP request")
	}
}
