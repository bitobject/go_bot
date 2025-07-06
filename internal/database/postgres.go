package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"log"
)

func Init() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic("failed to connect database:", err)
	}
	return db
}

func Close(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
} 