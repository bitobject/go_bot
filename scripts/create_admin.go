package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"go-bot/internal/services"
	"go-bot/internal/database"
)

func main() {
	// Парсим аргументы командной строки
	login := flag.String("login", "", "Admin login (required)")
	password := flag.String("password", "", "Admin password (required)")
	flag.Parse()

	if *login == "" || *password == "" {
		fmt.Println("Usage: go run scripts/create_admin.go -login=admin -password=secure_password")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Загружаем и валидируем конфигурацию. 
	// Это необходимо, чтобы database.Init() мог получить доступ к данным для подключения.
	cfg := config.Get()

	// Инициализируем базу данных
	db := database.Init(cfg)
	defer database.Close(db)

	// Создаем сервис администраторов
	adminService := database.NewAdminService(db)

	// Проверяем, существует ли уже администратор с таким логином
	existingAdmin, err := adminService.GetAdminByLogin(*login)
	if err == nil {
		fmt.Printf("Admin with login '%s' already exists (ID: %d)\n", existingAdmin.Login, existingAdmin.ID)
		os.Exit(1)
	}

	// Создаем нового администратора
	admin, err := adminService.CreateAdmin(*login, *password)
	if err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}

	fmt.Printf("Admin created successfully!\n")
	fmt.Printf("ID: %d\n", admin.ID)
	fmt.Printf("Login: %s\n", admin.Login)
	fmt.Printf("Created at: %s\n", admin.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("\nYou can now login using:\n")
	fmt.Printf("curl -X POST http://localhost:8080/api/admin/login \\\n")
	fmt.Printf("  -H \"Content-Type: application/json\" \\\n")
	fmt.Printf("  -d '{\"login\": \"%s\", \"password\": \"%s\"}'\n", *login, *password)
} 