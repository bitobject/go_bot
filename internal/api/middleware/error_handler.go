package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"go-bot/internal/api/apierror"

	"github.com/gin-gonic/gin"
)

// HandlerFunc определяет тип обработчика, который может возвращать ошибку.
// Это позволяет нам создавать обработчики с более чистой сигнатурой.
type HandlerFunc func(c *gin.Context) error

const loggerKey = "logger"

// LoggerInjector является middleware, которое внедряет slog.Logger в контекст Gin.
func LoggerInjector(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(loggerKey, logger)
		c.Next()
	}
}

// ErrorHandler является middleware, которое преобразует возвращаемые ошибки
// из обработчиков в стандартные HTTP-ответы.
func ErrorHandler(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := handler(c)
		if err == nil {
			return
		}

		// Пытаемся преобразовать ошибку в нашу кастомную APIError
		var apiErr apierror.APIError
		if errors.As(err, &apiErr) {
			// Если это известная ошибка API, возвращаем ее статус и сообщение
			c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
			return
		}

		// Если это любая другая (неожиданная) ошибка, логируем ее как внутреннюю
		// и возвращаем стандартный ответ 500, чтобы не раскрывать детали реализации.
		logger, _ := c.Get(loggerKey)
		if l, ok := logger.(*slog.Logger); ok {
			l.Error("unhandled internal server error", "error", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}
}
