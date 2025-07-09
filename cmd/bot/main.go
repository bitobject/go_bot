package main

import (
	"log/slog"
	"os"

	"go-bot/internal/api"
	"go-bot/internal/bot"
	"go-bot/internal/config"
	"go-bot/internal/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Инициализация логгера
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// 2. Загрузка .env файла (для локальной разработки)
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found, relying on environment variables")
	}

	// 3. Загрузка и валидация конфигурации
	cfg := config.Get()

	// 4. Подключение к базе данных
	db := database.Init(cfg)
	logger.Info("Successfully connected to the database")

	// 5. Инициализация Telegram бота
	tgBot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		logger.Error("Failed to initialize Telegram bot", "error", err)
		os.Exit(1)
	}
	logger.Info("Telegram bot authorized", "bot_username", tgBot.Self.UserName)

	// 6. Настройка вебхука
	// URL для вебхука жестко закодирован внутри функции SetupWebhook.
	bot.SetupWebhook(tgBot)
	logger.Info("Telegram webhook set successfully")

	// 7. Создание и запуск сервера
	server := api.NewServer(logger, db, tgBot, cfg)
	server.Start()
}