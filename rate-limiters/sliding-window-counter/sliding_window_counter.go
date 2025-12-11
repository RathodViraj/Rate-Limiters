package slidingwindowcounter

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

type SlidingWindowCounter struct {
	rdb       *redis.Client
	luaScript *redis.Script
	limit     int
	window    time.Duration
}

func NewSlidingWindowCounter(r *redis.Client, limit int, window time.Duration) *SlidingWindowCounter {
	script := redis.NewScript(slidingWindowCounterLua)
	if script == nil {
		log.Fatal("couldn't load the lua script")
	}
	return &SlidingWindowCounter{
		rdb:       r,
		luaScript: script,
		limit:     limit,
		window:    window,
	}
}

func (swc *SlidingWindowCounter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("swc:ip:%s", ip)

		ctx := context.Background()
		res, err := swc.luaScript.Run(
			ctx,
			swc.rdb,
			[]string{key},
			time.Now().Unix(),
			5,
			3,
		).Result()
		if err != nil {
			log.Printf("redis error: %s", err)
			c.Next()
		}

		arr := res.([]interface{})
		if arr[0].(int64) == 0 {
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				gin.H{
					"error":       "too many requets",
					"retry_after": arr[2].(int64),
				},
			)
		}

		c.Next()
	}
}

//go:embed sliding_window_counter.lua
var slidingWindowCounterLua string

// reference embed to satisfy linters that don't detect go:embed usage
var _ embed.FS
