package handlers

import (
	"log/slog"
	"net/http"

	"go-bot/internal/api/apierror"
	"go-bot/internal/bot"
	"go-bot/internal/config"
	"go-bot/internal/service"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// WebhookHandler handles webhook-related API endpoints.
type WebhookHandler struct {
	cfg    *config.Config
	logger *slog.Logger
	bot    bot.BotSender
	db     *gorm.DB
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(cfg *config.Config, logger *slog.Logger, bot bot.BotSender, db *gorm.DB) *WebhookHandler {
	return &WebhookHandler{
		cfg:    cfg,
		logger: logger,
		bot:    bot,
		db:     db,
	}
}

// HandleWebhook processes incoming webhooks from Telegram.
func (h *WebhookHandler) HandleWebhook(c *gin.Context) error {
	var update tgbotapi.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		h.logger.Error("Failed to bind JSON for webhook update", "error", err)
		return apierror.New(http.StatusBadRequest, "invalid request body")
	}

	// Создаем XUI сервис, используя зависимости из хендлера
	xuiService := service.NewXUIService(h.cfg, h.logger)

	// Асинхронно обрабатываем обновление, чтобы не блокировать ответ Telegram
	go bot.ProcessUpdate(c.Request.Context(), update, h.bot, h.db, xuiService)

	c.Status(http.StatusOK)
	return nil
}
