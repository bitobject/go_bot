package database

import (
	"goooo/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func Init() *gorm.DB {
	dsn := config.AppConfig.DatabaseURL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic("failed to connect database:", err)
	}

	// Auto migrate models
	err = db.AutoMigrate(&User{}, &Message{}, &Admin{})
	if err != nil {
		log.Panic("failed to migrate database:", err)
	}

	return db
}

func Close(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
} 