package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goooo/internal/api/handlers"
	"goooo/internal/api/middleware"
	"goooo/internal/bot"
	"goooo/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type Server struct {
	router *gin.Engine
	server *http.Server
	db     *gorm.DB
	bot    *tgbotapi.BotAPI
}

func NewServer(db *gorm.DB, bot *tgbotapi.BotAPI) *Server {
	// Настройка логирования
	setupLogging()

	// Создаем Gin router
	router := gin.New()

	// Middleware
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.RequestLogger())
	router.Use(gin.Recovery())

	// Создаем handlers
	adminHandler := handlers.NewAdminHandler(db)
	healthHandler := handlers.NewHealthHandler(db)
	webhookHandler := handlers.NewWebhookHandler(bot, db)

	// Настраиваем роуты
	setupRoutes(router, adminHandler, healthHandler, webhookHandler)

	// Создаем HTTP сервер
	server := &http.Server{
		Addr:    config.AppConfig.Host + ":" + config.AppConfig.Port,
		Handler: router,
	}

	return &Server{
		router: router,
		server: server,
		db:     db,
		bot:    bot,
	}
}

func setupLogging() {
	// Настройка уровня логирования
	level, err := logrus.ParseLevel(config.AppConfig.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// Настройка формата
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
}

func setupRoutes(
	router *gin.Engine,
	adminHandler *handlers.AdminHandler,
	healthHandler *handlers.HealthHandler,
	webhookHandler *handlers.WebhookHandler,
) {
	// Health checks (без rate limiting)
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)

	// API routes с rate limiting
	api := router.Group("/api")
	api.Use(middleware.RateLimitMiddleware())

	// Admin routes
	admin := api.Group("/admin")
	{
		admin.POST("/login", adminHandler.Login)
		
		// Защищенные routes
		protected := admin.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/profile", adminHandler.GetProfile)
			protected.POST("/change-password", adminHandler.ChangePassword)
		}
	}

	// Webhook (без rate limiting для Telegram)
	api.POST("/webhook", webhookHandler.HandleWebhook)
}

func (s *Server) Start() error {
	// Настраиваем webhook для Telegram
	bot.SetupWebhook(s.bot)

	logrus.Infof("Starting server on %s:%s", config.AppConfig.Host, config.AppConfig.Port)

	// Запускаем сервер в горутине
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Даем серверу 30 секунд на завершение
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
		return err
	}

	logrus.Info("Server exited")
	return nil
} 