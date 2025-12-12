package main

import (
	"rate-limiter-strategies/db"
	leaky_bucket "rate-limiter-strategies/rate-limiters/leaky-bucket"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	rdb := db.Connect()
	if rdb == nil {
		panic("Failed to connect to Redis")
	}

	r := gin.Default()

	lb := leaky_bucket.NewLeakyBucket(8, 100*time.Millisecond)
	defer lb.Stop()
	r.Use(lb.Middleware())

	// tb := tokenbucket.NewTokenBucketRL(time.Duration(5)*time.Second, 3)
	// go tb.StartFiller()
	// r.Use(tb.Middleware())

	// fwc := fixedwindowcounter.NewFixedWindowCounter(3, 5)
	// r.Use(fwc.Middleware())

	// swl, err := slidingwindowlog.NewRateLimiter(rdb, 6, time.Duration(10)*time.Second)
	// if err != nil {
	// 	panic("faild to create sliding window log rate limiter")
	// }
	// r.Use(swl.Middleware())

	// swc := slidingwindowcounter.NewSlidingWindowCounter(rdb, 3, 5)
	// r.Use(swc.Middleware())

	r.GET("/home", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the home page!",
		})
	})

	r.GET("/free", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the home page!",
		})
	})

	r.Run(":8080")
}
