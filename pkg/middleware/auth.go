package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	UserNameKey = "username"
	TenantKey   = "tenant"

	BearerSchema = "Bearer "
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, BearerSchema) {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "unauthorized: missing or invalid authorization header",
			})
			return
		}

		token := strings.TrimPrefix(auth, BearerSchema)

		parts := strings.Split(token, "@")
		if len(parts) != 2 {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "unauthorized: invalid token format, expected 'username@tenant'",
			})
			return
		}

		username := parts[0]
		tenant := parts[1]

		if username == "" || tenant == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "unauthorized: username and tenant cannot be empty",
			})
			return
		}

		c.Set(UserNameKey, username)
		c.Set(TenantKey, tenant)

		c.Next()
	}
}
