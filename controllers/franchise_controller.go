package controllers

import (
	"aquahome/database"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FranchiseDashboardData structure to hold dashboard response
type FranchiseDashboardData struct {
	Franchise              interface{} `json:"franchise"`
	Stats                  interface{} `json:"stats"`
	PendingOrders          interface{} `json:"pendingOrders"`
	PendingServiceRequests interface{} `json:"pendingServiceRequests"`
	RecentActivity         interface{} `json:"recentActivity"`
}

// ✅ GET /franchise/dashboard?franchiseId=xx
// ✅ GET /franchise/dashboard?franchiseId=xx
func GetFranchiseDashboard(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID := c.GetUint("userID") // ✅ safe and direct

	franchiseIDParam := c.Query("franchiseId")
	var franchiseID uint

	log.Println("🔍 Dashboard Fetching: Role =", role, "UserID =", userID, "FranchiseParam =", franchiseIDParam)

	if franchiseIDParam != "" {
		id, err := strconv.ParseUint(franchiseIDParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid franchise ID"})
			return
		}
		franchiseID = uint(id)
	} else {
		var user database.User
		if err := database.DB.First(&user, userID).Error; err != nil || user.FranchiseID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Franchise not found for user"})
			return
		}
		franchiseID = *user.FranchiseID
	}

	if role != "admin" {
		var f database.Franchise
		if err := database.DB.First(&f, franchiseID).Error; err != nil || f.OwnerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this dashboard"})
			return
		}
	}

	// 📊 Dashboard Stats
	var totalCustomers int64
	var totalOrders int64
	var activeSubscriptions int64
	var pendingServices int64

	database.DB.Model(&database.User{}).Where("franchise_id = ? AND role = ?", franchiseID, database.RoleCustomer).Count(&totalCustomers)
	database.DB.Model(&database.Order{}).Where("franchise_id = ?", franchiseID).Count(&totalOrders)
	database.DB.Model(&database.Subscription{}).Where("franchise_id = ? AND status = ?", franchiseID, "active").Count(&activeSubscriptions)
	database.DB.Model(&database.ServiceRequest{}).Where("franchise_id = ? AND status = ?", franchiseID, "pending").Count(&pendingServices)

	var pendingOrders []database.Order
	database.DB.Where("franchise_id = ? AND status = ?", franchiseID, "pending").Order("created_at DESC").Limit(5).Find(&pendingOrders)

	var pendingRequests []database.ServiceRequest
	database.DB.Where("franchise_id = ? AND status = ?", franchiseID, "pending").Order("created_at DESC").Limit(5).Find(&pendingRequests)

	var recentActivity []interface{} = []interface{}{} // optional

	var franchise database.Franchise
	if err := database.DB.First(&franchise, franchiseID).Error; err != nil {
		log.Printf("Franchise fetch error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch franchise info"})
		return
	}

	log.Println("✅ Dashboard returning for franchise:", franchiseID)

	c.JSON(http.StatusOK, FranchiseDashboardData{
		Franchise: franchise,
		Stats: gin.H{
			"totalCustomers":         totalCustomers,
			"totalOrders":            totalOrders,
			"activeSubscriptions":    activeSubscriptions,
			"pendingServiceRequests": pendingServices,
		},
		PendingOrders:          pendingOrders,
		PendingServiceRequests: pendingRequests,
		RecentActivity:         recentActivity,
	})
}

// ✅ GET /franchises - Admin Only
func GetAllFranchises(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var franchises []database.Franchise
	if err := database.DB.Order("created_at desc").Find(&franchises).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch franchises"})
		return
	}

	c.JSON(http.StatusOK, franchises)
}

// PATCH /franchises/:id - Admin updates franchise details
// PATCH /franchises/:id - Admin updates franchise details
func AdminUpdateFranchise(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid franchise ID"})
		return
	}

	var franchise database.Franchise
	if err := database.DB.First(&franchise, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Franchise not found"})
		return
	}

	var request struct {
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Email   string `json:"email"`
		City    string `json:"city"`
		State   string `json:"state"`
		ZipCode string `json:"zip_code"`
		Address string `json:"address"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Update fields
	franchise.Name = request.Name
	franchise.Phone = request.Phone
	franchise.Email = request.Email
	franchise.City = request.City
	franchise.State = request.State
	franchise.ZipCode = request.ZipCode
	franchise.Address = request.Address

	if err := database.DB.Save(&franchise).Error; err != nil {
		log.Printf("❌ Franchise update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Franchise updated successfully"})
}

// PATCH /admin/franchises/:id/toggle-status
func ToggleFranchiseStatus(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid franchise ID"})
		return
	}

	var input struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := database.DB.Model(&database.Franchise{}).
		Where("id = ?", id).
		Update("is_active", input.IsActive).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update franchise status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Franchise status updated"})
}
