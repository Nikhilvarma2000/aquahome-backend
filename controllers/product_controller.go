package controllers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"aquahome/database"
)

// ProductRequest contains the data for product creation or update
type ProductRequest struct {
	Name             string  `json:"name" binding:"required"`
	Description      string  `json:"description" binding:"required"`
	ImageURL         string  `json:"image_url"`
	MonthlyRent      float64 `json:"monthly_rent" binding:"required"`
	SecurityDeposit  float64 `json:"security_deposit" binding:"required"`
	InstallationFee  float64 `json:"installation_fee" binding:"required"`
	AvailableStock   int     `json:"available_stock" binding:"required"`
	Specifications   string  `json:"specifications"`
	MaintenanceCycle int     `json:"maintenance_cycle"`
	IsActive         bool    `json:"is_active"`
}

// CreateProduct creates a new product (Admin only)
func CreateProduct(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	var productRequest ProductRequest
	if err := c.ShouldBindJSON(&productRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if productRequest.MaintenanceCycle == 0 {
		productRequest.MaintenanceCycle = 90 // Default 90 days
	}

	// Create product with GORM
	product := database.Product{
		Name:             productRequest.Name,
		Description:      productRequest.Description,
		ImageURL:         productRequest.ImageURL,
		MonthlyRent:      productRequest.MonthlyRent,
		SecurityDeposit:  productRequest.SecurityDeposit,
		InstallationFee:  productRequest.InstallationFee,
		AvailableStock:   productRequest.AvailableStock,
		Specifications:   productRequest.Specifications,
		MaintenanceCycle: productRequest.MaintenanceCycle,
		IsActive:         productRequest.IsActive,
	}

	result := database.DB.Create(&product)
	if result.Error != nil {
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GetProducts gets all active products
func GetProducts(c *gin.Context) {
	activeOnly := true
	roleInterface, exists := c.Get("role")
	if exists {
		role := roleInterface.(string)
		if role == "admin" {
			// Admin can see all products, including inactive ones
			activeParam := c.DefaultQuery("active", "true")
			activeOnly = activeParam == "true"
		}
	}

	var products []database.Product
	query := database.DB

	// Apply filter for active products if needed
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	// Execute query
	result := query.Find(&products)
	if result.Error != nil {
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetProductByID gets a product by ID
func GetProductByID(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product database.Product
	result := database.DB.First(&product, productID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// If the product is inactive, only admins can see it
	if !product.IsActive {
		roleInterface, exists := c.Get("role")
		if !exists || roleInterface.(string) != "admin" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
	}

	c.JSON(http.StatusOK, product)
}

// UpdateProduct updates a product (Admin only)
func UpdateProduct(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var productRequest ProductRequest
	if err := c.ShouldBindJSON(&productRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Find the product
	var product database.Product
	result := database.DB.First(&product, productID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Update product with GORM
	product.Name = productRequest.Name
	product.Description = productRequest.Description
	product.ImageURL = productRequest.ImageURL
	product.MonthlyRent = productRequest.MonthlyRent
	product.SecurityDeposit = productRequest.SecurityDeposit
	product.InstallationFee = productRequest.InstallationFee
	product.AvailableStock = productRequest.AvailableStock
	product.Specifications = productRequest.Specifications
	product.MaintenanceCycle = productRequest.MaintenanceCycle
	product.IsActive = productRequest.IsActive

	result = database.DB.Save(&product)
	if result.Error != nil {
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating product"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct sets a product as inactive (Admin only)
func DeleteProduct(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// Find the product
	var product database.Product
	result := database.DB.First(&product, productID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Set product as inactive with GORM
	product.IsActive = false
	result = database.DB.Save(&product)
	if result.Error != nil {
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
