package middleware

import (
	"net/http"
	"strconv"

	"order-service/internal/cache"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	cache      *cache.RedisCache
	limit      int
	windowSize int
}

func NewRateLimiter(cache *cache.RedisCache, limit int, windowSize int) *RateLimiter {
	return &RateLimiter{
		cache:      cache,
		limit:      limit,
		windowSize: windowSize,
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.ClientIP()
		key := "rate_limit:" + clientID

		val, err := rl.cache.Get(key)

		if err != nil {
			_ = rl.cache.Set(key, "1", rl.windowSize)
			c.Next()
			return
		}

		count, err := strconv.Atoi(val)
		if err != nil {
			_ = rl.cache.Delete(key)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "rate limiter error",
			})
			return
		}

		if count >= rl.limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return
		}

		_ = rl.cache.Set(key, strconv.Itoa(count+1), rl.windowSize)

		c.Next()
	}
}
