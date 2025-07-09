package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"go-bot/internal/api/apierror"
	"go-bot/internal/api/handlers"
	"go-bot/internal/api/middleware"
	"go-bot/internal/auth"
	"go-bot/internal/database"
	"go-bot/internal/services"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	jwrSecret    = "test-secret-key-that-is-long-enough"
	jwtExpiresIn = 1 * time.Hour
)

// setupTestRouter создает тестовый роутер, handler и мок базы данных
func setupTestRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock, *handlers.AdminHandler) {
	gin.SetMode(gin.TestMode)

	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	// Создаем gorm.DB с моком
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	// Создаем логгер, который пишет в никуда
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	// Создаем handler
	adminHandler := handlers.NewAdminHandler(gormDB, logger, jwrSecret, jwtExpiresIn)

	// Создаем роутер и добавляем middleware и хендлеры
	router := gin.New()
	// Важно: используем наш кастомный ErrorHandler
	router.Use(func(c *gin.Context) {
		middleware.ErrorHandler(func(c *gin.Context) error {
			c.Next()
			// Мы не можем вернуть ошибку из c.Next(), поэтому возвращаем nil
			// Ошибки будут обработаны внутри хендлеров и возвращены как apierror
			return nil
		})(c)
	})

	return router, mock, adminHandler
}

func TestAdminHandler_Login_Success(t *testing.T) {
	router, mock, adminHandler := setupTestRouter(t)
	router.POST("/login", adminHandler.Login)

	// Хеш для "password123"
	hashedPassword, err := auth.HashPassword("password123")
	require.NoError(t, err)

	// Ожидания от мока
	admin := &database.Admin{
		Model:    gorm.Model{ID: 1},
		Login:    "testadmin",
		Password: hashedPassword,
	}
	rows := sqlmock.NewRows([]string{"id", "login", "password"}).
		AddRow(admin.ID, admin.Login, admin.Password)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "admins" WHERE login = $1 AND "admins"."deleted_at" IS NULL ORDER BY "admins"."id" LIMIT 1`)).
		WithArgs("testadmin").
		WillReturnRows(rows)

	// Создаем запрос
	body := bytes.NewBufferString(`{"login":"testadmin","password":"password123"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", body)
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	router.ServeHTTP(w, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response["token"])

	// Проверяем, что все ожидания от мока были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_Login_Failure(t *testing.T) {
	router, mock, adminHandler := setupTestRouter(t)
	router.POST("/login", adminHandler.Login)

	// Ожидания от мока (запись не найдена)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "admins" WHERE login = $1 AND "admins"."deleted_at" IS NULL ORDER BY "admins"."id" LIMIT 1`)).
		WithArgs("wronguser").
		WillReturnError(gorm.ErrRecordNotFound)

	// Создаем запрос
	body := bytes.NewBufferString(`{"login":"wronguser","password":"password123"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", body)
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	router.ServeHTTP(w, req)

	// Проверяем результат
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response apierror.APIError
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid credentials", response.Message)

	// Проверяем, что все ожидания от мока были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_GetProfile_Success(t *testing.T) {
	router, mock, adminHandler := setupTestRouter(t)
	router.GET("/profile", middleware.AuthMiddleware(jwrSecret), adminHandler.GetProfile)

	// Создаем токен
	admin := &database.Admin{Model: gorm.Model{ID: 1}, Login: "testadmin"}
	token, err := auth.GenerateToken(admin, jwrSecret, jwtExpiresIn)
	require.NoError(t, err)

	// Ожидания от мока
	rows := sqlmock.NewRows([]string{"id", "login"}).
		AddRow(admin.ID, admin.Login)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "admins" WHERE "admins"."id" = $1 AND "admins"."deleted_at" IS NULL ORDER BY "admins"."id" LIMIT 1`)).
		WithArgs(admin.ID).
		WillReturnRows(rows)

	// Создаем запрос
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Выполняем запрос
	router.ServeHTTP(w, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(1), response["id"])
	assert.Equal(t, "testadmin", response["login"])

	// Проверяем, что все ожидания от мока были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}
 