package database

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	// Security constants
	MaxLoginAttempts = 5
	LockoutDuration  = 15 * time.Minute
	BCryptCost       = 12 // Recommended cost for bcrypt
)

var (
	ErrAdminNotFound      = errors.New("admin not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account is locked")
	ErrAccountInactive    = errors.New("account is inactive")
)

// AdminService provides methods for admin authentication and management
type AdminService struct {
	db *gorm.DB
}

// NewAdminService creates a new admin service
func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{db: db}
}

// CreateAdmin creates a new admin with hashed password
func (s *AdminService) CreateAdmin(login, password string) (*Admin, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), BCryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	admin := &Admin{
		Login:          login,
		HashedPassword: string(hashedPassword),
		IsActive:       true,
	}

	if err := s.db.Create(admin).Error; err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	return admin, nil
}

// AuthenticateAdmin authenticates an admin with login and password
func (s *AdminService) AuthenticateAdmin(login, password string) (*Admin, error) {
	var admin Admin

	// Find admin by login (case-insensitive due to CITEXT)
	if err := s.db.Where("login = ?", login).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminNotFound
		}
		return nil, fmt.Errorf("failed to find admin: %w", err)
	}

	// Check if account is active
	if !admin.IsActive {
		return nil, ErrAccountInactive
	}

	// Check if account is locked
	if admin.LockedUntil != nil && admin.LockedUntil.After(time.Now()) {
		return nil, ErrAccountLocked
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.HashedPassword), []byte(password)); err != nil {
		// Increment failed login attempts
		s.incrementFailedAttempts(&admin)
		return nil, ErrInvalidCredentials
	}

	// Reset failed attempts on successful login
	if admin.FailedLoginAttempts > 0 {
		s.resetFailedAttempts(&admin)
	}

	// Update last login time
	now := time.Now()
	admin.LastLoginAt = &now
	s.db.Save(&admin)

	return &admin, nil
}

// incrementFailedAttempts increments failed login attempts and locks account if needed
func (s *AdminService) incrementFailedAttempts(admin *Admin) {
	admin.FailedLoginAttempts++

	if admin.FailedLoginAttempts >= MaxLoginAttempts {
		lockoutTime := time.Now().Add(LockoutDuration)
		admin.LockedUntil = &lockoutTime
	}

	s.db.Save(admin)
}

// resetFailedAttempts resets failed login attempts
func (s *AdminService) resetFailedAttempts(admin *Admin) {
	admin.FailedLoginAttempts = 0
	admin.LockedUntil = nil
	s.db.Save(admin)
}

// GetAdminByID retrieves an admin by ID
func (s *AdminService) GetAdminByID(id uint) (*Admin, error) {
	var admin Admin
	if err := s.db.First(&admin, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminNotFound
		}
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}
	return &admin, nil
}

// GetAdminByLogin retrieves an admin by login
func (s *AdminService) GetAdminByLogin(login string) (*Admin, error) {
	var admin Admin
	if err := s.db.Where("login = ?", login).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminNotFound
		}
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}
	return &admin, nil
}

// UpdateAdminPassword updates admin password with new hashed password
func (s *AdminService) UpdateAdminPassword(adminID uint, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), BCryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.db.Model(&Admin{}).Where("id = ?", adminID).Update("hashed_password", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// DeactivateAdmin deactivates an admin account
func (s *AdminService) DeactivateAdmin(adminID uint) error {
	if err := s.db.Model(&Admin{}).Where("id = ?", adminID).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate admin: %w", err)
	}
	return nil
}

// ActivateAdmin activates an admin account
func (s *AdminService) ActivateAdmin(adminID uint) error {
	if err := s.db.Model(&Admin{}).Where("id = ?", adminID).Update("is_active", true).Error; err != nil {
		return fmt.Errorf("failed to activate admin: %w", err)
	}
	return nil
}

// UnlockAdmin unlocks a locked admin account
func (s *AdminService) UnlockAdmin(adminID uint) error {
	updates := map[string]interface{}{
		"failed_login_attempts": 0,
		"locked_until":          nil,
	}

	if err := s.db.Model(&Admin{}).Where("id = ?", adminID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to unlock admin: %w", err)
	}
	return nil
}

// GenerateSecurePassword generates a secure random password
func GenerateSecurePassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)

	for i := range password {
		randomByte := make([]byte, 1)
		if _, err := rand.Read(randomByte); err != nil {
			return "", fmt.Errorf("failed to generate random password: %w", err)
		}
		password[i] = charset[randomByte[0]%byte(len(charset))]
	}

	return string(password), nil
}
