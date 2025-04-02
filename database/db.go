package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"aquahome/config"
)

// InitDB initializes the database connection
func InitDB() error {
	// Set up GORM configuration
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Initialize based on driver type
	var err error
	switch config.AppConfig.DBDriver {
	case "postgres":
		// Check if DATABASE_URL is provided (Replit environment)
		dbURL := os.Getenv("DATABASE_URL")
		var dsn string

		if dbURL != "" {
			// Use the DATABASE_URL directly
			dsn = dbURL
			log.Println("Using DATABASE_URL environment variable for PostgreSQL connection")
		} else {
			// Construct the PostgreSQL connection string from individual parameters
			dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
				config.AppConfig.DBHost,
				config.AppConfig.DBPort,
				config.AppConfig.DBUser,
				config.AppConfig.DBPassword,
				config.AppConfig.DBName)

			// Log connection attempt (without password)
			log.Printf("Attempting to connect to PostgreSQL at host=%s port=%s user=%s dbname=%s",
				config.AppConfig.DBHost,
				config.AppConfig.DBPort,
				config.AppConfig.DBUser,
				config.AppConfig.DBName)
		}

		// Try to establish the PostgreSQL connection
		DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err != nil {
			log.Printf("Failed to connect to PostgreSQL database: %v", err)
			return err
		}
		log.Println("PostgreSQL database connection established successfully")

	case "sqlite", "sqlite3":
		// Ensure the directory exists
		dbDir := filepath.Dir(config.AppConfig.DBPath)
		if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
			log.Printf("Failed to create directory for SQLite database: %v", err)
			return err
		}

		// Try to establish the SQLite connection
		DB, err = gorm.Open(sqlite.Open(config.AppConfig.DBPath), gormConfig)
		if err != nil {
			log.Printf("Failed to connect to SQLite database: %v", err)
			return err
		}
		log.Printf("SQLite database connection established successfully at %s", config.AppConfig.DBPath)

	default:
		return fmt.Errorf("unsupported database driver: %s", config.AppConfig.DBDriver)
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
