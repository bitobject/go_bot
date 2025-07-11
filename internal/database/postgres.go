package database

import (
	"fmt"
	"go-bot/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func Init(cfg *config.Config) *gorm.DB {
	// DEBUG: выводим значения полей конфигурации БД
	fmt.Println("DEBUG DBHost:", cfg.DBHost)
	fmt.Println("DEBUG DBUser:", cfg.DBUser)
	fmt.Println("DEBUG DBPassword:", cfg.DBPassword)
	fmt.Println("DEBUG DBName:", cfg.DBName)
	fmt.Println("DEBUG DBPort:", cfg.DBPort)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic("failed to connect database:", err)
	}

	// Auto migrate models one by one to isolate the issue
	log.Println("Migrating User model...")
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Panicf("failed to migrate User model: %v", err)
	}

	log.Println("Migrating Message model...")
	if err := db.AutoMigrate(&Message{}); err != nil {
		log.Panicf("failed to migrate Message model: %v", err)
	}

	log.Println("Migrating Admin model...")
	if err := db.AutoMigrate(&Admin{}); err != nil {
		log.Panicf("failed to migrate Admin model: %v", err)
	}

	log.Println("Database migration successful")

	return db
}

func Close(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
} 