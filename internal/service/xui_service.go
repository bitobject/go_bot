package service

import (
	"context"

	"go-bot/internal/config"
	"go-bot/internal/xui"
)

// XUIService provides a high-level interface for interacting with the 3x-ui API.
type XUIService struct {
	client *xui.Client
}

// NewXUIService creates a new XUIService.
func NewXUIService(cfg *config.Config) *XUIService {
	client := xui.NewClient(cfg.XUIURL, cfg.XUIUsername, cfg.XUIPassword)
	return &XUIService{
		client: client,
	}
}

// GetClientTraffics retrieves traffic data for a specific client by email.
func (s *XUIService) GetClientTraffics(ctx context.Context, email string) ([]xui.ClientTraffic, error) {
	return s.client.GetClientTraffics(ctx, email)
}
