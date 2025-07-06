package handlers

import (
	"net/http"

	"goooo/internal/auth"
	"goooo/internal/database"
	"goooo/internal/api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	adminService *database.AdminService
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{
		adminService: database.NewAdminService(db),
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
func (h *AdminHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "login and password are required",
		})
		return
	}

	// Аутентификация
	admin, err := h.adminService.AuthenticateAdmin(req.Login, req.Password)
	if err != nil {
		// Не раскрываем причину ошибки для безопасности
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid login or password",
		})
		return
	}

	// Генерация JWT токена
	token, err := auth.GenerateToken(admin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	// Формирование ответа
	response := LoginResponse{
		Token: token,
	}
	response.Admin.ID = admin.ID
	response.Admin.Login = admin.Login

	c.JSON(http.StatusOK, response)
}

// GetProfile возвращает профиль текущего администратора
func (h *AdminHandler) GetProfile(c *gin.Context) {
	adminID, exists := middleware.GetAdminID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	admin, err := h.adminService.GetAdminByID(adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         admin.ID,
		"login":      admin.Login,
		"is_active":  admin.IsActive,
		"last_login": admin.LastLoginAt,
		"created_at": admin.CreatedAt,
	})
}

// ChangePassword изменяет пароль администратора
func (h *AdminHandler) ChangePassword(c *gin.Context) {
	adminID, exists := middleware.GetAdminID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	type ChangePasswordRequest struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "current_password and new_password (min 8 chars) are required",
		})
		return
	}

	// Проверяем текущий пароль
	admin, err := h.adminService.GetAdminByID(adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin not found"})
		return
	}

	// Аутентифицируем с текущим паролем
	_, err = h.adminService.AuthenticateAdmin(admin.Login, req.CurrentPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid current password",
		})
		return
	}

	// Меняем пароль
	err = h.adminService.UpdateAdminPassword(adminID, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
} 