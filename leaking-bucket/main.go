package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

type LeakingBucket struct {
	capacity   int
	current    int
	leakRate   int
	leakPeriod time.Duration
	mutex      sync.Mutex
}

func NewLeakingBucket(capacity int, leakRate int, leakPeriod time.Duration) *LeakingBucket {
	lb := &LeakingBucket{
		capacity:   capacity,
		current:    0,
		leakRate:   leakRate,
		leakPeriod: leakPeriod,
	}

	go lb.startLeaking()

	return lb
}

func (lb *LeakingBucket) startLeaking() {
	ticker := time.NewTicker(lb.leakPeriod)
	defer ticker.Stop()
	for range ticker.C {
		lb.mutex.Lock()
		if lb.current > lb.leakRate {
			lb.current -= lb.leakRate
		} else {
			lb.current = 0
		}

		lb.mutex.Unlock()
	}
}

func (lb *LeakingBucket) AddRequest() bool {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	if lb.current < lb.capacity {
		lb.current++
		return true
	}

	return false
}

func RateLimiter(bucket *LeakingBucket) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !bucket.AddRequest() {
			ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func main() {
	bucket := NewLeakingBucket(5, 1, 10*time.Second)

	server := gin.Default()

	server.Use(RateLimiter(bucket))

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"capacity": bucket.capacity, "current": bucket.current})
	})

	if err := server.Run(":8080"); err != nil {
		panic(err)
	}
}
