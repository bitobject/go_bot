package main

import (
	"log/slog"
	"os"

	"go-bot/internal/api"
	"go-bot/internal/config"
	"go-bot/internal/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Загружаем конфигурацию.
	// config.Get() загрузит, провалидирует и вернет синглтон-экземпляр.
	// Если конфигурация некорректна, приложение завершится с паникой.
	cfg := config.Get()

	// Настройка логгера
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		log.Printf("Invalid log level '%s', defaulting to INFO", cfg.LogLevel)
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger) // Устанавливаем как логгер по умолчанию для удобства

	// Инициализируем базу данных
	db, err := database.Init(cfg)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close(db)

	// Создаем Telegram бота.
	// Проверка на пустой токен больше не нужна, т.к. валидатор в config.Get() уже это сделал.
	tgBot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		logger.Error("failed to create bot", "error", err)
		os.Exit(1)
	}

	// Инициализация API сервера
	server := api.NewServer(cfg, db, tgBot, logger)
	if err := server.Start(); err != nil {
		logger.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}