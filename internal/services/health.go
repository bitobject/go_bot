package services

import (
	"context"

	"gorm.io/gorm"
)

// HealthServiceInterface defines the interface for the health service.
type HealthServiceInterface interface {
	CheckDB(ctx context.Context) error
}

// HealthService provides health check functionalities.
type HealthService struct {
	db *gorm.DB
}

// NewHealthService creates a new HealthService.
func NewHealthService(db *gorm.DB) HealthServiceInterface {
	return &HealthService{db: db}
}

// CheckDB pings the database to check for connectivity.
func (s *HealthService) CheckDB(ctx context.Context) error {
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
}
