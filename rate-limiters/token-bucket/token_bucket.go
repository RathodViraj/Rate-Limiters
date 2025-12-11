package tokenbucket

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type TokenBucket struct {
	FillerDuration time.Duration
	MaxTokens      uint32
	Tokens         uint32
	mu             sync.Mutex
}

func NewTokenBucketRL(fd time.Duration, max uint32) *TokenBucket {
	return &TokenBucket{
		FillerDuration: fd,
		MaxTokens:      max,
		Tokens:         0,
	}
}

func (tb *TokenBucket) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for /free endpoint
		if c.Request.URL.Path == "/free" {
			c.Next()
			return
		}

		canHandle := true
		tb.mu.Lock()
		if tb.Tokens > 0 {
			tb.Tokens -= 1
		} else {
			canHandle = false
		}
		tb.mu.Unlock()

		if !canHandle {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		}

		c.Next()
	}
}

func (tb *TokenBucket) StartFiller() {
	for {
		tb.mu.Lock()
		tb.Tokens = tb.MaxTokens
		tb.mu.Unlock()
		time.Sleep(tb.FillerDuration)
	}
}
