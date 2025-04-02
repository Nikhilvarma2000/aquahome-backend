package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"aquahome/database"
)

// ServiceRequestCreateRequest contains data for creating a service request
type ServiceRequestCreateRequest struct {
	SubscriptionID int64  `json:"subscription_id" binding:"required"`
	RequestType    string `json:"request_type" binding:"required"`
	Description    string `json:"description" binding:"required"`
}

// ServiceRequestUpdateRequest contains data for updating a service request
type ServiceRequestUpdateRequest struct {
	Status         string `json:"status"`
	AgentID        uint   `json:"agent_id"`
	ScheduledDate  string `json:"scheduled_date"`
	CompletionDate string `json:"completion_date"`
	Notes          string `json:"notes"`
}

// FeedbackRequest contains feedback data for a completed service
type FeedbackRequest struct {
	Rating   int    `json:"rating" binding:"required,min=1,max=5"`
	Feedback string `json:"feedback" binding:"required"`
}

// GetServiceRequests returns service requests based on user role
func GetServiceRequests(c *gin.Context) {
	userID := c.GetString("user_id")
	userIDInt, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	role := c.GetString("role")

	type ServiceRequestWithDetails struct {
		ID               uint       `json:"id"`
		Type             string     `json:"type"`
		Status           string     `json:"status"`
		Description      string     `json:"description"`
		ScheduledTime    *time.Time `json:"scheduled_time"`
		CompletionTime   *time.Time `json:"completion_time"`
		Rating           *int       `json:"rating"`
		Feedback         string     `json:"feedback"`
		CreatedAt        time.Time  `json:"created_at"`
		UpdatedAt        time.Time  `json:"updated_at"`
		CustomerID       uint       `json:"customer_id"`
		CustomerName     string     `json:"customer_name"`
		CustomerEmail    string     `json:"customer_email"`
		CustomerPhone    string     `json:"customer_phone"`
		ProductID        uint       `json:"product_id"`
		ProductName      string     `json:"product_name"`
		SubscriptionID   uint       `json:"subscription_id"`
		FranchiseID      *uint      `json:"franchise_id"`
		FranchiseName    string     `json:"franchise_name"`
		ServiceAgentID   *uint      `json:"service_agent_id"`
		ServiceAgentName string     `json:"service_agent_name"`
	}

	var results []ServiceRequestWithDetails

	switch role {
	case database.RoleAdmin:
		// Admin can see all service requests
		err = database.DB.Table("service_requests").
			Select(`
                service_requests.id,
                service_requests.type,
                service_requests.status,
                service_requests.description,
                service_requests.scheduled_time,
                service_requests.completion_time,
                service_requests.rating,
                service_requests.feedback,
                service_requests.created_at,
                service_requests.updated_at,
                service_requests.customer_id,
                customer.name as customer_name,
                customer.email as customer_email,
                customer.phone as customer_phone,
                subscriptions.product_id,
                products.name as product_name,
                service_requests.subscription_id,
                franchises.id as franchise_id,
                franchises.name as franchise_name,
                service_requests.service_agent_id,
                service_agent.name as service_agent_name
            `).
			Joins("JOIN users as customer ON service_requests.customer_id = customer.id").
			Joins("JOIN subscriptions ON service_requests.subscription_id = subscriptions.id").
			Joins("JOIN products ON subscriptions.product_id = products.id").
			Joins("LEFT JOIN franchises ON subscriptions.franchise_id = franchises.id").
			Joins("LEFT JOIN users as service_agent ON service_requests.service_agent_id = service_agent.id").
			Order("service_requests.created_at DESC").
			Find(&results).Error

	case database.RoleFranchiseOwner:
		// Franchise owner can see service requests assigned to their franchise
		err = database.DB.Table("service_requests").
			Select(`
                service_requests.id,
                service_requests.type,
                service_requests.status,
                service_requests.description,
                service_requests.scheduled_time,
                service_requests.completion_time,
                service_requests.rating,
                service_requests.feedback,
                service_requests.created_at,
                service_requests.updated_at,
                service_requests.customer_id,
                customer.name as customer_name,
                customer.email as customer_email,
                customer.phone as customer_phone,
                subscriptions.product_id,
                products.name as product_name,
                service_requests.subscription_id,
                franchises.id as franchise_id,
                franchises.name as franchise_name,
                service_requests.service_agent_id,
                service_agent.name as service_agent_name
            `).
			Joins("JOIN users as customer ON service_requests.customer_id = customer.id").
			Joins("JOIN subscriptions ON service_requests.subscription_id = subscriptions.id").
			Joins("JOIN products ON subscriptions.product_id = products.id").
			Joins("JOIN franchises ON subscriptions.franchise_id = franchises.id").
			Joins("LEFT JOIN users as service_agent ON service_requests.service_agent_id = service_agent.id").
			Where("franchises.owner_id = ?", userIDInt).
			Order("service_requests.created_at DESC").
			Find(&results).Error

	case database.RoleServiceAgent:
		// Service agent can see service requests assigned to them
		err = database.DB.Table("service_requests").
			Select(`
                service_requests.id,
                service_requests.type,
                service_requests.status,
                service_requests.description,
                service_requests.scheduled_time,
                service_requests.completion_time,
                service_requests.rating,
                service_requests.feedback,
                service_requests.created_at,
                service_requests.updated_at,
                service_requests.customer_id,
                customer.name as customer_name,
                customer.email as customer_email,
                customer.phone as customer_phone,
                subscriptions.product_id,
                products.name as product_name,
                service_requests.subscription_id,
                franchises.id as franchise_id,
                franchises.name as franchise_name,
                service_requests.service_agent_id,
                service_agent.name as service_agent_name
            `).
			Joins("JOIN users as customer ON service_requests.customer_id = customer.id").
			Joins("JOIN subscriptions ON service_requests.subscription_id = subscriptions.id").
			Joins("JOIN products ON subscriptions.product_id = products.id").
			Joins("LEFT JOIN franchises ON subscriptions.franchise_id = franchises.id").
			Joins("LEFT JOIN users as service_agent ON service_requests.service_agent_id = service_agent.id").
			Where("service_requests.service_agent_id = ?", userIDInt).
			Order("service_requests.created_at DESC").
			Find(&results).Error

	case database.RoleCustomer:
		// Customer can see their own service requests
		err = database.DB.Table("service_requests").
			Select(`
                service_requests.id,
                service_requests.type,
                service_requests.status,
                service_requests.description,
                service_requests.scheduled_time,
                service_requests.completion_time,
                service_requests.rating,
                service_requests.feedback,
                service_requests.created_at,
                service_requests.updated_at,
                service_requests.customer_id,
                customer.name as customer_name,
                customer.email as customer_email,
                customer.phone as customer_phone,
                subscriptions.product_id,
                products.name as product_name,
                service_requests.subscription_id,
                franchises.id as franchise_id,
                franchises.name as franchise_name,
                service_requests.service_agent_id,
                service_agent.name as service_agent_name
            `).
			Joins("JOIN users as customer ON service_requests.customer_id = customer.id").
			Joins("JOIN subscriptions ON service_requests.subscription_id = subscriptions.id").
			Joins("JOIN products ON subscriptions.product_id = products.id").
			Joins("LEFT JOIN franchises ON subscriptions.franchise_id = franchises.id").
			Joins("LEFT JOIN users as service_agent ON service_requests.service_agent_id = service_agent.id").
			Where("service_requests.customer_id = ?", userIDInt).
			Order("service_requests.created_at DESC").
			Find(&results).Error

	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role"})
		return
	}

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetServiceRequestByID returns a specific service request
func GetServiceRequestByID(c *gin.Context) {
	requestID := c.Param("id")
	requestIDInt, err := strconv.ParseUint(requestID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	userID := c.GetString("user_id")
	userIDInt, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	role := c.GetString("role")

	// Check if the user has permission to view this service request
	var count int64
	switch role {
	case database.RoleAdmin:
		// Admin can see any service request
		// No additional check needed
		database.DB.Model(&database.ServiceRequest{}).Where("id = ?", requestIDInt).Count(&count)
	case database.RoleFranchiseOwner:
		// Check if the service request belongs to this franchise owner
		database.DB.Model(&database.ServiceRequest{}).
			Joins("JOIN subscriptions ON service_requests.subscription_id = subscriptions.id").
			Joins("JOIN franchises ON subscriptions.franchise_id = franchises.id").
			Where("service_requests.id = ? AND franchises.owner_id = ?", requestIDInt, userIDInt).
			Count(&count)
	case database.RoleServiceAgent:
		// Check if the service request is assigned to this service agent
		database.DB.Model(&database.ServiceRequest{}).
			Where("id = ? AND service_agent_id = ?", requestIDInt, userIDInt).
			Count(&count)
	case database.RoleCustomer:
		// Check if the service request belongs to this customer
		database.DB.Model(&database.ServiceRequest{}).
			Where("id = ? AND customer_id = ?", requestIDInt, userIDInt).
			Count(&count)
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role"})
		return
	}

	if count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this service request"})
		return
	}

	// Fetch the service request details
	type ServiceRequestWithDetails struct {
		ID               uint       `json:"id"`
		Type             string     `json:"type"`
		Status           string     `json:"status"`
		Description      string     `json:"description"`
		ScheduledTime    *time.Time `json:"scheduled_time"`
		CompletionTime   *time.Time `json:"completion_time"`
		Notes            string     `json:"notes"`
		Rating           *int       `json:"rating"`
		Feedback         string     `json:"feedback"`
		CreatedAt        time.Time  `json:"created_at"`
		UpdatedAt        time.Time  `json:"updated_at"`
		CustomerID       uint       `json:"customer_id"`
		CustomerName     string     `json:"customer_name"`
		CustomerEmail    string     `json:"customer_email"`
		CustomerPhone    string     `json:"customer_phone"`
		CustomerAddress  string     `json:"customer_address"`
		ProductID        uint       `json:"product_id"`
		ProductName      string     `json:"product_name"`
		SubscriptionID   uint       `json:"subscription_id"`
		FranchiseID      *uint      `json:"franchise_id"`
		FranchiseName    string     `json:"franchise_name"`
		ServiceAgentID   *uint      `json:"service_agent_id"`
		ServiceAgentName string     `json:"service_agent_name"`
	}

	var result ServiceRequestWithDetails

	err = database.DB.Table("service_requests").
		Select(`
                        service_requests.id,
                        service_requests.type,
                        service_requests.status,
                        service_requests.description,
                        service_requests.scheduled_time,
                        service_requests.completion_time,
                        service_requests.notes,
                        service_requests.rating,
                        service_requests.feedback,
                        service_requests.created_at,
                        service_requests.updated_at,
                        service_requests.customer_id,
                        customer.name as customer_name,
                        customer.email as customer_email,
                        customer.phone as customer_phone,
                        customer.address as customer_address,
                        subscriptions.product_id,
                        products.name as product_name,
                        service_requests.subscription_id,
                        franchises.id as franchise_id,
                        franchises.name as franchise_name,
                        service_requests.service_agent_id,
                        service_agent.name as service_agent_name
                `).
		Joins("JOIN users as customer ON service_requests.customer_id = customer.id").
		Joins("JOIN subscriptions ON service_requests.subscription_id = subscriptions.id").
		Joins("JOIN products ON subscriptions.product_id = products.id").
		Joins("LEFT JOIN franchises ON subscriptions.franchise_id = franchises.id").
		Joins("LEFT JOIN users as service_agent ON service_requests.service_agent_id = service_agent.id").
		Where("service_requests.id = ?", requestIDInt).
		First(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service request not found"})
		} else {
			log.Printf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateServiceRequest creates a new service request
func CreateServiceRequest(c *gin.Context) {
	var request ServiceRequestCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	userIDInt, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if subscription exists and belongs to the user
	var subscription database.Subscription
	err = database.DB.
		Preload("Franchise").
		Where("id = ? AND customer_id = ?", request.SubscriptionID, userIDInt).
		First(&subscription).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found or doesn't belong to you"})
		} else {
			log.Printf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		}
		return
	}

	// Check if subscription is active
	if subscription.Status != database.SubscriptionStatusActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot create service request for inactive subscription"})
		return
	}

	// Begin transaction
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("Transaction error: %v", tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Create service request
	serviceRequest := database.ServiceRequest{
		CustomerID:     uint(userIDInt),
		SubscriptionID: uint(request.SubscriptionID),
		Type:           request.RequestType,
		Status:         database.ServiceStatusPending,
		Description:    request.Description,
	}

	if err := tx.Create(&serviceRequest).Error; err != nil {
		tx.Rollback()
		log.Printf("Error creating service request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service request"})
		return
	}

	// Create notification for customer
	customerNotification := database.Notification{
		UserID:      uint(userIDInt),
		Title:       "Service Request Created",
		Message:     "Your service request has been created and is pending assignment.",
		Type:        "service_request",
		RelatedID:   &serviceRequest.ID,
		RelatedType: "service_request",
		IsRead:      false,
	}

	if err := tx.Create(&customerNotification).Error; err != nil {
		tx.Rollback()
		log.Printf("Error creating customer notification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	// If franchise exists, create notification for franchise owner
	if subscription.FranchiseID != 0 && subscription.Franchise.OwnerID != 0 {
		franchiseOwnerNotification := database.Notification{
			UserID:      subscription.Franchise.OwnerID,
			Title:       "New Service Request",
			Message:     "A new service request has been created and needs your attention.",
			Type:        "service_request",
			RelatedID:   &serviceRequest.ID,
			RelatedType: "service_request",
			IsRead:      false,
		}

		if err := tx.Create(&franchiseOwnerNotification).Error; err != nil {
			tx.Rollback()
			log.Printf("Error creating franchise owner notification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      serviceRequest.ID,
		"message": "Service request created successfully",
	})
}

// UpdateServiceRequest updates a service request
func UpdateServiceRequest(c *gin.Context) {
	requestID := c.Param("id")
	requestIDInt, err := strconv.ParseUint(requestID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	var updateRequest ServiceRequestUpdateRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	userIDInt, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	role := c.GetString("role")

	// Check if the user has permission to update this service request
	var count int64
	var serviceRequest database.ServiceRequest

	switch role {
	case database.RoleAdmin:
		// Admin can update any service request
		database.DB.Model(&database.ServiceRequest{}).Where("id = ?", requestIDInt).Count(&count)
	case database.RoleFranchiseOwner:
		// Check if the service request belongs to this franchise owner
		database.DB.Model(&database.ServiceRequest{}).
			Joins("JOIN subscriptions ON service_requests.subscription_id = subscriptions.id").
			Joins("JOIN franchises ON subscriptions.franchise_id = franchises.id").
			Where("service_requests.id = ? AND franchises.owner_id = ?", requestIDInt, userIDInt).
			Count(&count)
	case database.RoleServiceAgent:
		// Check if the service request is assigned to this service agent
		database.DB.Model(&database.ServiceRequest{}).
			Where("id = ? AND service_agent_id = ?", requestIDInt, userIDInt).
			Count(&count)
	case database.RoleCustomer:
		// Customers can only update their own service requests with limited fields
		// Only allow cancellation before it's assigned
		if updateRequest.Status != database.ServiceStatusCancelled {
			c.JSON(http.StatusForbidden, gin.H{"error": "Customers can only cancel service requests"})
			return
		}

		err = database.DB.Where("id = ? AND customer_id = ? AND status = ?",
			requestIDInt, userIDInt, database.ServiceStatusPending).First(&serviceRequest).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusForbidden, gin.H{"error": "You can only cancel pending service requests"})
			} else {
				log.Printf("Database error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			}
			return
		}
		count = 1
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role"})
		return
	}

	if count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this service request"})
		return
	}

	// Begin transaction
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("Transaction error: %v", tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Update service request
	updates := map[string]interface{}{}

	if updateRequest.Status != "" && (role == database.RoleAdmin ||
		role == database.RoleFranchiseOwner ||
		role == database.RoleServiceAgent ||
		(role == database.RoleCustomer && updateRequest.Status == database.ServiceStatusCancelled)) {
		updates["status"] = updateRequest.Status
	}

	if updateRequest.ScheduledDate != "" && (role == database.RoleAdmin ||
		role == database.RoleFranchiseOwner ||
		role == database.RoleServiceAgent) {
		scheduledDate, err := time.Parse(time.RFC3339, updateRequest.ScheduledDate)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled date format"})
			return
		}
		updates["scheduled_time"] = scheduledDate
	}

	if updateRequest.CompletionDate != "" && (role == database.RoleAdmin ||
		role == database.RoleFranchiseOwner ||
		role == database.RoleServiceAgent) {
		completionDate, err := time.Parse(time.RFC3339, updateRequest.CompletionDate)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid completion date format"})
			return
		}
		updates["completion_time"] = completionDate
	}

	if updateRequest.Notes != "" && (role == database.RoleAdmin ||
		role == database.RoleFranchiseOwner ||
		role == database.RoleServiceAgent) {
		updates["notes"] = updateRequest.Notes
	}

	// Check if agent ID is provided and valid
	if updateRequest.AgentID != 0 && (role == database.RoleAdmin || role == database.RoleFranchiseOwner) {
		// Verify agent exists and is a service agent
		var agentCount int64
		if role == database.RoleFranchiseOwner {
			// Franchise owners can only assign agents from their franchise
			err = database.DB.Model(&database.User{}).
				Joins("JOIN franchises ON franchises.id = users.franchise_id").
				Where("users.id = ? AND users.role = ? AND franchises.owner_id = ?",
					updateRequest.AgentID, database.RoleServiceAgent, userIDInt).
				Count(&agentCount).Error
		} else {
			// Admins can assign any service agent
			err = database.DB.Model(&database.User{}).
				Where("id = ? AND role = ?", updateRequest.AgentID, database.RoleServiceAgent).
				Count(&agentCount).Error
		}

		if err != nil {
			tx.Rollback()
			log.Printf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}

		if agentCount == 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service agent ID"})
			return
		}

		updates["service_agent_id"] = updateRequest.AgentID

		// If status is not already assigned or later, set it to assigned
		var currentStatus string
		err = database.DB.Model(&database.ServiceRequest{}).
			Select("status").
			Where("id = ?", requestIDInt).
			Pluck("status", &currentStatus).Error

		if err != nil {
			tx.Rollback()
			log.Printf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}

		if currentStatus == database.ServiceStatusPending {
			updates["status"] = database.ServiceStatusAssigned
		}
	}

	if len(updates) == 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid updates provided"})
		return
	}

	// Perform the update
	result := tx.Model(&database.ServiceRequest{}).Where("id = ?", requestIDInt).Updates(updates)
	if result.Error != nil {
		tx.Rollback()
		log.Printf("Error updating service request: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service request"})
		return
	}

	// Get the updated service request for notifications
	var updatedRequest database.ServiceRequest
	if err := tx.Preload("Customer").First(&updatedRequest, requestIDInt).Error; err != nil {
		tx.Rollback()
		log.Printf("Error retrieving updated service request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Create notifications based on changes
	if updateRequest.Status != "" {
		statusNotification := database.Notification{
			UserID:      updatedRequest.CustomerID,
			Title:       "Service Request Updated",
			Message:     fmt.Sprintf("Your service request status has been updated to %s.", updateRequest.Status),
			Type:        "service_request",
			RelatedID:   &updatedRequest.ID,
			RelatedType: "service_request",
			IsRead:      false,
		}

		if err := tx.Create(&statusNotification).Error; err != nil {
			tx.Rollback()
			log.Printf("Error creating status notification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
			return
		}
	}

	if updateRequest.AgentID != 0 {
		// Notify customer about agent assignment
		agentNotification := database.Notification{
			UserID:      updatedRequest.CustomerID,
			Title:       "Service Agent Assigned",
			Message:     "A service agent has been assigned to your service request.",
			Type:        "service_request",
			RelatedID:   &updatedRequest.ID,
			RelatedType: "service_request",
			IsRead:      false,
		}

		if err := tx.Create(&agentNotification).Error; err != nil {
			tx.Rollback()
			log.Printf("Error creating agent notification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
			return
		}

		// Notify agent about assignment
		assignmentNotification := database.Notification{
			UserID:      updateRequest.AgentID,
			Title:       "New Service Assignment",
			Message:     fmt.Sprintf("You have been assigned to service request #%d.", updatedRequest.ID),
			Type:        "service_request",
			RelatedID:   &updatedRequest.ID,
			RelatedType: "service_request",
			IsRead:      false,
		}

		if err := tx.Create(&assignmentNotification).Error; err != nil {
			tx.Rollback()
			log.Printf("Error creating assignment notification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
			return
		}
	}

	if updateRequest.ScheduledDate != "" {
		// Notify customer about scheduled date
		scheduleNotification := database.Notification{
			UserID:      updatedRequest.CustomerID,
			Title:       "Service Visit Scheduled",
			Message:     fmt.Sprintf("Your service request has been scheduled for %s.", updateRequest.ScheduledDate),
			Type:        "service_request",
			RelatedID:   &updatedRequest.ID,
			RelatedType: "service_request",
			IsRead:      false,
		}

		if err := tx.Create(&scheduleNotification).Error; err != nil {
			tx.Rollback()
			log.Printf("Error creating schedule notification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Service request updated successfully",
	})
}

// CancelServiceRequest cancels a service request (customer endpoint)
func CancelServiceRequest(c *gin.Context) {
	requestID := c.Param("id")
	requestIDInt, err := strconv.ParseUint(requestID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	userID := c.GetString("user_id")
	userIDInt, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if service request exists and belongs to the user
	var serviceRequest database.ServiceRequest
	err = database.DB.Where("id = ? AND customer_id = ?", requestIDInt, userIDInt).First(&serviceRequest).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service request not found or doesn't belong to you"})
		} else {
			log.Printf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		}
		return
	}

	// Check if the service request can be cancelled
	if serviceRequest.Status != database.ServiceStatusPending &&
		serviceRequest.Status != database.ServiceStatusAssigned &&
		serviceRequest.Status != database.ServiceStatusScheduled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service request cannot be cancelled in its current state"})
		return
	}

	// Begin transaction
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("Transaction error: %v", tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Update service request status
	if err := tx.Model(&serviceRequest).Update("status", database.ServiceStatusCancelled).Error; err != nil {
		tx.Rollback()
		log.Printf("Error updating service request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel service request"})
		return
	}

	// Create notification for customer
	customerNotification := database.Notification{
		UserID:      uint(userIDInt),
		Title:       "Service Request Cancelled",
		Message:     "Your service request has been cancelled.",
		Type:        "service_request",
		RelatedID:   &serviceRequest.ID,
		RelatedType: "service_request",
		IsRead:      false,
	}

	if err := tx.Create(&customerNotification).Error; err != nil {
		tx.Rollback()
		log.Printf("Error creating customer notification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	// If assigned to a service agent, notify them
	if serviceRequest.ServiceAgentID != nil {
		agentNotification := database.Notification{
			UserID:      *serviceRequest.ServiceAgentID,
			Title:       "Service Request Cancelled",
			Message:     "A service request assigned to you has been cancelled by the customer.",
			Type:        "service_request",
			RelatedID:   &serviceRequest.ID,
			RelatedType: "service_request",
			IsRead:      false,
		}

		if err := tx.Create(&agentNotification).Error; err != nil {
			tx.Rollback()
			log.Printf("Error creating agent notification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel service request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Service request cancelled successfully",
	})
}

// SubmitServiceFeedback submits customer feedback for a completed service
func SubmitServiceFeedback(c *gin.Context) {
	requestID := c.Param("id")
	requestIDInt, err := strconv.ParseUint(requestID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	var feedbackRequest FeedbackRequest
	if err := c.ShouldBindJSON(&feedbackRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	userIDInt, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if service request exists, belongs to the user, and is completed
	var serviceRequest database.ServiceRequest
	err = database.DB.Where("id = ? AND customer_id = ? AND status = ?",
		requestIDInt, userIDInt, database.ServiceStatusCompleted).First(&serviceRequest).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service request not found, doesn't belong to you, or is not completed"})
		} else {
			log.Printf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		}
		return
	}

	// Begin transaction
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("Transaction error: %v", tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Update service request with feedback
	rating := feedbackRequest.Rating
	updates := map[string]interface{}{
		"rating":   rating,
		"feedback": feedbackRequest.Feedback,
	}

	if err := tx.Model(&serviceRequest).Updates(updates).Error; err != nil {
		tx.Rollback()
		log.Printf("Error updating service request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit feedback"})
		return
	}

	// If service request had a service agent, create notification
	if serviceRequest.ServiceAgentID != nil {
		agentNotification := database.Notification{
			UserID:      *serviceRequest.ServiceAgentID,
			Title:       "Service Feedback Received",
			Message:     fmt.Sprintf("You received a %d-star rating for your service.", rating),
			Type:        "service_feedback",
			RelatedID:   &serviceRequest.ID,
			RelatedType: "service_request",
			IsRead:      false,
		}

		if err := tx.Create(&agentNotification).Error; err != nil {
			tx.Rollback()
			log.Printf("Error creating agent notification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit feedback"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Feedback submitted successfully",
	})
}
