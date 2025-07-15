package service

import (
	"context"
	"fmt"
	"log/slog"

	"go-bot/internal/config"
	"go-bot/internal/xui"
)

// ClientTrafficProvider defines the interface for services that can fetch client traffic data.
// This abstraction is crucial for decoupling and testing the bot handlers.
type ClientTrafficProvider interface {
	GetClientTraffics(ctx context.Context, email string) ([]xui.ClientTraffic, error)
}

// XUIService provides a high-level interface for interacting with the 3x-ui API.
type XUIService struct {
	client *xui.Client
	logger *slog.Logger
}

// NewXUIService creates a new XUIService.
func NewXUIService(cfg *config.Config, logger *slog.Logger) *XUIService {
	client := xui.NewClient(cfg.XUIURL, cfg.XUIUsername, cfg.XUIPassword, logger)
	return &XUIService{
		client: client,
		logger: logger,
	}
}

// GetClientTraffics retrieves traffic data for a specific client by email.
func (s *XUIService) GetClientTraffics(ctx context.Context, email string) ([]xui.ClientTraffic, error) {
	traffics, err := s.client.GetClientTraffics(ctx, email)
	if err != nil {
		wrappedErr := fmt.Errorf("XUIService error: %w", err)
		s.logger.Error(
			"failed to get client traffics from X-UI API",
			"error", wrappedErr,
			"email", email,
		)
		return nil, wrappedErr
	}
	return traffics, nil
}
