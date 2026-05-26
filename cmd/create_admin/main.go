package main

import (
	"blog-backend/internal/config"
	"blog-backend/internal/database"
	"blog-backend/internal/models"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load config
	config.LoadConfig()

	// Initialize database
	database.InitDB()

	if len(os.Args) < 4 {
		fmt.Println("Usage: go run cmd/create_admin/main.go <username> <email> <password>")
		os.Exit(1)
	}

	username := os.Args[1]
	email := os.Args[2]
	password := os.Args[3]

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create admin user
	admin := &models.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Role:     "admin",
	}

	result := database.DB.Create(admin)
	if result.Error != nil {
		log.Fatal("Failed to create admin:", result.Error)
	}

	fmt.Printf("Admin user created successfully!\nID: %d\nUsername: %s\nEmail: %s\n", admin.ID, admin.Username, admin.Email)
}
