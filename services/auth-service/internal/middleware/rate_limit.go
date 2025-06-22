package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements token bucket rate limiting per IP
type RateLimiter struct {
	logger    *slog.Logger
	clients   map[string]*TokenBucket
	mutex     sync.RWMutex
	rate      time.Duration // Time between tokens
	capacity  int           // Maximum tokens in bucket
	cleanupTicker *time.Ticker
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	tokens     int
	lastRefill time.Time
	mutex      sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int, logger *slog.Logger) *RateLimiter {
	rl := &RateLimiter{
		logger:   logger,
		clients:  make(map[string]*TokenBucket),
		rate:     time.Minute / time.Duration(requestsPerMinute),
		capacity: requestsPerMinute,
		cleanupTicker: time.NewTicker(5 * time.Minute), // Cleanup every 5 minutes
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Limit returns a middleware function that implements rate limiting
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		// Check if request is allowed
		if !rl.allow(clientIP) {
			rl.logger.Warn("Rate limit exceeded",
				"ip", clientIP,
				"path", c.Request.URL.Path,
				"method", c.Request.Method)
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate_limit_exceeded", 
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	bucket, exists := rl.clients[ip]
	if !exists {
		bucket = &TokenBucket{
			tokens:     rl.capacity - 1, // Use one token for this request
			lastRefill: time.Now(),
		}
		rl.clients[ip] = bucket
		return true
	}

	return bucket.consume(rl.rate, rl.capacity)
}

// consume attempts to consume a token from the bucket
func (tb *TokenBucket) consume(rate time.Duration, capacity int) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	
	// Calculate how many tokens to add based on time passed
	tokensToAdd := int(now.Sub(tb.lastRefill) / rate)
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > capacity {
			tb.tokens = capacity
		}
		tb.lastRefill = now
	}

	// Try to consume a token
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// cleanup removes stale clients to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	for range rl.cleanupTicker.C {
		rl.mutex.Lock()
		
		cutoff := time.Now().Add(-10 * time.Minute)
		for ip, bucket := range rl.clients {
			bucket.mutex.Lock()
			if bucket.lastRefill.Before(cutoff) {
				delete(rl.clients, ip)
			}
			bucket.mutex.Unlock()
		}
		
		rl.logger.Debug("Rate limiter cleanup completed",
			"active_clients", len(rl.clients))
		
		rl.mutex.Unlock()
	}
}

// Close stops the cleanup goroutine
func (rl *RateLimiter) Close() {
	if rl.cleanupTicker != nil {
		rl.cleanupTicker.Stop()
	}
}