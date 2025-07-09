package apierror

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIError представляет собой ошибку API с кодом состояния и сообщением.
// Она удовлетворяет стандартному интерфейсу error.
type APIError struct {
	StatusCode int
	Message    string
}

// Error удовлетворяет интерфейсу error.
func (e APIError) Error() string {
	return e.Message
}

// New создает новую ошибку API.
func New(statusCode int, message string) APIError {
	return APIError{StatusCode: statusCode, Message: message}
}

// Newf создает новую ошибку API с форматированием строки.
func Newf(statusCode int, format string, a ...interface{}) APIError {
	return APIError{StatusCode: statusCode, Message: fmt.Sprintf(format, a...)}
}

// ErrorWrapper оборачивает хендлер, возвращающий ошибку, в стандартный gin.HandlerFunc.
// Это позволяет централизованно обрабатывать ошибки API.
func ErrorWrapper(handler func(c *gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := handler(c)
		if err == nil {
			return
		}

		var apiErr APIError
		if errors.As(err, &apiErr) {
			c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
			return
		}

		slog.Error("Unhandled API error", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}
}
