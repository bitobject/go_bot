package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go-bot/internal/service"

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
func ProcessUpdate(ctx context.Context, update tgbotapi.Update, bot BotSender, db *gorm.DB, xuiService service.ClientTrafficProvider) {
	if update.Message != nil {
		handleMessage(ctx, update.Message, bot, db, xuiService)
	}

	// Handle other types of updates (callback queries, etc.)
	if update.CallbackQuery != nil {
		handleCallbackQuery(update.CallbackQuery, bot, db)
	}
}

func handleMessage(ctx context.Context, message *tgbotapi.Message, bot BotSender, db *gorm.DB, xuiService service.ClientTrafficProvider) {
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

func handleCommand(ctx context.Context, message *tgbotapi.Message, bot BotSender, db *gorm.DB, xuiService service.ClientTrafficProvider) {
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

// BotSender defines the interface for sending messages and making requests, allowing for mocking in tests.
type BotSender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

func handleCallbackQuery(callback *tgbotapi.CallbackQuery, bot BotSender, db *gorm.DB) {
	// Handle callback queries
	callbackConfig := tgbotapi.NewCallback(callback.ID, "Callback received!")
	_, err := bot.Request(callbackConfig)
	if err != nil {
		log.Printf("ERROR: failed to send callback response: %v", err)
	}

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
// handleGetClientCommand processes the /getclient command and sends the data as a formatted table.
func handleGetClientCommand(ctx context.Context, message *tgbotapi.Message, bot BotSender, xuiService service.ClientTrafficProvider) {
	email := strings.TrimSpace(message.CommandArguments())
	if email == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Пожалуйста, укажите email после команды. Пример: /getclient user@example.com")
		bot.Send(msg)
		return
	}

	clientTraffics, err := xuiService.GetClientTraffics(ctx, email)
	if err != nil {
		log.Printf("ERROR: Failed to get client traffics for email [%s]: %v", email, err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка при получении данных. Пожалуйста, попробуйте позже.")
		bot.Send(msg)
		return
	}

	if len(clientTraffics) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Клиент с email %s не найден.", email))
		bot.Send(msg)
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Данные для клиента:</b> <code>%s</code>\n", email))
	sb.WriteString("<pre>")
	sb.WriteString(fmt.Sprintf("%-22s | %-13s | %-10s | %s\n", "Email", "Usage (GB)", "Expiry", "Status"))
	sb.WriteString(strings.Repeat("-", 56) + "\n")

	for _, traffic := range clientTraffics {
		upGB := float64(traffic.Up) / (1024 * 1024 * 1024)
		downGB := float64(traffic.Down) / (1024 * 1024 * 1024)
		totalGB := float64(traffic.Total) / (1024 * 1024 * 1024)
		usageStr := fmt.Sprintf("%.2f/%.2f", upGB+downGB, totalGB)

		expiry := time.Unix(0, traffic.ExpiryTime*int64(time.Millisecond))
		expiryStr := expiry.Format("02.01.2006")

		status := "❌"
		if traffic.Enable {
			status = "✅"
		}

		// Truncate email if it's too long
		displayEmail := traffic.Email
		if len(displayEmail) > 20 {
			displayEmail = displayEmail[:19] + ".."
		}

		sb.WriteString(fmt.Sprintf("%-22s | %-13s | %-10s | %s\n", displayEmail, usageStr, expiryStr, status))
	}
	sb.WriteString("</pre>")

	msg := tgbotapi.NewMessage(message.Chat.ID, sb.String())
	msg.ParseMode = tgbotapi.ModeHTML
	bot.Send(msg)
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
