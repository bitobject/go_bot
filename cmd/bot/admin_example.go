package main

import (
	"fmt"
	"log"
	"go-bot/internal/config"
	"go-bot/internal/database"
)

// ExampleAdminUsage demonstrates how to use the AdminService
func ExampleAdminUsage(db *gorm.DB) {
	// Create admin service
	adminService := database.NewAdminService(db)

	// Example 1: Create a new admin
	fmt.Println("=== Creating new admin ===")
	admin, err := adminService.CreateAdmin("admin", "secure_password_123")
	if err != nil {
		log.Printf("Failed to create admin: %v", err)
	} else {
		fmt.Printf("Created admin with ID: %d, Login: %s\n", admin.ID, admin.Login)
	}

	// Example 2: Authenticate admin
	fmt.Println("\n=== Authenticating admin ===")
	authenticatedAdmin, err := adminService.AuthenticateAdmin("admin", "secure_password_123")
	if err != nil {
		log.Printf("Authentication failed: %v", err)
	} else {
		fmt.Printf("Successfully authenticated admin: %s\n", authenticatedAdmin.Login)
	}

	// Example 3: Generate secure password
	fmt.Println("\n=== Generating secure password ===")
	securePassword, err := database.GenerateSecurePassword(16)
	if err != nil {
		log.Printf("Failed to generate password: %v", err)
	} else {
		fmt.Printf("Generated secure password: %s\n", securePassword)
	}

	// Example 4: Update admin password
	fmt.Println("\n=== Updating admin password ===")
	err = adminService.UpdateAdminPassword(admin.ID, "new_secure_password_456")
	if err != nil {
		log.Printf("Failed to update password: %v", err)
	} else {
		fmt.Println("Password updated successfully")
	}

	// Example 5: Get admin by login
	fmt.Println("\n=== Getting admin by login ===")
	foundAdmin, err := adminService.GetAdminByLogin("admin")
	if err != nil {
		log.Printf("Failed to get admin: %v", err)
	} else {
		fmt.Printf("Found admin: ID=%d, Login=%s, Active=%t\n", 
			foundAdmin.ID, foundAdmin.Login, foundAdmin.IsActive)
	}
} 