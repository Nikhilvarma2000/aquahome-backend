package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"aquahome/config"
	"aquahome/database"
	"aquahome/routes"
)

func main() {
	// Load environment variables (optional for local dev)
	_ = godotenv.Load()

	// Initialize config
	config.InitConfig()

	// Initialize DBs
	if err := database.InitDB(); err != nil {
		log.Fatalf("❌ Failed to initialize GORM database: %v", err)
	}
	if err := database.InitLegacyDB(); err != nil {
		log.Fatalf("❌ Failed to initialize legacy database: %v", err)
	}

	// Run migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}

	// Setup router
	r := gin.Default()

	// Enable CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Register routes
	routes.SetupRoutes(r)
	for _, route := range r.Routes() {
		log.Printf("🔗 %s %s", route.Method, route.Path)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000" // fallback for local dev
	}
	log.Printf("🚀 Server running at http://0.0.0.0:%s", port)
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
