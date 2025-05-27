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
		&Location{},          // ✅ NEW: Service ZIP areas
		&FranchiseLocation{}, // ✅ NEW: Join table
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

// SeedDefaultAdmin creates a default admin if none exists
func SeedDefaultAdmin() {
	var count int64
	if err := DB.Model(&User{}).Where("role = ?", RoleAdmin).Count(&count).Error; err != nil {
		log.Printf("\u274c Failed to check existing admin: %v", err)
		return
	}

	if count == 0 {
		admin := User{
			Name:     "Super Admin",
			Email:    "admin@aquahome.com",
			Password: "admin123", //  You can hash it later
			Role:     RoleAdmin,
			Phone:    "9999999999",
			Address:  "Admin HQ",
			City:     "Hyderabad",
			State:    "Telangana",
			ZipCode:  "500001",
		}

		if err := DB.Create(&admin).Error; err != nil {
			log.Printf("\u274c Failed to create admin: %v", err)
		} else {
			log.Println("\u2705 Default admin user created successfully.")
		}
	} else {
		log.Println("ℹ️ Admin user already exists.")
	}
}
