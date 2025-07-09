package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-bot/internal/api/handlers"
	"go-bot/internal/config"
	"go-bot/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestServer представляет тестовый сервер
type TestServer struct {
	*Server
	db  *gorm.DB
	cfg *config.Config
}

// setupTestServer создает тестовый сервер с in-memory БД
func setupTestServer(t *testing.T) *TestServer {
	// Загружаем конфигурацию
	cfg := config.Get()

	// Создаем in-memory БД
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Мигрируем модели
	err = db.AutoMigrate(&database.User{}, &database.Message{}, &database.Admin{})
	require.NoError(t, err)

	// Создаем мок бота
	// В интеграционных тестах нам не нужен реальный бот
	bot := &tgbotapi.BotAPI{}

	// Создаем сервер
	server := NewServer(db, bot)

	return &TestServer{
		Server: server,
		db:     db,
		cfg:    cfg,
	}
}

// createTestAdmin создает тестового администратора
func (ts *TestServer) createTestAdmin(t *testing.T, login, password string) (*database.Admin, string) {
	adminService := database.NewAdminService(ts.db)
	admin, err := adminService.CreateAdmin(login, password)
	require.NoError(t, err)

	// Генерируем JWT токен
	expiresIn := time.Duration(ts.cfg.JWTExpiresIn) * time.Hour
	token, err := auth.GenerateToken(admin, ts.cfg.JWTSecretKey, expiresIn)
	require.NoError(t, err)

	return admin, token
}

// makeRequest выполняет HTTP запрос к тестовому серверу
func (ts *TestServer) makeRequest(t *testing.T, method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	
	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	return w
}

func TestAPI_AdminLogin_Integration(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.db.Migrator().DropTable(&database.Admin{})

	// Создаем тестового администратора
	admin := ts.createTestAdmin(t, "testadmin", "testpassword123")

	t.Run("successful_login", func(t *testing.T) {
		loginReq := map[string]string{
			"login":    "testadmin",
			"password": "testpassword123",
		}

		w := ts.makeRequest(t, "POST", "/api/admin/login", loginReq, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response["token"])
		assert.Equal(t, float64(admin.ID), response["admin"].(map[string]interface{})["id"])
		assert.Equal(t, admin.Login, response["admin"].(map[string]interface{})["login"])
	})

	t.Run("invalid_credentials", func(t *testing.T) {
		loginReq := map[string]string{
			"login":    "testadmin",
			"password": "wrongpassword",
		}

		w := ts.makeRequest(t, "POST", "/api/admin/login", loginReq, nil)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid login or password", response["error"])
	})

	t.Run("missing_fields", func(t *testing.T) {
		loginReq := map[string]string{
			"login": "testadmin",
			// password отсутствует
		}

		w := ts.makeRequest(t, "POST", "/api/admin/login", loginReq, nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "login and password are required", response["error"])
	})
}

func TestAPI_AdminProfile_Integration(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.db.Migrator().DropTable(&database.Admin{})

	// Создаем тестового администратора и получаем токен
	admin, token := ts.createTestAdmin(t, "testadmin", "testpassword123")

	t.Run("get_profile_with_valid_token", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := ts.makeRequest(t, "GET", "/api/admin/profile", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Проверяем, что ID администратора в токене совпадает
		adminID, err := auth.ExtractAdminID(token, ts.cfg.JWTSecretKey)
		require.NoError(t, err)
		assert.Equal(t, admin.ID, adminID)
		assert.Equal(t, float64(admin.ID), response["id"])
		assert.Equal(t, admin.Login, response["login"])
		assert.Equal(t, admin.IsActive, response["is_active"])
	})

	t.Run("get_profile_without_token", func(t *testing.T) {
		w := ts.makeRequest(t, "GET", "/api/admin/profile", nil, nil)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "authorization header required", response["error"])
	})

	t.Run("get_profile_with_invalid_token", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer invalid_token",
		}

		w := ts.makeRequest(t, "GET", "/api/admin/profile", nil, headers)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid token", response["error"])
	})
}

func TestAPI_AdminChangePassword_Integration(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.db.Migrator().DropTable(&database.Admin{})

	// Создаем тестового администратора и получаем токен
	admin, token := ts.createTestAdmin(t, "testadmin", "testpassword123")

	t.Run("change_password_success", func(t *testing.T) {
		changePasswordReq := map[string]string{
			"current_password": "testpassword123",
			"new_password":     "newpassword456",
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := ts.makeRequest(t, "POST", "/api/admin/change-password", changePasswordReq, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "password updated successfully", response["message"])

		// Проверяем, что можем войти с новым паролем
		loginReq := map[string]string{
			"login":    "testadmin",
			"password": "newpassword456",
		}

		w = ts.makeRequest(t, "POST", "/api/admin/login", loginReq, nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("change_password_wrong_current_password", func(t *testing.T) {
		changePasswordReq := map[string]string{
			"current_password": "wrongpassword",
			"new_password":     "newpassword789",
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := ts.makeRequest(t, "POST", "/api/admin/change-password", changePasswordReq, headers)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid current password", response["error"])
	})

	t.Run("change_password_short_new_password", func(t *testing.T) {
		changePasswordReq := map[string]string{
			"current_password": "testpassword123",
			"new_password":     "123", // слишком короткий
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := ts.makeRequest(t, "POST", "/api/admin/change-password", changePasswordReq, headers)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "min 8 chars")
	})
}

func TestAPI_HealthChecks_Integration(t *testing.T) {
	ts := setupTestServer(t)

	t.Run("health_check", func(t *testing.T) {
		w := ts.makeRequest(t, "GET", "/health", nil, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "goooo-bot", response["service"])
		assert.NotEmpty(t, response["timestamp"])
	})

	t.Run("readiness_check", func(t *testing.T) {
		w := ts.makeRequest(t, "GET", "/ready", nil, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "ready", response["status"])
		assert.Equal(t, "goooo-bot", response["service"])
		assert.Equal(t, "ready", response["database"])
		assert.NotEmpty(t, response["timestamp"])
	})
}

func TestAPI_RateLimiting_Integration(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.db.Migrator().DropTable(&database.Admin{})

	// Создаем тестового администратора
	ts.createTestAdmin(t, "testadmin", "testpassword123")

	t.Run("rate_limit_exceeded", func(t *testing.T) {
		loginReq := map[string]string{
			"login":    "testadmin",
			"password": "wrongpassword",
		}

		// Делаем больше запросов, чем разрешено (200 в минуту)
			// Для теста временно уменьшаем лимит в конфигурации.
		// Важно делать это на копии или с осторожностью в конкурентной среде.
		// Здесь мы меняем значение прямо в синглтоне, что допустимо для последовательных тестов.
		originalLimit := ts.cfg.RateLimitRequests
		ts.cfg.RateLimitRequests = 5
		defer func() {
			ts.cfg.RateLimitRequests = originalLimit
		}()

		// Делаем 6 запросов (превышаем лимит)
		for i := 0; i < 5; i++ {
			w := ts.makeRequest(t, "POST", "/api/admin/login", loginReq, nil)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		}

		// 6-й запрос должен быть заблокирован
		w := ts.makeRequest(t, "POST", "/api/admin/login", loginReq, nil)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "rate limit exceeded", response["error"])
		assert.NotEmpty(t, response["retry_after"])
	})
}

// Benchmark тесты для интеграционных тестов
func BenchmarkAPI_AdminLogin_Integration(b *testing.B) {
	ts := setupTestServer(&testing.T{})
	defer ts.db.Migrator().DropTable(&database.Admin{})

	// Создаем тестового администратора
	ts.createTestAdmin(&testing.T{}, "testadmin", "testpassword123")

	loginReq := map[string]string{
		"login":    "testadmin",
		"password": "testpassword123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := ts.makeRequest(&testing.T{}, "POST", "/api/admin/login", loginReq, nil)
		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
} 