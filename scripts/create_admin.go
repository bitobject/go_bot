package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"go-bot/internal/config"
	"go-bot/internal/database"
	"go-bot/internal/services"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем .env файл из корня проекта для локального запуска.
	// В Docker-окружении переменные будут переданы напрямую.
	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env file not found, relying on environment variables")
	}

	cfg := config.Get()

	db := database.Init(cfg)

	if len(os.Args) < 3 {
		fmt.Println("Usage: go run scripts/create_admin.go <login> <password>")
		os.Exit(1)
	}

	login := os.Args[1]
	password := os.Args[2]

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	adminService := services.NewAdminService(db, logger)

	admin, err := adminService.CreateAdmin(context.Background(), login, password)
	if err != nil {
		log.Fatalf("Error creating admin: %v", err)
	}

	fmt.Printf("Admin '%s' created successfully with ID: %d\n", admin.Login, admin.ID)
}