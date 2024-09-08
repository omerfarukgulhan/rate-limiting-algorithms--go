package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

type TokenBucket struct {
	capacity     int
	tokens       int
	refillRate   int
	refillPeriod time.Duration
	mutex        sync.Mutex
}

func NewTokenBucket(capacity, refillRate int, refillPeriod time.Duration) *TokenBucket {
	tb := &TokenBucket{
		capacity:     capacity,
		tokens:       capacity,
		refillRate:   refillRate,
		refillPeriod: refillPeriod,
	}

	go tb.startRefilling()

	return tb
}

func (tb *TokenBucket) startRefilling() {
	ticker := time.NewTicker(tb.refillPeriod)
	for range ticker.C {
		tb.mutex.Lock()
		tb.tokens += tb.refillRate
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.mutex.Unlock()
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

func main() {
	bucket := NewTokenBucket(5, 1, time.Second*10)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		if !bucket.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			c.Abort()
			return
		}
		c.Next()
	})

	r.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
	})

	err := r.Run(":8080")
	if err != nil {
		return
	}
}
