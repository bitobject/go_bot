package database

import (
	"time"
)

// User represents a Telegram user
type User struct {
	ID         uint      `gorm:"primaryKey"`
	TelegramID int64     `gorm:"uniqueIndex:idx_users_telegram_id;not null"`
	Username   string    `gorm:"size:255"`
	FirstName  string    `gorm:"size:255"`
	LastName   string    `gorm:"size:255"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Message represents a user message
type Message struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      int64     `gorm:"index"`
	MessageText string    `gorm:"type:text"`
	MessageType string    `gorm:"size:50"`
	CreatedAt   time.Time
}

// Admin represents an administrator with security features
type Admin struct {
	ID                  uint      `gorm:"primaryKey"`
		Login               string    `gorm:"type:citext;uniqueIndex;not null"`
		HashedPassword      string    `gorm:"size:255;not null"`
	IsActive            bool      `gorm:"default:true"`
	LastLoginAt         *time.Time
	FailedLoginAttempts int       `gorm:"default:0"`
	LockedUntil         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// TableName specifies the table name for Admin
func (Admin) TableName() string {
	return "admins"
} 