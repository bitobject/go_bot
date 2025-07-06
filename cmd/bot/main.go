package main

import (
	"log"
	"goooo/internal/api"
	"goooo/internal/config"
	"goooo/internal/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Загружаем конфигурацию
	config.LoadEnv()
	config.Init()

	// Инициализируем базу данных
	db := database.Init()
	defer database.Close(db)

	// Создаем Telegram бота
	botToken := config.AppConfig.TelegramToken
	if botToken == "" {
		log.Fatal("TELEGRAM_TOKEN is required")
	}

	tgBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// Создаем и запускаем сервер
	server := api.NewServer(db, tgBot)
	if err := server.Start(); err != nil {
		log.Fatal("Server error:", err)
	}
} 