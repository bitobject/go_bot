package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goooo/internal/auth"
	"goooo/internal/config"
	"goooo/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestAdminHandler представляет тестовый admin handler
type TestAdminHandler struct {
	*AdminHandler
	db *gorm.DB
}

// setupTestDB создает тестовую базу данных в памяти
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Мигрируем модели
	err = db.AutoMigrate(&database.User{}, &database.Message{}, &database.Admin{})
	require.NoError(t, err)

	return db
}

// setupTestHandler создает тестовый handler
func setupTestHandler(t *testing.T) *TestAdminHandler {
	// Загружаем тестовую конфигурацию
	config.LoadEnv()
	config.Init()

	db := setupTestDB(t)
	handler := NewAdminHandler(db)

	return &TestAdminHandler{
		AdminHandler: handler,
		db:           db,
	}
}

// setupGinContext создает тестовый Gin контекст
func setupGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

// createTestAdmin создает тестового администратора
func (h *TestAdminHandler) createTestAdmin(t *testing.T, login, password string) *database.Admin {
	admin, err := h.adminService.CreateAdmin(login, password)
	require.NoError(t, err)
	return admin
}

func TestAdminHandler_Login_Success(t *testing.T) {
	handler := setupTestHandler(t)
	defer handler.db.Migrator().DropTable(&database.Admin{})

	// Создаем тестового администратора
	admin := handler.createTestAdmin(t, "testadmin", "testpassword123")

	// Создаем запрос
	loginReq := LoginRequest{
		Login:    "testadmin",
		Password: "testpassword123",
	}
	body, _ := json.Marshal(loginReq)

	// Создаем HTTP запрос
	req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Создаем Gin контекст
	c, w := setupGinContext()
	c.Request = req

	// Выполняем запрос
	handler.Login(c)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)

	var response LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Проверяем структуру ответа
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, admin.ID, response.Admin.ID)
	assert.Equal(t, admin.Login, response.Admin.Login)

	// Проверяем валидность JWT токена
	claims, err := auth.ValidateToken(response.Token)
	require.NoError(t, err)
	assert.Equal(t, admin.ID, claims.AdminID)
	assert.Equal(t, admin.Login, claims.Login)
}

func TestAdminHandler_Login_InvalidCredentials(t *testing.T) {
	handler := setupTestHandler(t)
	defer handler.db.Migrator().DropTable(&database.Admin{})

	// Создаем тестового администратора
	handler.createTestAdmin(t, "testadmin", "testpassword123")

	testCases := []struct {
		name     string
		login    string
		password string
	}{
		{
			name:     "wrong_password",
			login:    "testadmin",
			password: "wrongpassword",
		},
		{
			name:     "wrong_login",
			login:    "wrongadmin",
			password: "testpassword123",
		},
		{
			name:     "both_wrong",
			login:    "wrongadmin",
			password: "wrongpassword",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			loginReq := LoginRequest{
				Login:    tc.login,
				Password: tc.password,
			}
			body, _ := json.Marshal(loginReq)

			req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			c, w := setupGinContext()
			c.Request = req

			handler.Login(c)

			// Проверяем, что возвращается 401 и правильная ошибка
			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "invalid login or password", response["error"])
		})
	}
}

func TestAdminHandler_Login_InvalidRequest(t *testing.T) {
	handler := setupTestHandler(t)
	defer handler.db.Migrator().DropTable(&database.Admin{})

	testCases := []struct {
		name        string
		requestBody interface{}
		expectedMsg string
	}{
		{
			name:        "missing_login",
			requestBody: map[string]string{"password": "test"},
			expectedMsg: "login and password are required",
		},
		{
			name:        "missing_password",
			requestBody: map[string]string{"login": "test"},
			expectedMsg: "login and password are required",
		},
		{
			name:        "empty_json",
			requestBody: map[string]string{},
			expectedMsg: "login and password are required",
		},
		{
			name:        "invalid_json",
			requestBody: "invalid json",
			expectedMsg: "login and password are required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tc.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tc.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			c, w := setupGinContext()
			c.Request = req

			handler.Login(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedMsg, response["error"])
		})
	}
}

func TestAdminHandler_Login_AccountLocked(t *testing.T) {
	handler := setupTestHandler(t)
	defer handler.db.Migrator().DropTable(&database.Admin{})

	// Создаем тестового администратора
	admin := handler.createTestAdmin(t, "testadmin", "testpassword123")

	// Блокируем аккаунт, установив failed_login_attempts = 5 и locked_until в будущем
	lockoutTime := time.Now().Add(15 * time.Minute)
	err := handler.db.Model(&database.Admin{}).Where("id = ?", admin.ID).
		Updates(map[string]interface{}{
			"failed_login_attempts": 5,
			"locked_until":          lockoutTime,
		}).Error
	require.NoError(t, err)

	// Пытаемся войти с правильными данными
	loginReq := LoginRequest{
		Login:    "testadmin",
		Password: "testpassword123",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	c, w := setupGinContext()
	c.Request = req

	handler.Login(c)

	// Должны получить 401, так как аккаунт заблокирован
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid login or password", response["error"])
}

func TestAdminHandler_Login_AccountInactive(t *testing.T) {
	handler := setupTestHandler(t)
	defer handler.db.Migrator().DropTable(&database.Admin{})

	// Создаем тестового администратора
	admin := handler.createTestAdmin(t, "testadmin", "testpassword123")

	// Деактивируем аккаунт
	err := handler.db.Model(&database.Admin{}).Where("id = ?", admin.ID).
		Update("is_active", false).Error
	require.NoError(t, err)

	// Пытаемся войти
	loginReq := LoginRequest{
		Login:    "testadmin",
		Password: "testpassword123",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	c, w := setupGinContext()
	c.Request = req

	handler.Login(c)

	// Должны получить 401, так как аккаунт неактивен
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid login or password", response["error"])
}

// Benchmark тесты для производительности
func BenchmarkAdminHandler_Login_Success(b *testing.B) {
	handler := setupTestHandler(&testing.T{})
	defer handler.db.Migrator().DropTable(&database.Admin{})

	// Создаем тестового администратора
	handler.createTestAdmin(&testing.T{}, "testadmin", "testpassword123")

	loginReq := LoginRequest{
		Login:    "testadmin",
		Password: "testpassword123",
	}
	body, _ := json.Marshal(loginReq)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		c, w := setupGinContext()
		c.Request = req

		handler.Login(c)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
} 