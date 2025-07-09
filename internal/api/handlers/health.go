package handlers

import (
	"net/http"

	"go-bot/internal/services"

	"github.com/gin-gonic/gin"
)

// HealthHandler обрабатывает эндпоинты, связанные со здоровьем сервиса.
type HealthHandler struct {
	service services.HealthServiceInterface
}

// NewHealthHandler создает новый HealthHandler.
func NewHealthHandler(s services.HealthServiceInterface) *HealthHandler {
	return &HealthHandler{service: s}
}

// HealthCheck проверяет базовую работоспособность сервиса.
// Возвращает 200 OK, если сервис запущен.
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ReadinessCheck проверяет, готов ли сервис принимать трафик.
// В данном случае, он проверяет подключение к базе данных.
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	if err := h.service.CheckDB(c.Request.Context()); err != nil {
		// Не логируем ошибку здесь, так как это может зафлудить логи при проблемах с БД.
		// Просто отвечаем, что сервис не готов.
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"reason": "database connection failed",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
 