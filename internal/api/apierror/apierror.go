package apierror

import "fmt"

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
