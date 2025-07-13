package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go-bot/internal/service"
	"go-bot/internal/xui"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// SetupWebhook настраивает webhook для Telegram бота
func SetupWebhook(bot *tgbotapi.BotAPI, webhookURL string) {
	// Delete existing webhook first
	_, err := bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		log.Printf("Error deleting webhook: %v", err)
	}

	// Set new webhook
	webhookConfig, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		log.Printf("Error creating webhook config: %v", err)
		return
	}
	webhookConfig.MaxConnections = 100

	_, err = bot.Request(webhookConfig)
	if err != nil {
		log.Printf("Error setting webhook: %v", err)
	} else {
		log.Printf("Webhook set successfully: %s", webhookURL)
	}
}

// ProcessUpdate обрабатывает входящие update от Telegram
func ProcessUpdate(ctx context.Context, update tgbotapi.Update, bot *tgbotapi.BotAPI, db *gorm.DB, xuiService *service.XUIService) {
	if update.Message != nil {
		handleMessage(ctx, update.Message, bot, db, xuiService)
	}

	// Handle other types of updates (callback queries, etc.)
	if update.CallbackQuery != nil {
		handleCallbackQuery(update.CallbackQuery, bot, db)
	}
}

func handleMessage(ctx context.Context, message *tgbotapi.Message, bot *tgbotapi.BotAPI, db *gorm.DB, xuiService *service.XUIService) {
	// Ignore messages without a sender
	if message.From == nil {
		return
	}

	// Save user and message to the database
	saveUser(message.From, db)
	saveMessage(message, db)

	// Handle commands
	if message.IsCommand() {
		handleCommand(ctx, message, bot, db, xuiService)
		return
	}

	// Handle regular messages
	log.Printf("Message from %s: %s", formatUserInfo(message.From), message.Text)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Ваше сообщение получено: "+message.Text)
	bot.Send(msg)
}

func handleCommand(ctx context.Context, message *tgbotapi.Message, bot *tgbotapi.BotAPI, db *gorm.DB, xuiService *service.XUIService) {
	userInfo := formatUserInfo(message.From)

	switch message.Command() {
	case "start":
		// Corrected greeting message
		msgText := "Привет, " + userInfo + ", рады видеть вас снова!"
		msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
		bot.Send(msg)
	case "help":
		msg := tgbotapi.NewMessage(message.Chat.ID, "Доступные команды:\n/start - Начать работу с ботом\n/help - Показать справку\n/getclient <email> - Получить данные по клиенту")
		bot.Send(msg)
	case "getclient":
		handleGetClientCommand(ctx, message, bot, xuiService)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная команда. Используй /help для получения справки.")
		bot.Send(msg)
	}
}

func handleCallbackQuery(callback *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, db *gorm.DB) {
	// Handle callback queries
	callbackConfig := tgbotapi.NewCallback(callback.ID, "Callback received!")
	bot.Request(callbackConfig)

	// You might want to log the callback data
	log.Printf("Callback from %s: %s", callback.From.UserName, callback.Data)
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

// formatUserInfo creates a user-friendly string from a User object.
func handleGetClientCommand(ctx context.Context, message *tgbotapi.Message, bot *tgbotapi.BotAPI, xuiService *service.XUIService) {
	email := strings.TrimSpace(message.CommandArguments())
	if email == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Пожалуйста, укажите email после команды. Пример: /getclient user@example.com")
		bot.Send(msg)
		return
	}

	clientTraffics, err := xuiService.GetClientTraffics(ctx, email)
	if err != nil {
		log.Printf("Error getting client traffics: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Не удалось получить данные для клиента %s. Возможно, сервис временно недоступен.", email))
		bot.Send(msg)
		return
	}

	if len(clientTraffics) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Клиент с email %s не найден.", email))
		bot.Send(msg)
		return
	}

	var responseText strings.Builder
	responseText.WriteString(fmt.Sprintf("<b>Данные для клиента %s:</b>\n\n", email))
	for _, traffic := range clientTraffics {
		responseText.WriteString(formatClientTraffic(traffic))
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, responseText.String())
	msg.ParseMode = tgbotapi.ModeHTML
	bot.Send(msg)
}

func formatClientTraffic(traffic xui.ClientTraffic) string {
	// Convert Unix timestamp (milliseconds) to time.Time
	expiry := time.Unix(0, traffic.ExpiryTime*int64(time.Millisecond))

	// Convert bytes to a more readable format (e.g., GB)
	upGB := float64(traffic.Up) / (1024 * 1024 * 1024)
	downGB := float64(traffic.Down) / (1024 * 1024 * 1024)
	totalGB := float64(traffic.Total) / (1024 * 1024 * 1024)

	return fmt.Sprintf("<b>Email:</b> %s\n<b>Статус:</b> %s\n<b>Трафик:</b> %.2f / %.2f GB\n<b>Сброс:</b> %s\n\n",
		traffic.Email,
		map[bool]string{true: "✅ Включен", false: "❌ Отключен"}[traffic.Enable],
		upGB+downGB,
		totalGB,
		expiry.Format("02.01.2006"))
}

// formatUserInfo creates a user-friendly string from a User object.
func formatUserInfo(user *tgbotapi.User) string {
	if user == nil {
		return "Anonymous"
	}
	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}
	return name
}
