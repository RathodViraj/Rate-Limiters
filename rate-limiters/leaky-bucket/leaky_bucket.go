package leaky_bucket

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type LeakyBucket struct {
	queue       chan struct{}
	maxQueue    int
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
		queue:       make(chan struct{}, maxQueue),
		maxQueue:    maxQueue,
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
			select {
			case <-lb.queue:
				// Request processed, do nothing
			case <-lb.stopChannel:
				return
			default:
				// Queue is empty
			}
		}
	}
}

func (lb *LeakyBucket) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		select {
		case lb.queue <- struct{}{}:
			c.Next()
		default:
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		}
	}
}

func (lb *LeakyBucket) Stop() {
	close(lb.stopChannel)
	lb.wg.Wait()
}
