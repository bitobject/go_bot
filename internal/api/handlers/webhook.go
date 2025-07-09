package handlers

import (
	"log/slog"
	"net/http"

	"go-bot/internal/api/apierror"
	"go-bot/internal/services"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WebhookHandler handles webhook-related API endpoints.
type WebhookHandler struct {
	webhookService services.WebhookServiceInterface
	logger         *slog.Logger
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(webhookService services.WebhookServiceInterface, logger *slog.Logger) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
		logger:         logger,
	}
}

// HandleWebhook processes incoming webhooks from Telegram.
func (h *WebhookHandler) HandleWebhook(c *gin.Context) error {
	var update tgbotapi.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		h.logger.Error("Failed to bind JSON for webhook update", "error", err)
		return apierror.New(http.StatusBadRequest, "invalid request body")
	}

	if err := h.webhookService.ProcessUpdate(c.Request.Context(), update); err != nil {
		// The service layer should handle its own logging of the error.
		return apierror.New(http.StatusInternalServerError, "failed to process update")
	}

	c.Status(http.StatusOK)
	return nil
}
 