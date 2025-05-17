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
			auth.POST("/login", controllers.Login)
			auth.POST("/register", controllers.Register)
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
		protected.POST("/auth/refresh", controllers.RefreshToken)
		protected.POST("/auth/refresh/v2", controllers.RefreshTokenNew)

		protected.GET("/profile", controllers.GetUserProfile)
		protected.PUT("/profile", controllers.UpdateUserProfile)
		protected.POST("/profile/change-password", controllers.ChangePassword)
		protected.GET("/profile/v2", controllers.GetUserProfileNew)
		protected.PUT("/profile/v2", controllers.UpdateUserProfileNew)
		protected.POST("/profile/change-password/v2", controllers.ChangePasswordNew)

		// Admin routes
		admin := protected.Group("/admin")
		admin.Use(middleware.AdminAuthMiddleware())
		{
			admin.GET("/users/:id", controllers.GetUserByID)
			admin.GET("/users/role/:role", controllers.GetUsersByRole)
			admin.GET("/orders", controllers.AdminGetOrders)
			admin.GET("/users/:id/v2", controllers.GetUserByIDNew)
			admin.GET("/users/role/:role/v2", controllers.GetUsersByRoleNew)
			admin.GET("/dashboard", controllers.AdminDashboard)

			// ✅ Products Management
			admin.POST("/products", controllers.CreateProduct)
			admin.GET("/products", controllers.GetProducts)
			admin.GET("/products/:id", controllers.GetProductByID)
			admin.PUT("/products/:id", controllers.UpdateProduct)
			admin.DELETE("/products/:id", controllers.DeleteProduct)
			admin.PATCH("/products/:id/toggle-status", controllers.ToggleProductStatus)
			admin.PATCH("/franchises/:id", controllers.AdminUpdateFranchise)
			admin.POST("/franchises", controllers.CreateFranchise)
			admin.PATCH("/orders/:id/assign", controllers.AssignOrderToFranchise)
			admin.GET("/customers/:id/subscriptions", controllers.GetCustomerSubscriptionsByAdmin)

			//  this route for fetching all franchises
			admin.GET("/franchises", controllers.GetAllFranchises)
		}

		// 🧑‍🔧 Service Agent Routes
		agent := protected.Group("/agent")
		agent.Use(middleware.ServiceAgentAuthMiddleware())
		{
			agent.GET("/tasks", controllers.GetAgentTasks)
			agent.GET("/dashboard", controllers.GetServiceAgentDashboard)
		}

		// Orders
		orders := protected.Group("/orders")
		{
			orders.POST("", middleware.CustomerAuthMiddleware(), controllers.CreateOrder)
			orders.GET("/customer", middleware.CustomerAuthMiddleware(), controllers.GetCustomerOrders)
			orders.PUT("/:id/status", middleware.AdminOrFranchiseAuthMiddleware(), controllers.UpdateOrderStatus)
			orders.GET("/:id", controllers.GetOrderByID)
		}

		// Subscriptions
		subscriptions := protected.Group("/subscriptions")
		{
			subscriptions.POST("", middleware.CustomerAuthMiddleware(), controllers.CreateSubscription)
			subscriptions.GET("/customer", middleware.CustomerAuthMiddleware(), controllers.GetMySubscriptions)
			subscriptions.PUT("/:id", middleware.CustomerAuthMiddleware(), controllers.UpdateSubscription)
			subscriptions.POST("/:id/cancel", middleware.CustomerAuthMiddleware(), controllers.CancelSubscription)

			subscriptions.GET("/franchise", middleware.FranchiseOwnerAuthMiddleware(), controllers.GetFranchiseSubscriptions)

		}

		// Service requests
		services := protected.Group("/services")
		{
			services.POST("", middleware.CustomerAuthMiddleware(), controllers.CreateServiceRequest)
			services.POST("/:id/feedback", middleware.CustomerAuthMiddleware(), controllers.SubmitServiceFeedback)
			services.POST("/:id/cancel", middleware.CustomerAuthMiddleware(), controllers.CancelServiceRequest)
			services.GET("", controllers.GetServiceRequestsNew)
			services.GET("/:id", controllers.GetServiceRequestByIDNew)
			services.PUT("/:id", controllers.UpdateServiceRequestNew)
		}
		// Service agents

		// Franchises
		franchises := protected.Group("/franchises")
		{
			franchises.POST("", middleware.FranchiseOwnerAuthMiddleware(), controllers.CreateFranchise)
			franchises.POST("/:id/approve", middleware.AdminAuthMiddleware(), controllers.ApproveFranchise)
			franchises.POST("/:id/reject", middleware.AdminAuthMiddleware(), controllers.RejectFranchise)
			franchises.PUT("/:id", middleware.AdminOrFranchiseAuthMiddleware(), controllers.UpdateFranchise)
			franchises.GET("/:id/service-agents", middleware.AdminOrFranchiseAuthMiddleware(), controllers.GetFranchiseServiceAgents)
			franchises.GET("/search", controllers.SearchFranchises)

			//this route for dashboard
			franchises.GET("/dashboard", controllers.GetFranchiseDashboard)

		}

		// Payments
		payments := protected.Group("/payments")
		{
			payments.POST("/generate-order", middleware.CustomerAuthMiddleware(), controllers.GeneratePaymentOrder)
			payments.POST("/generate-monthly", middleware.CustomerAuthMiddleware(), controllers.GenerateMonthlyPayment)
			payments.POST("/verify", middleware.CustomerAuthMiddleware(), controllers.VerifyPayment)
			payments.GET("", controllers.GetPaymentHistory)
			payments.GET("/:id", controllers.GetPaymentByID)
		}

		// Add this route for franchise dashboard
		protected.GET("/franchise/dashboard", controllers.GetFranchiseDashboard)
	}
}
