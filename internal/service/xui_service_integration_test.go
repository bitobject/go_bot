package service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"go-bot/internal/config"

	"github.com/stretchr/testify/require"
)

// TestXUIService_Integration_GetClientTraffics выполняет интеграционный тест для XUIService.
// Он использует реальные учетные данные из .env для взаимодействия с API 3x-ui.
// Для запуска этого теста необходимо наличие файла .env в директории deploy.
func TestXUIService_Integration_GetClientTraffics(t *testing.T) {
	// Пропускаем тест, если он запускается в CI/CD, где нет .env
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	// Загружаем конфигурацию из .env
	cfg, err := config.Load("../../deploy/.env")
	require.NoError(t, err, "Failed to load .env config")

	// Инициализируем логгер
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Создаем реальный сервис, передавая ему конфиг
	xuiService := NewXUIService(cfg, logger)

	// Вызываем метод, который хотим протестировать
	// Используем известный email, который точно существует в 3x-ui
	clientTraffics, err := xuiService.GetClientTraffics(context.Background(), "ipad@ipad.ru")

	// Проверяем результат
	require.NoError(t, err, "Failed to get client traffic data from XUI service")

	// Проверяем, что получили непустой результат и структура данных корректна
	require.NotEmpty(t, clientTraffics, "Expected to receive client traffic data, but got an empty slice")
	traffic := clientTraffics[0]
	require.NotZero(t, traffic.ID, "ClientTraffic.ID should not be zero")
	require.NotEmpty(t, traffic.Email, "ClientTraffic.Email should not be empty")

	t.Logf("Successfully retrieved and validated client traffic data for %s", traffic.Email)
}
