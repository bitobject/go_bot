package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

const (
	// loggerKey - ключ для хранения логгера в контексте Gin.
	loggerKey = "slog-logger"
)

// LoggerInjector - это middleware, которое добавляет slog.Logger в контекст Gin.
// Это позволяет другим middleware и хендлерам получать доступ к логгеру.
func LoggerInjector(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(loggerKey, logger)
		c.Next()
	}
}
