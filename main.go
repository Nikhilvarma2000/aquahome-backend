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
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Set PostgreSQL environment variables if available
	if os.Getenv("PGHOST") != "" {
		os.Setenv("DB_HOST", os.Getenv("PGHOST"))
	}
	if os.Getenv("PGPORT") != "" {
		os.Setenv("DB_PORT", os.Getenv("PGPORT"))
	}
	if os.Getenv("PGUSER") != "" {
		os.Setenv("DB_USER", os.Getenv("PGUSER"))
	}
	if os.Getenv("PGPASSWORD") != "" {
		os.Setenv("DB_PASSWORD", os.Getenv("PGPASSWORD"))
	}
	if os.Getenv("PGDATABASE") != "" {
		os.Setenv("DB_NAME", os.Getenv("PGDATABASE"))
	}

	// Initialize config
	config.InitConfig()

	// Setup router
	r := gin.Default()

	// CORS settings
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Initialize DBs
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize GORM database: %v", err)
	}
	if err := database.InitLegacyDB(); err != nil {
		log.Fatalf("Failed to initialize legacy database: %v", err)
	}

	// Run migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Setup routes (real AuthMiddleware is applied inside routes)
	routes.SetupRoutes(r)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	log.Printf("🚀 Server running at http://0.0.0.0:%s", port)
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
