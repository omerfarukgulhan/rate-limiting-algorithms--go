package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

type FixedWindowRateLimiter struct {
	limit       int
	window      time.Duration
	requestLogs map[string][]time.Time
	mu          sync.Mutex
}

func NewFixedWindowRateLimiter(limit int, window time.Duration) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		limit:       limit,
		window:      window,
		requestLogs: make(map[string][]time.Time),
	}
}

func (f *FixedWindowRateLimiter) Allow(ip string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now()
	requests, exists := f.requestLogs[ip]
	if !exists {
		requests = []time.Time{}
	}

	windowStart := now.Add(-f.window)
	i := 0
	for ; i < len(requests) && requests[i].Before(windowStart); i++ {
	}

	requests = requests[i:]
	if len(requests) >= f.limit {
		return false
	}

	requests = append(requests, now)
	f.requestLogs[ip] = requests

	return true
}

func RateLimiter(rateLimiter *FixedWindowRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rateLimiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func main() {
	rateLimiter := NewFixedWindowRateLimiter(5, time.Minute)

	server := gin.Default()

	server.Use(RateLimiter(rateLimiter))

	server.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
	})

	if err := server.Run(":8080"); err != nil {
		panic(err)
	}
}
