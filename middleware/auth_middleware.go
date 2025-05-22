package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"aquahome/utils"
)

// AuthMiddleware validates JWT tokens and extracts user information
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// ✅ Directly use UserID, already uint
		c.Set("userID", claims.UserID)  // camelCase (used in new code)
		c.Set("user_id", claims.UserID) // snake_case (used in old code)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RoleAuthMiddleware validates user roles
func RoleAuthMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, r := range roles {
			if r == userRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		c.Abort()
	}
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return RoleAuthMiddleware("admin")
}

func FranchiseOwnerAuthMiddleware() gin.HandlerFunc {
	return RoleAuthMiddleware("admin", "franchise_owner")
}

func CustomerAuthMiddleware() gin.HandlerFunc {
	return RoleAuthMiddleware("customer")
}

func ServiceAgentAuthMiddleware() gin.HandlerFunc {
	return RoleAuthMiddleware("admin", "service_agent")
}

func AdminOrFranchiseAuthMiddleware() gin.HandlerFunc {
	return RoleAuthMiddleware("admin", "franchise_owner")
}

func AdminOrServiceAgentAuthMiddleware() gin.HandlerFunc {
	return RoleAuthMiddleware("admin", "service_agent")
}
