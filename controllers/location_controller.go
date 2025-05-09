package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"aquahome/database"
)

// FranchiseWithOwner represents a franchise with owner details
type FranchiseWithOwner struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	OwnerName     string `json:"owner_name"`
	OwnerEmail    string `json:"owner_email"`
	OwnerPhone    string `json:"owner_phone"`
	Address       string `json:"address"`
	City          string `json:"city"`
	State         string `json:"state"`
	ZipCode       string `json:"zip_code"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	AreaPolygon   string `json:"area_polygon"`
	IsActive      bool   `json:"is_active"`
	ApprovalState string `json:"approval_state"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// FranchiseRequest contains data for franchise creation or update
type FranchiseRequest struct {
	Name        string          `json:"name" binding:"required"`
	Address     string          `json:"address" binding:"required"`
	City        string          `json:"city" binding:"required"`
	State       string          `json:"state" binding:"required"`
	ZipCode     string          `json:"zip_code" binding:"required"`
	Phone       string          `json:"phone" binding:"required"`
	Email       string          `json:"email" binding:"required,email"`
	AreaPolygon json.RawMessage `json:"area_polygon" binding:"required"`
}

// CreateFranchise creates a new franchise (Franchise Owner only)
func CreateFranchise(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || (role != "franchise_owner" && role != "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	userID, _ := c.Get("userID")
	ownerID := uint(userID.(float64)) // Convert to uint for GORM

	var franchiseRequest FranchiseRequest
	if err := c.ShouldBindJSON(&franchiseRequest); err != nil {
		log.Printf("Invalid request data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Validate GeoJSON polygon
	if !isValidGeoJSONPolygon(franchiseRequest.AreaPolygon) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid GeoJSON polygon format"})
		return
	}

	// Check for overlapping service areas
	if overlapsExistingFranchise(franchiseRequest.AreaPolygon) {
		c.JSON(http.StatusConflict, gin.H{"error": "Your service area overlaps with an existing franchise"})
		return
	}

	// Begin transaction
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("Transaction error: %v", tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Create franchise using GORM
	franchise := database.Franchise{
		Name:          franchiseRequest.Name,
		OwnerID:       ownerID,
		Address:       franchiseRequest.Address,
		City:          franchiseRequest.City,
		State:         franchiseRequest.State,
		ZipCode:       franchiseRequest.ZipCode,
		Phone:         franchiseRequest.Phone,
		Email:         franchiseRequest.Email,
		AreaPolygon:   string(franchiseRequest.AreaPolygon),
		IsActive:      false,     // Initially inactive until approved
		ApprovalState: "pending", // Initial approval state
	}

	result := tx.Create(&franchise)
	if result.Error != nil {
		tx.Rollback()
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating franchise"})
		return
	}

	franchiseID := franchise.ID

	// Create notification for franchise owner
	ownerNotification := database.Notification{
		UserID:      ownerID,
		Title:       "Franchise Application Submitted",
		Message:     "Your franchise application for " + franchiseRequest.Name + " has been submitted and is pending approval.",
		Type:        "franchise",
		RelatedID:   &franchise.ID,
		RelatedType: "franchise",
	}

	result = tx.Create(&ownerNotification)
	if result.Error != nil {
		tx.Rollback()
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating owner notification"})
		return
	}

	// Create notification for admin
	// First, find an admin user to notify
	var adminUser database.User
	adminResult := database.DB.Where("role = ?", database.RoleAdmin).First(&adminUser)

	if adminResult.Error == nil {
		adminNotification := database.Notification{
			UserID:      adminUser.ID,
			Title:       "New Franchise Application",
			Message:     "A new franchise application has been submitted by " + franchiseRequest.Name + " and requires your approval.",
			Type:        "franchise",
			RelatedID:   &franchise.ID,
			RelatedType: "franchise",
		}

		if err := tx.Create(&adminNotification).Error; err != nil {
			log.Printf("Error creating admin notification: %v", err)
			// Don't roll back for this error, it's not critical
		}
	}

	// Update user with franchise_id
	var user database.User
	if err := tx.First(&user, ownerID).Error; err != nil {
		tx.Rollback()
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding user"})
		return
	}

	user.FranchiseID = &franchise.ID
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user with franchise ID"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Transaction commit error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Franchise application submitted successfully. It is pending approval.",
		"id":      franchiseID,
	})
}

// GetFranchises gets all franchises based on user role
func GetFranchises(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, _ := c.Get("userID")
	var userIDUint uint

	if role != "admin" {
		userIDUint = uint(userID.(float64))
	}

	// Define the response structure
	// Using the already defined FranchiseWithOwner struct

	var franchises []FranchiseWithOwner

	query := database.DB.Table("franchises").
		Select(`
			franchises.id, 
			franchises.name, 
			franchises.address, 
			franchises.city, 
			franchises.state, 
			franchises.zip_code, 
			franchises.phone, 
			franchises.email, 
			franchises.area_polygon, 
			franchises.is_active, 
			franchises.approval_state, 
			franchises.created_at, 
			franchises.updated_at, 
			users.name as owner_name, 
			users.email as owner_email, 
			users.phone as owner_phone
		`).
		Joins("JOIN users ON franchises.owner_id = users.id").
		Order("franchises.created_at DESC")

	// Apply role-based filtering
	switch role {
	case "admin":
		// Admin can see all franchises - no additional filters
	case "franchise_owner":
		// Franchise owner can only see their own franchises
		query = query.Where("franchises.owner_id = ?", userIDUint)
	default:
		// Other roles can only see active franchises
		query = query.Where("franchises.is_active = ? AND franchises.approval_state = ?", true, "approved")
	}

	// Execute the query
	result := query.Find(&franchises)
	if result.Error != nil {
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Process AreaPolygon to ensure it's proper JSON format for the response
	for i := range franchises {
		if franchises[i].AreaPolygon != "" {
			franchises[i].AreaPolygon = string(json.RawMessage(franchises[i].AreaPolygon))
		}
	}

	c.JSON(http.StatusOK, franchises)
}

// GetFranchiseByID gets a franchise by ID
func PublicGetFranchiseByID(c *gin.Context) {
	franchiseIDStr := c.Param("id")
	franchiseID, err := strconv.ParseUint(franchiseIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid franchise ID"})
		return
	}

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, _ := c.Get("userID")
	userIDUint := uint(userID.(float64))

	// Define response structure using FranchiseWithOwner and adding missing fields
	type FranchiseDetail struct {
		database.Franchise
		OwnerName string `json:"owner_name"`
	}

	var franchise FranchiseDetail

	// Create base query
	query := database.DB.Table("franchises").
		Select("franchises.*, users.name as owner_name").
		Joins("JOIN users ON franchises.owner_id = users.id").
		Where("franchises.id = ?", franchiseID)

	// Apply role-based conditions
	switch role {
	case "admin":
		// Admin can see any franchise - no additional filters
	case "franchise_owner":
		// Franchise owner can only see their own franchises
		query = query.Where("franchises.owner_id = ?", userIDUint)
	default:
		// Other roles can only see active franchises
		query = query.Where("franchises.is_active = ? AND franchises.approval_state = ?", true, "approved")
	}

	// Execute query
	result := query.First(&franchise)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Franchise not found or you don't have permission to view it"})
			return
		}
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Ensure area_polygon is properly formatted JSON
	if franchise.AreaPolygon != "" {
		franchise.Franchise.AreaPolygon = string(json.RawMessage(franchise.Franchise.AreaPolygon))
	}

	// Get statistics if admin or franchise owner
	if role == "admin" || (role == "franchise_owner" && franchise.OwnerID == userIDUint) {
		var activeSubscriptions int64
		var pendingServices int64

		// Get active subscriptions count
		database.DB.Model(&database.Subscription{}).
			Where("franchise_id = ? AND status = ?", franchiseID, database.SubscriptionStatusActive).
			Count(&activeSubscriptions)

		// Get pending service requests count
		database.DB.Model(&database.ServiceRequest{}).
			Where("franchise_id = ? AND status IN (?, ?)",
				franchiseID, database.ServiceStatusPending, database.ServiceStatusScheduled).
			Count(&pendingServices)

		// Return franchise with statistics
		c.JSON(http.StatusOK, gin.H{
			"franchise": franchise,
			"stats": gin.H{
				"active_subscriptions": activeSubscriptions,
				"pending_services":     pendingServices,
			},
		})
		return
	}

	c.JSON(http.StatusOK, franchise)
}

// UpdateFranchise updates a franchise (Franchise Owner or Admin only)
func UpdateFranchise(c *gin.Context) {
	franchiseIDStr := c.Param("id")
	franchiseID, err := strconv.ParseUint(franchiseIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid franchise ID"})
		return
	}

	role, exists := c.Get("role")
	if !exists || (role != "admin" && role != "franchise_owner") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	userID, _ := c.Get("userID")
	userIDUint := uint(userID.(float64))

	// Find franchise to check existence and ownership
	var franchise database.Franchise
	result := database.DB.First(&franchise, franchiseID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Franchise not found"})
			return
		}
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// If franchise owner, check if they own the franchise
	if role == "franchise_owner" && franchise.OwnerID != userIDUint {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this franchise"})
		return
	}

	var franchiseRequest FranchiseRequest
	if err := c.ShouldBindJSON(&franchiseRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Validate GeoJSON polygon
	if !isValidGeoJSONPolygon(franchiseRequest.AreaPolygon) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid GeoJSON polygon format"})
		return
	}

	// Check for overlapping service areas (excluding this franchise)
	if overlapsOtherFranchise(franchiseRequest.AreaPolygon, int64(franchiseID)) {
		c.JSON(http.StatusConflict, gin.H{"error": "Your service area overlaps with another franchise"})
		return
	}

	// Update franchise fields
	franchise.Name = franchiseRequest.Name
	franchise.Address = franchiseRequest.Address
	franchise.City = franchiseRequest.City
	franchise.State = franchiseRequest.State
	franchise.ZipCode = franchiseRequest.ZipCode
	franchise.Phone = franchiseRequest.Phone
	franchise.Email = franchiseRequest.Email
	franchise.AreaPolygon = string(franchiseRequest.AreaPolygon)

	// If franchise owner is resubmitting a rejected application, update approval state
	if role == "franchise_owner" && franchise.ApprovalState == "rejected" {
		franchise.ApprovalState = "pending"
	}

	// Save changes
	result = database.DB.Save(&franchise)
	if result.Error != nil {
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating franchise"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Franchise updated successfully"})
}

// ApproveFranchise approves a franchise application (Admin only)
func ApproveFranchise(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	franchiseIDStr := c.Param("id")
	franchiseID, err := strconv.ParseUint(franchiseIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid franchise ID"})
		return
	}

	// Find franchise to check existence and status
	var franchise database.Franchise
	result := database.DB.First(&franchise, franchiseID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Franchise not found"})
			return
		}
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	if franchise.ApprovalState == "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Franchise is already approved"})
		return
	}

	// Begin transaction
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("Transaction error: %v", tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Update franchise status
	franchise.ApprovalState = "approved"
	franchise.IsActive = true

	if err := tx.Save(&franchise).Error; err != nil {
		tx.Rollback()
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error approving franchise"})
		return
	}

	// Create notification for franchise owner
	notification := database.Notification{
		UserID:      franchise.OwnerID,
		Title:       "Franchise Application Approved",
		Message:     "Your franchise application has been approved. You can now start serving customers.",
		Type:        "franchise",
		RelatedID:   &franchise.ID,
		RelatedType: "franchise",
	}

	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating notification"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Transaction commit error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Franchise approved successfully"})
}

// RejectFranchise rejects a franchise application (Admin only)
func RejectFranchise(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	franchiseIDStr := c.Param("id")
	franchiseID, err := strconv.ParseUint(franchiseIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid franchise ID"})
		return
	}

	type RejectRequest struct {
		Reason string `json:"reason" binding:"required"`
	}

	var rejectRequest RejectRequest
	if err := c.ShouldBindJSON(&rejectRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reason for rejection is required"})
		return
	}

	// Find franchise to check existence and status
	var franchise database.Franchise
	result := database.DB.First(&franchise, franchiseID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Franchise not found"})
			return
		}
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	if franchise.ApprovalState == "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Franchise is already rejected"})
		return
	}

	// Begin transaction
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("Transaction error: %v", tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Update franchise status
	franchise.ApprovalState = "rejected"
	franchise.IsActive = false

	if err := tx.Save(&franchise).Error; err != nil {
		tx.Rollback()
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error rejecting franchise"})
		return
	}

	// Create notification for franchise owner
	notification := database.Notification{
		UserID:      franchise.OwnerID,
		Title:       "Franchise Application Rejected",
		Message:     "Your franchise application has been rejected. Reason: " + rejectRequest.Reason,
		Type:        "franchise",
		RelatedID:   &franchise.ID,
		RelatedType: "franchise",
	}

	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating notification"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Transaction commit error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Franchise rejected successfully"})
}

// Helper function to validate GeoJSON polygon
func isValidGeoJSONPolygon(polygon json.RawMessage) bool {
	var geoJSON map[string]interface{}
	if err := json.Unmarshal(polygon, &geoJSON); err != nil {
		log.Printf("Invalid GeoJSON: %v", err)
		return false
	}

	// Check if it's a valid GeoJSON
	if geoJSON["type"] != "Polygon" && geoJSON["type"] != "MultiPolygon" {
		log.Printf("Invalid GeoJSON type: %v", geoJSON["type"])
		return false
	}

	// Check if coordinates exist
	if _, ok := geoJSON["coordinates"]; !ok {
		log.Printf("GeoJSON missing coordinates")
		return false
	}

	return true
}

// Helper function to check if a polygon overlaps with any existing franchise
func overlapsExistingFranchise(polygon json.RawMessage) bool {
	// This is a simplified check. In a real app, you'd use spatial extensions or libraries.
	// For SQLite, you might want to use SpatiaLite extension for proper spatial queries.

	// For now, let's assume no overlap to allow application to work
	log.Printf("Checking for overlapping franchises (not implemented)")
	return false
}

// Helper function to check if a polygon overlaps with other franchises (excluding the given one)
func overlapsOtherFranchise(polygon json.RawMessage, excludeFranchiseID int64) bool {
	// Similar to above, this is a simplified check
	log.Printf("Checking for overlapping franchises excluding ID %d (not implemented)", excludeFranchiseID)
	return false
}

// GetFranchiseServiceAgents gets service agents associated with a franchise
func GetFranchiseServiceAgents(c *gin.Context) {
	franchiseIDStr := c.Param("id")
	franchiseID, err := strconv.ParseUint(franchiseIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid franchise ID"})
		return
	}

	role, exists := c.Get("role")
	if !exists || (role != "admin" && role != "franchise_owner") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	userID, _ := c.Get("userID")
	userIDUint := uint(userID.(float64))

	// If franchise owner, check if they own the franchise
	if role == "franchise_owner" {
		var franchise database.Franchise
		result := database.DB.Select("owner_id").First(&franchise, franchiseID)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Franchise not found"})
				return
			}
			log.Printf("Database error: %v", result.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}

		if franchise.OwnerID != userIDUint {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this franchise's service agents"})
			return
		}
	}

	// Define response structure for service agents
	type ServiceAgentInfo struct {
		ID             uint   `json:"id"`
		Name           string `json:"name"`
		Email          string `json:"email"`
		Phone          string `json:"phone"`
		ProfilePicture string `json:"profile_picture"`
	}

	var serviceAgents []ServiceAgentInfo

	// Get service agents for the franchise using GORM
	result := database.DB.Model(&database.User{}).
		Select("id, name, email, phone, profile_picture").
		Where("franchise_id = ? AND role = ?", franchiseID, database.RoleServiceAgent).
		Find(&serviceAgents)

	if result.Error != nil {
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	c.JSON(http.StatusOK, serviceAgents)
}

// SearchFranchises searches for franchises by location (Customer only)
func SearchFranchises(c *gin.Context) {
	// This is a simplified search by zip code
	// In a real app, you'd use spatial queries to find franchises serving the customer's location

	zipCode := c.Query("zip_code")
	if zipCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zip code is required"})
		return
	}

	// Define response structure
	type FranchiseLocation struct {
		ID      uint   `json:"id"`
		Name    string `json:"name"`
		Address string `json:"address"`
		City    string `json:"city"`
		State   string `json:"state"`
		ZipCode string `json:"zip_code"`
	}

	var franchises []FranchiseLocation

	// Get franchises that serve this zip code using GORM
	result := database.DB.Model(&database.Franchise{}).
		Select("id, name, address, city, state, zip_code").
		Where("is_active = ? AND approval_state = ? AND zip_code = ?", true, "approved", zipCode).
		Find(&franchises)

	if result.Error != nil {
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	c.JSON(http.StatusOK, franchises)
}
