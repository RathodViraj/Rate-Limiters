package leaky_bucket

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type LeakyBucket struct {
	tokens      int
	maxTokens   int
	leakRate    time.Duration
	mu          sync.Mutex
	stopChannel chan struct{}
	wg          sync.WaitGroup
}

func NewLeakyBucket(maxQueue int, leakRate time.Duration) *LeakyBucket {
	if maxQueue <= 0 {
		maxQueue = 8
	}

	lb := &LeakyBucket{
		tokens:      0,
		maxTokens:   maxQueue,
		leakRate:    leakRate,
		stopChannel: make(chan struct{}),
	}

	lb.wg.Add(1)
	go lb.startLeaking()

	return lb
}

func (lb *LeakyBucket) startLeaking() {
	defer lb.wg.Done()
	ticker := time.NewTicker(lb.leakRate)
	defer ticker.Stop()

	for {
		select {
		case <-lb.stopChannel:
			return
		case <-ticker.C:
			lb.mu.Lock()
			if lb.tokens > 0 {
				lb.tokens--
			}
			lb.mu.Unlock()
		}
	}
}

func (lb *LeakyBucket) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/free" {
			c.Next()
			return
		}

		lb.mu.Lock()
		if lb.tokens < lb.maxTokens {
			lb.tokens++
			lb.mu.Unlock()
			c.Next()
		} else {
			lb.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		}
	}
}

func (lb *LeakyBucket) Stop() {
	close(lb.stopChannel)
	lb.wg.Wait()
}
