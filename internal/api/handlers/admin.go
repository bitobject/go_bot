package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"go-bot/internal/api/apierror"
	"go-bot/internal/auth"
	"go-bot/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var validate = validator.New()

// AdminHandler handles admin-related API endpoints.
type AdminHandler struct {
	adminService services.AdminServiceInterface
	logger       *slog.Logger
	jwtSecret    string
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(adminService services.AdminServiceInterface, logger *slog.Logger, jwtSecret string) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
		logger:       logger,
		jwtSecret:    jwtSecret,
	}
}

// LoginRequest represents the request body for admin login.
type LoginRequest struct {
	Login    string `json:"login" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=8"`
}

// Login handles admin authentication and returns a JWT.
func (h *AdminHandler) Login(c *gin.Context) error {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return apierror.New(http.StatusBadRequest, "invalid request body: "+err.Error())
	}

	if err := validate.Struct(req); err != nil {
		return apierror.New(http.StatusBadRequest, "validation failed: "+err.Error())
	}

	// Authenticate admin
	admin, err := h.adminService.Authenticate(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		if err.Error() == "invalid credentials" {
			return apierror.New(http.StatusUnauthorized, "invalid credentials")
		}
		return err // Internal server error
	}

	// Generate JWT
	token, err := auth.GenerateToken(admin, h.jwtSecret, 24*time.Hour)
	if err != nil {
		return err // Internal server error
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
	return nil
}

// GetProfile handles retrieving the admin's profile.
func (h *AdminHandler) GetProfile(c *gin.Context) error {
	adminID, exists := c.Get("adminID")
	if !exists {
		return apierror.New(http.StatusUnauthorized, "missing adminID in context")
	}

	id, ok := adminID.(uint)
	if !ok {
		return apierror.New(http.StatusInternalServerError, "invalid admin ID type in token")
	}

	admin, err := h.adminService.GetProfile(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apierror.New(http.StatusNotFound, "admin not found")
		}
		return err // Internal server error
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        admin.ID,
		"login":     admin.Login,
		"is_active": admin.IsActive,
	})
	return nil
}

// ChangePasswordRequest represents the request body for changing password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ChangePassword handles changing the admin's password.
func (h *AdminHandler) ChangePassword(c *gin.Context) error {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return apierror.New(http.StatusBadRequest, "invalid request body: "+err.Error())
	}

	if err := validate.Struct(req); err != nil {
		return apierror.New(http.StatusBadRequest, "validation failed: "+err.Error())
	}

	adminID, exists := c.Get("adminID")
	if !exists {
		return apierror.New(http.StatusUnauthorized, "missing adminID in context")
	}

	id, ok := adminID.(uint)
	if !ok {
		return apierror.New(http.StatusInternalServerError, "invalid admin ID type in token")
	}

	if err := h.adminService.ChangePassword(c.Request.Context(), id, req.OldPassword, req.NewPassword); err != nil {
		if err.Error() == "incorrect old password" {
			return apierror.New(http.StatusUnauthorized, "incorrect old password")
		}
		return err // Internal server error
	}

	c.Status(http.StatusNoContent)
	return nil
}
