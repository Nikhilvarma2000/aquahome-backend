// Package database provides database connection and models
package database

import (
	"gorm.io/gorm"
)

// DB is the global database instance
var DB *gorm.DB

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
