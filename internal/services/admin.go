package services

import (
	"errors"
	"fmt"
	"log/slog"

	"go-bot/internal/auth"
	"go-bot/internal/database"

	"gorm.io/gorm"
)

// AdminService provides operations for admin users.
type AdminService struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewAdminService creates a new AdminService.
func NewAdminService(db *gorm.DB, logger *slog.Logger) *AdminService {
	return &AdminService{
		db:     db,
		logger: logger,
	}
}

// Authenticate checks admin credentials and returns the admin if successful.
func (s *AdminService) Authenticate(login, password string) (*database.Admin, error) {
	var admin database.Admin
	if err := s.db.Where("login = ?", login).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		s.logger.Error("failed to find admin by login", "error", err, "login", login)
		return nil, err
	}

	if !auth.CheckPasswordHash(password, admin.Password) {
		return nil, errors.New("invalid credentials")
	}

	return &admin, nil
}

// GetProfile retrieves an admin's profile by their ID.
func (s *AdminService) GetProfile(adminID uint) (*database.Admin, error) {
	var admin database.Admin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("admin with ID %d not found", adminID)
		}
		s.logger.Error("failed to find admin by ID", "error", err, "adminID", adminID)
		return nil, err
	}
	return &admin, nil
}

// ChangePassword updates an admin's password after verifying the old one.
func (s *AdminService) ChangePassword(adminID uint, oldPassword, newPassword string) error {
	admin, err := s.GetProfile(adminID)
	if err != nil {
		return err
	}

	if !auth.CheckPasswordHash(oldPassword, admin.Password) {
		return errors.New("incorrect old password")
	}

	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		s.logger.Error("failed to hash new password", "error", err, "adminID", adminID)
		return errors.New("failed to update password")
	}

	if err := s.db.Model(&admin).Update("password", hashedPassword).Error; err != nil {
		s.logger.Error("failed to update password in db", "error", err, "adminID", adminID)
		return errors.New("failed to update password")
	}

	return nil
}

// CreateAdmin creates a new admin user.
func (s *AdminService) CreateAdmin(login, password string) (*database.Admin, error) {
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("could not hash password: %w", err)
	}

	admin := &database.Admin{
		Login:    login,
		Password: hashedPassword,
		IsActive: true,
	}

	if err := s.db.Create(admin).Error; err != nil {
		return nil, fmt.Errorf("could not create admin: %w", err)
	}

	return admin, nil
}
