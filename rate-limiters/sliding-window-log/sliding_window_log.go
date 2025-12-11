package slidingwindowlog

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb       *redis.Client
	luaScript *redis.Script
	limit     int
	window    time.Duration
}

func NewRateLimiter(r *redis.Client, limit int, window time.Duration) (*RateLimiter, error) {
	script := redis.NewScript(slidingWindowLogLua)
	rl := &RateLimiter{
		rdb:       r,
		luaScript: script,
		limit:     limit,
		window:    window,
	}

	return rl, nil
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for /free endpoint
		if c.Request.URL.Path == "/free" {
			c.Next()
			return
		}

		ip := c.ClientIP()
		key := fmt.Sprintf("swl:ip:%s", ip)
		now := time.Now().UnixMilli()

		ctx := context.Background()
		allowed, err := rl.luaScript.Run(ctx, rl.rdb, []string{key}, now, rl.window.Milliseconds(), rl.limit).Int()
		if err != nil {
			log.Printf("redis error: %v", err)
			c.Next()
		}

		if allowed == 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, try again later",
			})
			return
		}

		c.Next()
	}
}

//go:embed sliding_window_log.lua
var slidingWindowLogLua string

// reference embed to satisfy linters that don't detect go:embed usage
var _ embed.FS
