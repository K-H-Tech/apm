package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	// RequestsPerMinute is the maximum requests per minute per client
	RequestsPerMinute int
	// BurstSize is the maximum burst size
	BurstSize int
	// CleanupInterval is how often to clean up expired entries
	CleanupInterval time.Duration
	// KeyFunc extracts the rate limit key from the request
	KeyFunc func(*gin.Context) string
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
		CleanupInterval:   5 * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			// Use user ID if authenticated, otherwise IP
			if userID, exists := c.Get("user_id"); exists {
				return "user:" + userID.(string)
			}
			return "ip:" + c.ClientIP()
		},
	}
}

// rateLimiter implements a simple token bucket rate limiter
type rateLimiter struct {
	config   RateLimitConfig
	buckets  map[string]*tokenBucket
	mu       sync.RWMutex
	stopChan chan struct{}
}

// tokenBucket represents a token bucket for rate limiting
type tokenBucket struct {
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(config RateLimitConfig) *rateLimiter {
	rl := &rateLimiter{
		config:   config,
		buckets:  make(map[string]*tokenBucket),
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// allow checks if a request should be allowed
func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:     float64(rl.config.BurstSize),
			lastUpdate: time.Now(),
		}
		rl.buckets[key] = bucket
	}
	rl.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Calculate tokens to add based on time elapsed
	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdate)
	tokensToAdd := elapsed.Seconds() * float64(rl.config.RequestsPerMinute) / 60.0
	bucket.tokens = min(float64(rl.config.BurstSize), bucket.tokens+tokensToAdd)
	bucket.lastUpdate = now

	// Check if we have tokens available
	if bucket.tokens >= 1 {
		bucket.tokens--
		return true
	}

	return false
}

// remaining returns the number of remaining tokens
func (rl *rateLimiter) remaining(key string) int {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if !exists {
		return rl.config.BurstSize
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	return int(bucket.tokens)
}

// cleanup removes expired buckets
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.doCleanup()
		case <-rl.stopChan:
			return
		}
	}
}

// doCleanup performs the actual cleanup
func (rl *rateLimiter) doCleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.config.CleanupInterval)
	for key, bucket := range rl.buckets {
		bucket.mu.Lock()
		if bucket.lastUpdate.Before(cutoff) {
			delete(rl.buckets, key)
		}
		bucket.mu.Unlock()
	}
}

// stop stops the rate limiter cleanup goroutine
func (rl *rateLimiter) stop() {
	close(rl.stopChan)
}

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	limiter := newRateLimiter(config)

	return func(c *gin.Context) {
		key := config.KeyFunc(c)

		if !limiter.allow(key) {
			remaining := limiter.remaining(key)
			c.Header("X-RateLimit-Limit", string(rune(config.RequestsPerMinute)))
			c.Header("X-RateLimit-Remaining", string(rune(remaining)))
			c.Header("Retry-After", "60")

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": 60,
			})
			return
		}

		c.Next()
	}
}

// EndpointRateLimitConfig holds per-endpoint rate limit configuration
type EndpointRateLimitConfig struct {
	Path              string
	RequestsPerMinute int
	BurstSize         int
}

// EndpointRateLimitMiddleware creates per-endpoint rate limiting
func EndpointRateLimitMiddleware(configs []EndpointRateLimitConfig) gin.HandlerFunc {
	limiters := make(map[string]*rateLimiter)

	for _, cfg := range configs {
		limiters[cfg.Path] = newRateLimiter(RateLimitConfig{
			RequestsPerMinute: cfg.RequestsPerMinute,
			BurstSize:         cfg.BurstSize,
			CleanupInterval:   5 * time.Minute,
			KeyFunc: func(c *gin.Context) string {
				if userID, exists := c.Get("user_id"); exists {
					return "user:" + userID.(string)
				}
				return "ip:" + c.ClientIP()
			},
		})
	}

	return func(c *gin.Context) {
		limiter, exists := limiters[c.FullPath()]
		if !exists {
			c.Next()
			return
		}

		key := limiter.config.KeyFunc(c)
		if !limiter.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded for this endpoint",
			})
			return
		}

		c.Next()
	}
}

// AIRateLimitMiddleware creates rate limiting specifically for AI endpoints
func AIRateLimitMiddleware() gin.HandlerFunc {
	// Lower rate limits for expensive AI operations
	config := RateLimitConfig{
		RequestsPerMinute: 10, // 10 AI requests per minute
		BurstSize:         3,  // Allow burst of 3
		CleanupInterval:   5 * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			// Rate limit by user
			if userID, exists := c.Get("user_id"); exists {
				return "ai:user:" + userID.(string)
			}
			return "ai:ip:" + c.ClientIP()
		},
	}

	limiter := newRateLimiter(config)

	return func(c *gin.Context) {
		key := config.KeyFunc(c)

		if !limiter.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "AI rate limit exceeded",
				"message":     "Too many AI requests. Please wait before making more AI-powered requests.",
				"retry_after": 60,
			})
			return
		}

		c.Next()
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
