package main

import (
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"sync"
	"time"
)

type SlidingWindow struct {
	limit              int
	windowDuration     time.Duration
	currentWindowStart time.Time
	previousWindowReqs int
	currentWindowReqs  int
	mutex              sync.Mutex
}

func NewSlidingWindow(limit int, windowDuration time.Duration) *SlidingWindow {
	return &SlidingWindow{
		limit:              limit,
		windowDuration:     windowDuration,
		currentWindowStart: time.Now(),
		previousWindowReqs: 0,
		currentWindowReqs:  0,
	}
}

func (sw *SlidingWindow) Allow() bool {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	now := time.Now()
	elapsed := now.Sub(sw.currentWindowStart)
	if elapsed >= sw.windowDuration {
		sw.previousWindowReqs = sw.currentWindowReqs
		sw.currentWindowReqs = 0
		sw.currentWindowStart = now
		elapsed = 0
	}

	percentagePassed := float64(elapsed) / float64(sw.windowDuration)
	totalRequests := float64(sw.currentWindowReqs) + float64(sw.previousWindowReqs)*(1-percentagePassed)
	if math.Floor(totalRequests) < float64(sw.limit) {
		sw.currentWindowReqs++
		return true
	}

	return false
}

func RateLimiter(bucket *SlidingWindow) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !bucket.Allow() {
			ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func main() {
	window := NewSlidingWindow(7, 60*time.Second)

	server := gin.Default()

	server.Use(RateLimiter(window))

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
	})

	if err := server.Run(":8080"); err != nil {
		panic(err)
	}
}
