package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

type FixedWindowCounter struct {
	windowSize time.Duration
	threshold  int
	counters   map[int]int
	lock       sync.Mutex
	startTime  time.Time
}

func NewFixedWindowCounter(windowSize time.Duration, threshold int) *FixedWindowCounter {
	return &FixedWindowCounter{
		windowSize: windowSize,
		threshold:  threshold,
		counters:   make(map[int]int),
		startTime:  time.Now(),
	}
}

func (fwc *FixedWindowCounter) Increment() bool {
	fwc.lock.Lock()
	defer fwc.lock.Unlock()
	currentTime := time.Now()
	elapsedTime := currentTime.Sub(fwc.startTime)
	currentWindow := int(elapsedTime / fwc.windowSize)
	for window := range fwc.counters {
		if window < currentWindow {
			delete(fwc.counters, window)
		}
	}

	if _, exists := fwc.counters[currentWindow]; !exists {
		fwc.counters[currentWindow] = 0
	}

	if fwc.counters[currentWindow] < fwc.threshold {
		fwc.counters[currentWindow]++
		return true
	}

	return false
}

func RateLimiter(fwc *FixedWindowCounter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if fwc.Increment() {
			c.Next()
		} else {
			c.JSON(429, gin.H{"message": "Too Many Requests"})
			c.Abort()
		}
	}
}

func main() {
	counter := NewFixedWindowCounter(1*time.Minute, 5)

	server := gin.Default()

	server.Use(RateLimiter(counter))

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
	})

	if err := server.Run(":8080"); err != nil {
		panic(err)
	}
}
