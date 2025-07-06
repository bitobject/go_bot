package handlers

import (
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthCheck возвращает статус здоровья сервиса
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	status := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "goooo-bot",
	}

	// Проверяем подключение к БД
	sqlDB, err := h.db.DB()
	if err != nil {
		status["status"] = "unhealthy"
		status["database"] = "error"
		status["error"] = err.Error()
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	// Проверяем ping к БД
	if err := sqlDB.Ping(); err != nil {
		status["status"] = "unhealthy"
		status["database"] = "disconnected"
		status["error"] = err.Error()
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	status["database"] = "connected"
	c.JSON(http.StatusOK, status)
}

// ReadinessCheck проверяет готовность сервиса к работе
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	status := gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC(),
		"service":   "goooo-bot",
	}

	// Проверяем подключение к БД
	sqlDB, err := h.db.DB()
	if err != nil {
		status["status"] = "not ready"
		status["database"] = "error"
		status["error"] = err.Error()
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	// Проверяем ping к БД
	if err := sqlDB.Ping(); err != nil {
		status["status"] = "not ready"
		status["database"] = "disconnected"
		status["error"] = err.Error()
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	status["database"] = "ready"
	c.JSON(http.StatusOK, status)
} 