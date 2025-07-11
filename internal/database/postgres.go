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



	return db
}

func Close(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
} 