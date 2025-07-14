package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// SlogLogger is a middleware that logs requests using the slog logger from the context.
// It logs the method, path, status, latency, and client IP of each request.
func SlogLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		latency := time.Since(start)

		logger, exists := c.Get(loggerKey) // loggerKey is defined in error_handler.go
		if !exists {
			// This should not happen if LoggerInjector middleware is used correctly
			return
		}

		slogLogger, ok := logger.(*slog.Logger)
		if !ok {
			// This also should not happen
			return
		}

		slogLogger.Info("http request",
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"latency", latency,
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}
