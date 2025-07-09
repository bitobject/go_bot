package api

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-bot/internal/api/handlers"
	"go-bot/internal/api/middleware"
	"go-bot/internal/bot"
	"go-bot/internal/config"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type Server struct {
	router *gin.Engine
	server *http.Server
	db     *gorm.DB
	bot    *tgbotapi.BotAPI
	logger *slog.Logger
}

func NewServer(cfg *config.Config, db *gorm.DB, bot *tgbotapi.BotAPI, logger *slog.Logger) *Server {

	// Middleware
	router := gin.New()
	router.Use(middleware.LoggerInjector(logger))
	router.Use(middleware.SlogLogger())
	router.Use(gin.Recovery())

	// Создаем handlers
	jwtExpiresIn := time.Hour * time.Duration(cfg.JWTExpiresIn)
	adminHandler := handlers.NewAdminHandler(db, logger, cfg.JWTSecretKey, jwtExpiresIn)
	healthHandler := handlers.NewHealthHandler(db)
	webhookHandler := handlers.NewWebhookHandler(bot, db)

	// Настраиваем роуты
	setupRoutes(router, cfg, adminHandler, healthHandler, webhookHandler)

	// Создаем HTTP сервер
	server := &http.Server{
		Addr:    cfg.Host + ":" + cfg.Port,
		Handler: router,
	}

	return &Server{
		router: router,
		server: server,
		db:     db,
		bot:    bot,
		logger: logger,
	}
}



func setupRoutes(
	router *gin.Engine,
	cfg *config.Config,
	adminHandler *handlers.AdminHandler,
	healthHandler *handlers.HealthHandler,
	webhookHandler *handlers.WebhookHandler,
) {
	// Health checks (без rate limiting)
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)

	// API routes с rate limiting
	api := router.Group("/api")
	api.Use(middleware.RateLimitMiddleware(cfg))

	// Admin routes
	admin := api.Group("/admin")
	{
		admin.POST("/login", middleware.ErrorHandler(adminHandler.Login))
		
		// Защищенные routes
		protected := admin.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecretKey))
		{
			protected.GET("/profile", middleware.ErrorHandler(adminHandler.GetProfile))
			protected.POST("/change-password", middleware.ErrorHandler(adminHandler.ChangePassword))
		}
	}

	// Webhook (без rate limiting для Telegram)
	api.POST("/webhook", webhookHandler.HandleWebhook)
}

func (s *Server) Start() error {
	// Настраиваем webhook для Telegram
	bot.SetupWebhook(s.bot)

	s.logger.Info("starting server", "address", s.server.Addr)

	// Запускаем сервер в горутине
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("server error", "error", err)
			// Поскольку это горутина, мы не можем вернуть ошибку. Выход - один из вариантов.
			// В реальном приложении здесь может быть более сложная логика.
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("shutting down server...")

	// Даем серверу 30 секунд на завершение
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("server forced to shutdown", "error", err)
		return err
	}

	s.logger.Info("server exited gracefully")
	return nil
} 