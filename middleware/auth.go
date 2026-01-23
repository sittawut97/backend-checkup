package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sittawut/backend-appointment/config"
)

type Claims struct {
	UserID     string `json:"user_id"`
	Phone      string `json:"phone"`
	Role       string `json:"role"`
	Email      string `json:"email,omitempty"`
	FullName   string `json:"full_name,omitempty"`
	EmployeeID string `json:"employee_id,omitempty"`
	Department string `json:"department,omitempty"`
	JobTitle   string `json:"job_title,omitempty"`
	Provider   string `json:"provider,omitempty"`
	jwt.RegisteredClaims
}

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("[AuthMiddleware] Request: %s %s\n", c.Request.Method, c.Request.URL.Path)
		authHeader := c.GetHeader("Authorization")
		tokenString := ""
		if authHeader != "" {
			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "Invalid authorization header format",
				})
				c.Abort()
				return
			}
			tokenString = parts[1]
		} else {
			// Fallback to HttpOnly cookie (used by Azure login flow)
			cookieToken, err := c.Cookie("token")
			if err != nil || cookieToken == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "Authorization required",
				})
				c.Abort()
				return
			}
			tokenString = cookieToken
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid token claims",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("phone", claims.Phone)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Set("full_name", claims.FullName)
		c.Set("employee_id", claims.EmployeeID)
		c.Set("department", claims.Department)
		c.Set("job_title", claims.JobTitle)
		c.Set("provider", claims.Provider)

		c.Next()
	}
}

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "User role not found",
			})
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Invalid user role type",
			})
			c.Abort()
			return
		}

		allowed := false
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
