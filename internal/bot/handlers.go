package bot

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"gorm.io/gorm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Start(bot *tgbotapi.BotAPI, db *gorm.DB) {
	// Setup webhook
	setupWebhook(bot)
	
	// Start HTTP server for webhook
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		handleWebhook(w, r, bot, db)
	})
	
	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy"))
	})
	
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Starting webhook server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Panic(err)
	}
}

func setupWebhook(bot *tgbotapi.BotAPI) {
	webhookURL := "https://body-architect.ru/webhook"
	
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

func handleWebhook(w http.ResponseWriter, r *http.Request, bot *tgbotapi.BotAPI, db *gorm.DB) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var update tgbotapi.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("Error decoding update: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	
	// Process the update
	processUpdate(update, bot, db)
	
	w.WriteHeader(http.StatusOK)
}

func processUpdate(update tgbotapi.Update, bot *tgbotapi.BotAPI, db *gorm.DB) {
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
	
	// Send response
	msg := tgbotapi.NewMessage(message.Chat.ID, "Привет! Я бот с webhook.")
	bot.Send(msg)
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