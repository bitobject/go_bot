package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-bot/internal/api/apierror"
	"go-bot/internal/api/handlers"
	"go-bot/internal/api/middleware"
	"go-bot/internal/config"
	"go-bot/internal/services"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// Server is the main application server.
type Server struct {
	router *gin.Engine
	logger *slog.Logger
	db     *gorm.DB
	bot    *tgbotapi.BotAPI
	cfg    *config.Config
}

// NewServer creates a new server instance.
func NewServer(logger *slog.Logger, db *gorm.DB, bot *tgbotapi.BotAPI, cfg *config.Config) *Server {
	server := &Server{
		router: gin.Default(),
		logger: logger,
		db:     db,
		bot:    bot,
		cfg:    cfg,
	}
	server.setupRouter()
	return server
}

// setupRouter configures the API routes.
func (s *Server) setupRouter() {
	// Middlewares
	s.router.Use(middleware.LoggerInjector(s.logger)) // Внедряем логгер в контекст
	s.router.Use(middleware.SlogLogger())    // Используем логгер для запросов
	s.router.Use(gin.Recovery())

	// Health checks are public
	healthService := services.NewHealthService(s.db)
	healthHandler := handlers.NewHealthHandler(healthService)
	s.router.GET("/health", healthHandler.HealthCheck)
	s.router.GET("/ready", healthHandler.ReadinessCheck)

	// API group with rate limiting
	api := s.router.Group("/api")
	api.Use(middleware.RateLimitMiddleware(s.cfg))
	{
		// Services
		adminService := services.NewAdminService(s.db, s.logger)
		webhookService := services.NewWebhookService(s.db, s.bot, s.logger)

		// Handlers
		adminHandler := handlers.NewAdminHandler(adminService, s.logger, s.cfg.JWTSecretKey)
		webhookHandler := handlers.NewWebhookHandler(webhookService, s.logger) // Добавлен логгер

		// Webhook for Telegram
		api.POST("/webhook", apierror.ErrorWrapper(webhookHandler.HandleWebhook))

		// Admin routes
		admin := api.Group("/admin")
		{
			admin.POST("/login", apierror.ErrorWrapper(adminHandler.Login))

			// Protected routes
			authRequired := admin.Group("/", middleware.AuthMiddleware(s.cfg.JWTSecretKey)) // Используем middleware.JWT
			{
				authRequired.GET("/profile", apierror.ErrorWrapper(adminHandler.GetProfile))
				authRequired.POST("/change-password", apierror.ErrorWrapper(adminHandler.ChangePassword))
			}
		}
	}
}

// Start runs the HTTP server with graceful shutdown.
func (s *Server) Start() {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port),
		Handler: s.router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("Could not start server", "error", err)
			os.Exit(1)
		}
	}()

	s.logger.Info("Server started", "address", server.Addr)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
			s.logger.Error("Server forced to shutdown", "error", err)
	}

	s.logger.Info("Server exiting")
}