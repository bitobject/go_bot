package auth

import (
	"errors"
	"time"

	"go-bot/internal/database"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	AdminID uint64 `json:"admin_id"`
	Login   string `json:"login"`
	jwt.RegisteredClaims
}

// GenerateToken создает JWT токен для администратора
func GenerateToken(admin *database.Admin, secretKey string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		AdminID: admin.ID,
		Login:   admin.Login,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "goooo-admin",
			Subject:   admin.Login,
			Audience:  []string{"admin-panel"},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ValidateToken валидирует JWT токен и возвращает claims
func ValidateToken(tokenString string, secretKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// ExtractAdminID извлекает ID администратора из токена
func ExtractAdminID(tokenString string, secretKey string) (uint64, error) {
	claims, err := ValidateToken(tokenString, secretKey)
	if err != nil {
		return 0, err
	}
	return claims.AdminID, nil
}
