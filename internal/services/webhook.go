package services

import (
	"context"
	"log/slog"

	"go-bot/internal/bot"
	"go-bot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// WebhookServiceInterface defines the contract for the webhook service.
type WebhookServiceInterface interface {
	ProcessUpdate(ctx context.Context, update tgbotapi.Update, xuiService *service.XUIService) error
}

// WebhookService handles business logic for Telegram updates.
type WebhookService struct {
	db     *gorm.DB
	bot    *tgbotapi.BotAPI
	logger *slog.Logger
}

// NewWebhookService creates a new WebhookService.
func NewWebhookService(db *gorm.DB, bot *tgbotapi.BotAPI, logger *slog.Logger) WebhookServiceInterface {
	return &WebhookService{
		db:     db,
		bot:    bot,
		logger: logger,
	}
}

// ProcessUpdate processes a single update from Telegram by delegating to the bot package.
func (s *WebhookService) ProcessUpdate(ctx context.Context, update tgbotapi.Update, xuiService *service.XUIService) error {
	// Delegate the update processing to the bot package, which contains the core logic.
	s.logger.Info("Delegating update to bot processor", "update_id", update.UpdateID)
	bot.ProcessUpdate(ctx, update, s.bot, s.db, xuiService)
	return nil // The bot package handles errors internally by logging them.
}
