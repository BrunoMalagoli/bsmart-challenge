package middleware

import (
	"net/http"
	"strings"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/auth"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/gin-gonic/gin"
)

// Context keys for storing user information
const (
	UserIDKey    = "user_id"
	UserEmailKey = "user_email"
	UserRoleKey  = "user_role"
)

// RequireAuth is middleware that validates JWT token
func RequireAuth(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			models.RespondError(c, http.StatusUnauthorized, "MISSING_AUTH", "Authorization header is required")
			c.Abort()
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			models.RespondError(c, http.StatusUnauthorized, "INVALID_AUTH_FORMAT", "Authorization header must be in format: Bearer <token>")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			models.RespondError(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token")
			c.Abort()
			return
		}

		// Store user information in context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Set(UserRoleKey, claims.Role)

		c.Next()
	}
}

// RequireRole is middleware that checks if user has required role
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context
		userRole, exists := c.Get(UserRoleKey)
		if !exists {
			models.RespondError(c, http.StatusUnauthorized, "NO_USER", "User not authenticated")
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		roleStr, ok := userRole.(string)
		if !ok {
			models.RespondError(c, http.StatusInternalServerError, "INVALID_ROLE_TYPE", "Invalid role type in context")
			c.Abort()
			return
		}

		// Check if user's role matches any of the required roles
		hasRole := false
		for _, role := range roles {
			if roleStr == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			models.RespondError(c, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", "You don't have permission to access this resource")
			c.Abort()
			return
		}

		c.Next()
	}
}

// gets the user ID from context
func GetUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	id, ok := userID.(int)
	return id, ok
}

// gets the user email from context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(UserEmailKey)
	if !exists {
		return "", false
	}
	emailStr, ok := email.(string)
	return emailStr, ok
}

// gets the user role from context
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(UserRoleKey)
	if !exists {
		return "", false
	}
	roleStr, ok := role.(string)
	return roleStr, ok
}
