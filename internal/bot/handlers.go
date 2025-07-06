package bot

import (
	"log"
	"gorm.io/gorm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SetupWebhook настраивает webhook для Telegram бота
func SetupWebhook(bot *tgbotapi.BotAPI) {
	webhookURL := "https://body-architect.ru/api/webhook"
	
	// Delete existing webhook first
	_, err := bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		log.Printf("Error deleting webhook: %v", err)
	}
	
	// Set new webhook
	webhookConfig, _ := tgbotapi.NewWebhook(webhookURL)
	webhookConfig.MaxConnections = 100
	
	_, err = bot.Request(webhookConfig)
	if err != nil {
		log.Printf("Error setting webhook: %v", err)
	} else {
		log.Printf("Webhook set successfully: %s", webhookURL)
	}
}

// ProcessUpdate обрабатывает входящие update от Telegram
func ProcessUpdate(update tgbotapi.Update, bot *tgbotapi.BotAPI, db *gorm.DB) {
	if update.Message != nil {
		handleMessage(update.Message, bot, db)
	}
	
	// Handle other types of updates (callback queries, etc.)
	if update.CallbackQuery != nil {
		handleCallbackQuery(update.CallbackQuery, bot, db)
	}
}

func handleMessage(message *tgbotapi.Message, bot *tgbotapi.BotAPI, db *gorm.DB) {
	// Save user to database
	saveUser(message.From, db)
	
	// Save message to database
	saveMessage(message, db)
	
	// Handle commands
	if message.IsCommand() {
		handleCommand(message, bot, db)
		return
	}
	
	// Handle regular messages
	msg := tgbotapi.NewMessage(message.Chat.ID, "Привет! Я бот с webhook.")
	bot.Send(msg)
}

func handleCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, db *gorm.DB) {
	switch message.Command() {
	case "start":
		msg := tgbotapi.NewMessage(message.Chat.ID, "Привет! Я бот. Используй /help для получения справки.")
		bot.Send(msg)
	case "help":
		msg := tgbotapi.NewMessage(message.Chat.ID, "Доступные команды:\n/start - Начать работу с ботом\n/help - Показать справку")
		bot.Send(msg)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная команда. Используй /help для получения справки.")
		bot.Send(msg)
	}
}

func handleCallbackQuery(callback *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, db *gorm.DB) {
	// Handle callback queries
	callbackConfig := tgbotapi.NewCallback(callback.ID, "Callback received!")
	bot.Request(callbackConfig)
}

func saveUser(user *tgbotapi.User, db *gorm.DB) {
	// Implementation for saving user to database
	// This is a placeholder - implement based on your database schema
	log.Printf("User: %d - %s", user.ID, user.UserName)
}

func saveMessage(message *tgbotapi.Message, db *gorm.DB) {
	// Implementation for saving message to database
	// This is a placeholder - implement based on your database schema
	log.Printf("Message from %d: %s", message.From.ID, message.Text)
} 