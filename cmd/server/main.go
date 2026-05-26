package main

import (
	"blog-backend/internal/config"
	"blog-backend/internal/database"
	"blog-backend/internal/router"
	"fmt"
	"log"
)

// @title           你的 API 标题
// @version         1.0
// @description     这是 API 的描述信息
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name   Apache 2.0
// @license.url    http://www.apache.org/licenses/LICENSE-2.0.html

// @host           localhost:8080
// @BasePath       /api/v1

// @schemes        http https

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
