package controllers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"aquahome/database"
	"aquahome/utils"
)

// GetUserProfile returns the profile of the authenticated user
func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user database.User
	query := `SELECT id, name, email, phone, role, address, franchise_id, created_at, updated_at FROM users WHERE id = $1`
	err := database.LegacyDB.QueryRow(query, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.Role, &user.Address, &user.FranchiseID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfileRequest contains the data for profile update
type UpdateProfileRequest struct {
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	Address        string `json:"address"`
	ProfilePicture string `json:"profile_picture"`
}

// UpdateUserProfile updates the profile of the authenticated user
func UpdateUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var updateRequest UpdateProfileRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update user profile
	query := `
                UPDATE users 
                SET name = COALESCE(?, name),
                        phone = COALESCE(?, phone),
                        address = COALESCE(?, address),
                        profile_picture = COALESCE(?, profile_picture),
                        updated_at = CURRENT_TIMESTAMP
                WHERE id = ?
        `
	_, err := database.LegacyDB.Exec(
		query,
		nullIfEmpty(updateRequest.Name),
		nullIfEmpty(updateRequest.Phone),
		nullIfEmpty(updateRequest.Address),
		nullIfEmpty(updateRequest.Address), // Replace ProfilePicture with Address
		userID,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating profile"})
		return
	}

	// Get updated user profile
	var user database.User
	getQuery := `SELECT id, name, email, phone, role, address, franchise_id, created_at, updated_at FROM users WHERE id = $1`
	err = database.LegacyDB.QueryRow(getQuery, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.Role, &user.Address, &user.FranchiseID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving updated profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ChangePasswordRequest contains the data for password change
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword changes the user's password
func ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var changePassRequest ChangePasswordRequest
	if err := c.ShouldBindJSON(&changePassRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Get current password hash
	var passwordHash string
	err := database.LegacyDB.QueryRow("SELECT password_hash FROM users WHERE id = $1", userID).Scan(&passwordHash)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Verify current password
	if !utils.CheckPasswordHash(changePassRequest.CurrentPassword, passwordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(changePassRequest.NewPassword)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing request"})
		return
	}

	// Update password
	_, err = database.LegacyDB.Exec(
		"UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		newPasswordHash, userID,
	)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// GetUserByID gets user details by ID (Admin only)
func GetUserByID(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user database.User
	query := `SELECT id, name, email, phone, role, address, franchise_id, created_at, updated_at FROM users WHERE id = $1`
	err = database.LegacyDB.QueryRow(query, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.Role, &user.Address, &user.FranchiseID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUsersByRole gets users by role (Admin only)
func GetUsersByRole(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	userRole := c.Param("role")
	if userRole == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role parameter is required"})
		return
	}

	rows, err := database.LegacyDB.Query(
		`SELECT id, name, email, phone, role, address, franchise_id, created_at, updated_at 
                FROM users WHERE role = $1`,
		userRole,
	)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}
	defer rows.Close()

	var users []database.User
	for rows.Next() {
		var user database.User
		err := rows.Scan(
			&user.ID, &user.Name, &user.Email, &user.Phone, &user.Role, &user.Address, &user.FranchiseID, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// Helper function to return nil for empty strings (for SQL NULL values)
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
