package ratelimit

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client    *redis.Client
	requests  int
	window    time.Duration
	keyPrefix string
}

func NewRateLimiter(client *redis.Client, requests int, window time.Duration, keyPrefix string) *RateLimiter {
	return &RateLimiter{
		client:    client,
		requests:  requests,
		window:    window,
		keyPrefix: keyPrefix,
	}
}

func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if rl.client == nil {
			c.Next()
			return
		}

		key := rl.keyPrefix + ":" + c.ClientIP()
		ctx := c.Request.Context()

		count, err := rl.client.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			rl.client.Expire(ctx, key, rl.window)
		}

		if count > int64(rl.requests) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
