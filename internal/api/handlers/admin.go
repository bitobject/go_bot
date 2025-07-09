package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"go-bot/internal/api/apierror"
	"go-bot/internal/auth"
	"go-bot/internal/services"
	"go-bot/internal/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	adminService *services.AdminService
	jwtSecretKey string
	jwtExpiresIn time.Duration
	logger       *slog.Logger
}

func NewAdminHandler(db *gorm.DB, logger *slog.Logger, jwtSecretKey string, jwtExpiresIn time.Duration) *AdminHandler {
	adminService := services.NewAdminService(db)
	return &AdminHandler{
		adminService: adminService,
		jwtSecretKey: jwtSecretKey,
		jwtExpiresIn: jwtExpiresIn,
		logger:       logger,
	}
}

// LoginRequest структура для запроса входа
type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse структура для ответа при входе
type LoginResponse struct {
	Token string `json:"token"`
	Admin struct {
		ID    uint   `json:"id"`
		Login string `json:"login"`
	} `json:"admin"`
}

// Login обрабатывает вход администратора
func (h *AdminHandler) Login(c *gin.Context) error {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return apierror.New(http.StatusBadRequest, "invalid request body")
	}

	// Аутентификация
	admin, err := h.adminService.AuthenticateAdmin(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apierror.New(http.StatusUnauthorized, "invalid credentials")
		}
		// Возвращаем внутреннюю ошибку, которую middleware обработает как 500
		return err
	}

	// Генерация JWT токена
	token, err := auth.GenerateToken(admin, h.jwtSecretKey, h.jwtExpiresIn)
	if err != nil {
		// Возвращаем внутреннюю ошибку
		return err
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
	return nil
}

// GetProfile возвращает профиль текущего администратора
func (h *AdminHandler) GetProfile(c *gin.Context) error {
	adminID, exists := c.Get("admin_id")
	if !exists {
		return apierror.New(http.StatusUnauthorized, "unauthorized")
	}

	id, ok := adminID.(uint)
	if !ok {
		return apierror.New(http.StatusInternalServerError, "invalid admin ID type in token")
	}

	admin, err := h.adminService.GetAdminByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apierror.New(http.StatusNotFound, "admin not found")
		}
		return err // Внутренняя ошибка
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    admin.ID,
		"login": admin.Login,
	})
	return nil
}

// ChangePassword изменяет пароль администратора
func (h *AdminHandler) ChangePassword(c *gin.Context) error {
	adminID, exists := c.Get("admin_id")
	if !exists {
		return apierror.New(http.StatusUnauthorized, "unauthorized")
	}

	id, ok := adminID.(uint)
	if !ok {
		return apierror.New(http.StatusInternalServerError, "invalid admin ID type in token")
	}

	type ChangePasswordRequest struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return apierror.New(http.StatusBadRequest, err.Error())
	}

	// Проверяем текущий пароль
	admin, err := h.adminService.GetAdminByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apierror.New(http.StatusNotFound, "admin not found")
		}
		return err // Внутренняя ошибка
	}

	if !auth.CheckPasswordHash(req.CurrentPassword, admin.Password) {
		return apierror.New(http.StatusUnauthorized, "invalid current password")
	}

	// Обновляем пароль
	if err := h.adminService.UpdateAdminPassword(id, req.NewPassword); err != nil {
		return err // Внутренняя ошибка
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
	return nil
}