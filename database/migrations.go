package database

import (
	"log"
)

// RunMigrations runs all database migrations
func RunMigrations() error {
	log.Println("Running database migrations...")

	// AutoMigrate will create tables if they don't exist
	if err := DB.AutoMigrate(
		&User{},
		&Product{},
		&Franchise{},
		&Order{},
		&Subscription{},
		&ServiceRequest{},
		&Payment{},
		&Notification{},
		&PasswordReset{},
		&Audit{},
		&AuditLog{},
	); err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}
