package middleware

import (
	"net/http"
	"strings"

	"go-bot/internal/auth"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет JWT токен и добавляет claims в контекст
func AuthMiddleware(jwtSecretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Проверяем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := auth.ValidateToken(tokenString, jwtSecretKey)
		if err != nil {
			switch err {
			case auth.ErrExpiredToken:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			case auth.ErrInvalidToken:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			default:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
			}
			c.Abort()
			return
		}

		// Добавляем claims в контекст
		c.Set("admin_id", claims.AdminID)
		c.Set("admin_login", claims.Login)
		c.Set("claims", claims)

		c.Next()
	}
}

// GetAdminID извлекает ID администратора из контекста
func GetAdminID(c *gin.Context) (uint, bool) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		return 0, false
	}
	if id, ok := adminID.(uint); ok {
		return id, true
	}
	return 0, false
}

// GetAdminLogin извлекает логин администратора из контекста
func GetAdminLogin(c *gin.Context) (string, bool) {
	adminLogin, exists := c.Get("admin_login")
	if !exists {
		return "", false
	}
	if login, ok := adminLogin.(string); ok {
		return login, true
	}
	return "", false
} 