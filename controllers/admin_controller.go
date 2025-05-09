package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"aquahome/database"
)

// AdminDashboard returns key statistics for the admin dashboard
func AdminDashboard(c *gin.Context) {
	var totalCustomers int64
	var totalOrders int64

	// Count customers with role 'customer'
	if err := database.DB.Model(&database.User{}).Where("role = ?", "customer").Count(&totalCustomers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count customers"})
		return
	}

	// Count total orders
	if err := database.DB.Model(&database.Order{}).Count(&totalOrders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count orders"})
		return
	}

	// Return simplified dashboard data
	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"totalCustomers":         totalCustomers,
			"totalOrders":            totalOrders,
			"totalRevenue":           0, // Optional: implement if needed
			"activeSubscriptions":    0,
			"pendingServiceRequests": 0,
			"franchiseApplications":  0,
		},
	})
}

// AdminGetOrders returns all orders with related data
func AdminGetOrders(c *gin.Context) {
	var orders []database.Order

	if err := database.DB.Preload("Customer").
		Preload("Franchise").
		//Preload("OrderItems.Product").
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}
