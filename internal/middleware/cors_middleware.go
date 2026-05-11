package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		
		if allowedOrigins == "" || allowedOrigins == "*" {
			// If no allowed origins specified, mirror the requesting origin
			// Note: When Allow-Credentials is true, Allow-Origin cannot be '*'
			if origin != "" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
		} else {
			origins := strings.Split(allowedOrigins, ",")
			for _, o := range origins {
				if strings.TrimSpace(o) == origin {
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Timestamp, X-Request-ID, X-Signature, X-Internal-Secret")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
