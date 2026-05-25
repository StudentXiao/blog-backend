package main

import (
	"blog-backend/internal/config"
	"blog-backend/internal/database"
	"blog-backend/internal/router"
	"fmt"
	"log"
)

func main() {
	// Load configuation
	config.LoadConfig()

	// Initialize database
	database.InitDB()

	// Setup router
	r := router.SetupRouter()

	// Start server
	addr := fmt.Sprintf(":%s", config.AppConfig.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
