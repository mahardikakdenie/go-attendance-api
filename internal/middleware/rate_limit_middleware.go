package middleware

import (
	"fmt"
	"net/http"
	"time"

	"go-attendance-api/internal/config"

	"github.com/gin-gonic/gin"
)

func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Bypass for localhost
		if clientIP == "127.0.0.1" || clientIP == "::1" {
			c.Next()
			return
		}

		key := fmt.Sprintf("rate_limit:%s", clientIP)

		// Limit: 100 requests per minute per IP
		limit := 100
		window := 1 * time.Minute

		count, err := config.NewRedis().Incr(config.Ctx, key).Result()
		if err != nil {
			c.Next() // Fallback if redis is down
			return
		}

		if count == 1 {
			config.NewRedis().Expire(config.Ctx, key, window)
		}

		if int(count) > limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "Too many requests. Please try again later.",
			})
			return
		}

		c.Next()
	}
}
