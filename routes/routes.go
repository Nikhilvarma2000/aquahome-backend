package routes

import (
	"github.com/gin-gonic/gin"

	"aquahome/controllers"
	"aquahome/middleware"
)

// SetupRoutes configures all application routes
func SetupRoutes(r *gin.Engine) {
	// Public routes (no authentication required)
	public := r.Group("/api")
	{
		// Authentication routes
		auth := public.Group("/auth")
		{
			// Old functions maintained for backward compatibility
			auth.POST("/login", controllers.Login)
			auth.POST("/register", controllers.Register)

			// New GORM-based functions
			auth.POST("/login/v2", controllers.LoginNew)
			auth.POST("/register/v2", controllers.RegisterNew)
		}

		// Products (public view for non-authenticated users)
		public.GET("/products", controllers.GetProducts)
		public.GET("/products/:id", controllers.GetProductByID)
	}

	// Protected routes (authentication required)
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// Refresh token
		protected.POST("/auth/refresh", controllers.RefreshToken)
		protected.POST("/auth/refresh/v2", controllers.RefreshTokenNew)

		// User profile - Legacy SQL versions
		protected.GET("/profile", controllers.GetUserProfile)
		protected.PUT("/profile", controllers.UpdateUserProfile)
		protected.POST("/profile/change-password", controllers.ChangePassword)

		// User profile - GORM versions
		protected.GET("/profile/v2", controllers.GetUserProfileNew)
		protected.PUT("/profile/v2", controllers.UpdateUserProfileNew)
		protected.POST("/profile/change-password/v2", controllers.ChangePasswordNew)

		// User management (Admin only)
		admin := protected.Group("/admin")
		admin.Use(middleware.AdminAuthMiddleware())
		{
			// Legacy SQL versions
			admin.GET("/users/:id", controllers.GetUserByID)
			admin.GET("/users/role/:role", controllers.GetUsersByRole)

			// GORM versions
			admin.GET("/users/:id/v2", controllers.GetUserByIDNew)
			admin.GET("/users/role/:role/v2", controllers.GetUsersByRoleNew)
		}

		// Products management
		products := protected.Group("/products")
		{
			// Admin only endpoints
			products.POST("", middleware.AdminAuthMiddleware(), controllers.CreateProduct)
			products.PUT("/:id", middleware.AdminAuthMiddleware(), controllers.UpdateProduct)
			products.DELETE("/:id", middleware.AdminAuthMiddleware(), controllers.DeleteProduct)
		}

		// Orders
		orders := protected.Group("/orders")
		{
			// Customer endpoints
			orders.POST("", middleware.CustomerAuthMiddleware(), controllers.CreateOrder)
			orders.GET("/customer", middleware.CustomerAuthMiddleware(), controllers.GetCustomerOrders)

			// Admin and Franchise owner endpoints
			orders.PUT("/:id/status", middleware.AdminOrFranchiseAuthMiddleware(), controllers.UpdateOrderStatus)

			// Common endpoints (with role-based permissions within the handler)
			orders.GET("/:id", controllers.GetOrderByID)
		}

		// Subscriptions
		subscriptions := protected.Group("/subscriptions")
		{
			// Customer endpoints
			subscriptions.GET("/customer", middleware.CustomerAuthMiddleware(), controllers.GetCustomerSubscriptions)
			subscriptions.PUT("/:id", middleware.CustomerAuthMiddleware(), controllers.UpdateSubscription)
			subscriptions.POST("/:id/cancel", middleware.CustomerAuthMiddleware(), controllers.CancelSubscription)

			// Common endpoints (with role-based permissions within the handler)
			// subscriptions.GET("/:id", controllers.GetSubscriptionByID)
		}

		// Service requests
		services := protected.Group("/services")
		{
			// Customer endpoints
			services.POST("", middleware.CustomerAuthMiddleware(), controllers.CreateServiceRequestNew)
			services.POST("/:id/feedback", middleware.CustomerAuthMiddleware(), controllers.SubmitServiceFeedbackNew)
			services.POST("/:id/cancel", middleware.CustomerAuthMiddleware(), controllers.CancelServiceRequestNew)

			// Old routes kept for backward compatibility until we fully migrate
			// services.POST("/:id/rate", middleware.CustomerAuthMiddleware(), controllers.RateServiceRequest)
			// services.POST("/:id/assign", middleware.AdminOrFranchiseAuthMiddleware(), controllers.AssignServiceRequest)

			// Common endpoints (with role-based permissions within the handler)
			services.GET("", controllers.GetServiceRequestsNew)
			services.GET("/:id", controllers.GetServiceRequestByIDNew)
			services.PUT("/:id", controllers.UpdateServiceRequestNew)
		}

		// Franchises
		franchises := protected.Group("/franchises")
		{
			// Franchise owner endpoints
			franchises.POST("", middleware.FranchiseOwnerAuthMiddleware(), controllers.CreateFranchise)

			// Admin endpoints
			franchises.POST("/:id/approve", middleware.AdminAuthMiddleware(), controllers.ApproveFranchise)
			franchises.POST("/:id/reject", middleware.AdminAuthMiddleware(), controllers.RejectFranchise)

			// Common endpoints
			franchises.GET("", controllers.GetFranchises)
			franchises.GET("/:id", controllers.GetFranchiseByID)
			franchises.PUT("/:id", middleware.AdminOrFranchiseAuthMiddleware(), controllers.UpdateFranchise)
			franchises.GET("/:id/service-agents", middleware.AdminOrFranchiseAuthMiddleware(), controllers.GetFranchiseServiceAgents)
			franchises.GET("/search", controllers.SearchFranchises)
		}

		// Payments
		payments := protected.Group("/payments")
		{
			// Customer endpoints
			payments.POST("/generate-order", middleware.CustomerAuthMiddleware(), controllers.GeneratePaymentOrder)
			payments.POST("/generate-monthly", middleware.CustomerAuthMiddleware(), controllers.GenerateMonthlyPayment)
			payments.POST("/verify", middleware.CustomerAuthMiddleware(), controllers.VerifyPayment)

			// Common endpoints (with role-based permissions within the handler)
			payments.GET("", controllers.GetPaymentHistory)
			payments.GET("/:id", controllers.GetPaymentByID)
		}
	}
}
