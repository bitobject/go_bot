package api

import (
	"bytes"
	"encoding/json"
	"go-bot/internal/config"
	"go-bot/internal/service"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

// TestWebhookHandler_Integration_GetClientCommand выполняет интеграционный тест для команды /getclient.
// Он использует реальный XUIService для взаимодействия с API 3x-ui.
func TestWebhookHandler_Integration_GetClientCommand(t *testing.T) {
	// Загружаем переменные окружения из папки deploy
	err := godotenv.Load("../../deploy/.env")
	require.NoError(t, err, "Failed to load .env file")

	// Инициализируем конфигурацию
	cfg, err := config.Load("../../deploy/.env")
	require.NoError(t, err, "Failed to load .env config")

	// Проверяем, что необходимые для теста переменные установлены
	require.NotEmpty(t, cfg.XUIURL, "XUI_URL must be set in .env for this test")
	require.NotEmpty(t, cfg.XUIUsername, "XUI_USERNAME must be set in .env for this test")
	require.NotEmpty(t, cfg.XUIPassword, "XUI_PASSWORD must be set in .env for this test")
	require.NotEmpty(t, cfg.TelegramToken, "TELEGRAM_TOKEN must be set in .env for this test")

	// Настраиваем логирование
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Инициализируем реальный сервис XUI
	xuiService := service.NewXUIService(cfg, logger)

	// Инициализируем бота (без реальной отправки сообщений)
	tgBot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	require.NoError(t, err)

	// Создаем сервер (без БД, так как для этой команды она не нужна)
	server := NewServer(logger, nil, tgBot, cfg, xuiService)

	// --- Подготовка тестового запроса ---

	// Создаем фейковое обновление от Telegram
	update := tgbotapi.Update{
		UpdateID: 12345,
		Message: &tgbotapi.Message{
			MessageID: 54321,
			From: &tgbotapi.User{
				ID:       98765,
				UserName: "testuser",
			},
			Chat: &tgbotapi.Chat{
				ID: 98765,
			},
			Text: "/getclient ipad@ipad.ru", // Используем реальный email для теста
		},
	}

	body, err := json.Marshal(update)
	require.NoError(t, err)

	// Создаем HTTP-запрос, имитирующий вызов вебхука
	url := APIPrefix + WebhookPath
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Создаем рекордер для записи ответа
	rr := httptest.NewRecorder()

	// --- Выполнение запроса ---
	server.router.ServeHTTP(rr, req)

	// --- Проверка результата ---

	// Ожидаем, что сервер ответит Telegram'у статусом 200 OK,
	// подтверждая успешное получение и обработку обновления.
	require.Equal(t, http.StatusOK, rr.Code, "Expected status OK")

	// В логах приложения мы должны увидеть результат отправки сообщения ботом.
	// Для этого теста достаточно убедиться, что обработчик отработал без паники и вернул OK.
	t.Log("Webhook handler successfully processed the /getclient command.")
}
