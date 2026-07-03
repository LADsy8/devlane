package middleware

import (
	"net/http"
	"time"

	"github.com/Devlaner/devlane/api/internal/redis"
	"github.com/gin-gonic/gin"
)

// RateLimit caps requests per client IP to limit uses per window, backed by
// Redis. Like other optional-infra integrations in this codebase, it fails
// open (allows the request) when rdb is nil or a Redis error occurs, rather
// than turning a Redis outage into an outage of the whole API.
func RateLimit(rdb *redis.Client, prefix string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil {
			c.Next()
			return
		}
		key := redis.PrefixRateLimit + prefix + ":" + c.ClientIP()
		ok, err := rdb.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			c.Next()
			return
		}
		if !ok {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests, please try again later"})
			return
		}
		c.Next()
	}
}
