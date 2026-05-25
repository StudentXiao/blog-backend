package middleware

import (
	"blog-backend/internal/handler"
	"blog-backend/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			handler.Unauthorized(c, "Missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			handler.Unauthorized(c, "Invalid authorization format")
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			handler.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

func AdmiOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists || role != "admin" {
			handler.Forbidden(c, "Admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}
