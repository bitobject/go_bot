package handlers

import (
	"log"

	"go-bot/internal/services"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type WebhookHandler struct {
	bot *tgbotapi.BotAPI
	db  *gorm.DB
}

func NewWebhookHandler(bot *tgbotapi.BotAPI, db *gorm.DB) *WebhookHandler {
	return &WebhookHandler{
		bot: bot,
		db:  db,
	}
}

// HandleWebhook обрабатывает входящие webhook'и от Telegram
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	var update tgbotapi.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		log.Printf("Error decoding webhook update: %v", err)
		c.JSON(400, gin.H{"error": "invalid webhook data"})
		return
	}

	// Обрабатываем update через bot handler
	bot.ProcessUpdate(update, h.bot, h.db)

	c.Status(200)
} 