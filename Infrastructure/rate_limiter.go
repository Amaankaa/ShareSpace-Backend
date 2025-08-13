package infrastructure

import (
	"net/http"
	// "time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// rateLimitMiddleware is a simple IP-based rate limiter for Gin.
func RateLimitMiddleware(limit rate.Limit, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(limit, burst)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please try again later."})
			return
		}
		c.Next()
	}
}

func NewRateLimiter(r float64, b int) gin.HandlerFunc {
	return RateLimitMiddleware(rate.Limit(r), b)
}

const (
	RateLimit  = 1
	BurstLimit = 5
)
