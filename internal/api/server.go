package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"go-bot/internal/api/apierror"
	"go-bot/internal/api/handlers"
	"go-bot/internal/api/middleware"
	"go-bot/internal/config"
	"go-bot/internal/service"
	"go-bot/internal/services"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// Server is the main application server.
type Server struct {
	router     *gin.Engine
	logger     *slog.Logger
	db         *gorm.DB
	bot        *tgbotapi.BotAPI
	cfg        *config.Config
	xuiService *service.XUIService
	httpServer *http.Server
}

// NewServer creates a new server instance.
func NewServer(logger *slog.Logger, db *gorm.DB, bot *tgbotapi.BotAPI, cfg *config.Config, xuiService *service.XUIService) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	server := &Server{
		router:     router,
		logger:     logger,
		db:         db,
		bot:        bot,
		cfg:        cfg,
		xuiService: xuiService,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			Handler: router,
		},
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
		webhookHandler := handlers.NewWebhookHandler(webhookService, s.logger, s.xuiService) // Добавлен логгер

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

// Start runs the HTTP server.
func (s *Server) Start() error {
	s.logger.Info("Server starting", "address", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}