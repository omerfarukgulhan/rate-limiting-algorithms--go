package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

type SlidingWindow struct {
	timestamps []time.Time
	limit      int
	window     time.Duration
	mutex      sync.Mutex
}

func NewSlidingWindow(limit int, window time.Duration) *SlidingWindow {
	return &SlidingWindow{
		timestamps: []time.Time{},
		limit:      limit,
		window:     window,
	}
}

func (sw *SlidingWindow) Allow() bool {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	now := time.Now()
	cutoff := now.Add(-sw.window)
	newTimestamps := []time.Time{}
	for _, ts := range sw.timestamps {
		if ts.After(cutoff) {
			newTimestamps = append(newTimestamps, ts)
		}
	}

	sw.timestamps = newTimestamps
	if len(sw.timestamps) < sw.limit {
		sw.timestamps = append(sw.timestamps, now)
		return true
	}

	return false
}

func RateLimiter(window *SlidingWindow) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !window.Allow() {
			ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func main() {
	window := NewSlidingWindow(5, 10*time.Second)

	server := gin.Default()

	server.Use(RateLimiter(window))

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
	})

	if err := server.Run(":8080"); err != nil {
		panic(err)
	}
}
