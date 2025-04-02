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

	// Initialize configuration
	config.InitConfig()

	// Set up the Gin router
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Set up routes
	routes.SetupRoutes(r)

	// Determine port for HTTP service
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000" // Default port for Replit compatibility
	}

	// Start the server
	log.Printf("Starting server on port %s...", port)
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
