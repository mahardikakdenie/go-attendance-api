package middleware

import (
	"net/http"

	"go-attendance-api/internal/service"

	"github.com/gin-gonic/gin"
)

func CookieAuth(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			c.Abort()
			return
		}

		user, err := authService.GetMe(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid session"})
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("tenant_id", user.TenantID)
		c.Set("role", user.Role)

		c.Next()
	}
}
