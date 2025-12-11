package fixedwindowcounter

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type FixedWindowCounter struct {
	limit      int
	windowSize int64
	counters   map[int64]int
	startTime  int64
	mu         sync.Mutex
}

func NewFixedWindowCounter(limit int, windowSize int64) *FixedWindowCounter {
	return &FixedWindowCounter{
		limit:      limit,
		windowSize: windowSize,
		startTime:  time.Now().Unix(),
		counters:   make(map[int64]int),
	}
}

func (fwc *FixedWindowCounter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		idx := (time.Now().Unix() - fwc.startTime) / fwc.windowSize
		valid := true
		fwc.mu.Lock()
		if l, ok := fwc.counters[idx]; ok {
			if l < fwc.limit {
				fwc.counters[idx] += 1
			} else {
				valid = false
			}
		} else {
			fwc.counters[idx] = 1
		}
		fwc.mu.Unlock()
		// for k, v := range fwc.counters {
		// 	fmt.Println(k, " -> ", v)
		// }

		if !valid {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		}

		c.Next()
	}
}
