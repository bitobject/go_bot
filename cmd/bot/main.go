package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-bot/internal/api"
	"go-bot/internal/bot"
	"go-bot/internal/config"
	"go-bot/internal/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Инициализация логгера
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// 2. Загрузка .env файла (для локальной разработки)
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, relying on environment variables")
	}

	// 3. Загрузка и валидация конфигурации
	cfg := config.Get()

	// 4. Подключение к базе данных
	db := database.Init(cfg)
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("failed to get underlying sql.DB", "error", err)
		return
	}
	defer sqlDB.Close()
	slog.Info("Successfully connected to the database")

	// 5. Инициализация Telegram бота
	tgBot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		slog.Error("Failed to initialize Telegram bot", "error", err)
		return // Заменяем os.Exit(1) на return для корректного defer
	}
	slog.Info("Telegram bot authorized", "bot_username", tgBot.Self.UserName)

	// 6. Настройка вебхука
	bot.SetupWebhook(tgBot, cfg.WebhookURL)
	slog.Info("Telegram webhook set successfully")

	// 7. Создание и запуск сервера с Graceful Shutdown
	server := api.NewServer(logger, db, tgBot, cfg)

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Ожидаем сигнал для завершения работы
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	// Даем 5 секунд на завершение всех активных запросов
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exiting")
}