package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"aquahome/config"
)

// LegacyDB is the global database instance for raw SQL operations
var LegacyDB *sql.DB

// InitLegacyDB initializes the legacy database connection for raw SQL operations
func InitLegacyDB() error {
	var err error

	switch config.AppConfig.DBDriver {
	case "postgres":
		// Check if DATABASE_URL is provided (Replit environment)
		dbURL := os.Getenv("DATABASE_URL")
		var connStr string

		if dbURL != "" {
			// Use the DATABASE_URL directly
			connStr = dbURL
			log.Println("Using DATABASE_URL environment variable for PostgreSQL legacy connection")
		} else {
			// Construct the PostgreSQL connection string from individual parameters
			connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				config.AppConfig.DBHost,
				config.AppConfig.DBPort,
				config.AppConfig.DBUser,
				config.AppConfig.DBPassword,
				config.AppConfig.DBName)

			// Log connection attempt (without password)
			log.Printf("Attempting to connect to PostgreSQL legacy DB at host=%s port=%s user=%s dbname=%s",
				config.AppConfig.DBHost,
				config.AppConfig.DBPort,
				config.AppConfig.DBUser,
				config.AppConfig.DBName)
		}

		LegacyDB, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Failed to connect to PostgreSQL legacy database: %v", err)
			return err
		}

		log.Println("PostgreSQL legacy database connection established successfully")

	case "sqlite", "sqlite3":
		// Ensure the directory exists
		dbDir := filepath.Dir(config.AppConfig.DBPath)
		if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
			log.Printf("Failed to create directory for SQLite database: %v", err)
			return err
		}

		// Open SQLite connection
		LegacyDB, err = sql.Open("sqlite3", config.AppConfig.DBPath)
		if err != nil {
			log.Printf("Failed to connect to SQLite legacy database: %v", err)
			return err
		}

		// Enable foreign key constraints for SQLite
		_, err = LegacyDB.Exec("PRAGMA foreign_keys = ON")
		if err != nil {
			log.Printf("Failed to enable foreign keys in SQLite: %v", err)
			return err
		}

		log.Printf("SQLite legacy database connection established successfully at %s", config.AppConfig.DBPath)

	default:
		return fmt.Errorf("unsupported database driver: %s", config.AppConfig.DBDriver)
	}

	// Test the connection
	err = LegacyDB.Ping()
	if err != nil {
		log.Printf("Failed to ping legacy database: %v", err)
		return err
	}

	return nil
}

// CloseLegacyDB closes the legacy database connection
func CloseLegacyDB() error {
	if LegacyDB != nil {
		return LegacyDB.Close()
	}
	return nil
}
