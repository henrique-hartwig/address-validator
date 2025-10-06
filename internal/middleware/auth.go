package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func BearerAuth(validToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is required",
			})
			c.Abort()
			return
		}

		if token != validToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func ValidateHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch {
			contentType := c.GetHeader("Content-Type")

			if contentType != "application/json" && !strings.HasPrefix(contentType, "application/json;") {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error": "Content-Type must be application/json",
				})
				c.Abort()
				return
			}
		}

		accept := c.GetHeader("Accept")
		if accept != "" && accept != "*/*" && !strings.Contains(accept, "application/json") {
			c.JSON(http.StatusNotAcceptable, gin.H{
				"error": "Accept header must include application/json",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
