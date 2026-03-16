package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	DefaultAPIKeyRateLimit = 60  // requests per minute
	DefaultJWTRateLimit    = 120 // requests per minute
)

type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

func (b *tokenBucket) allow() bool {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

func (b *tokenBucket) retryAfter() float64 {
	if b.refillRate == 0 {
		return 60
	}
	return (1 - b.tokens) / b.refillRate
}

type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*tokenBucket),
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			if now.Sub(bucket.lastRefill) > 10*time.Minute {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(key string, ratePerMinute int) (bool, float64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	maxTokens := float64(ratePerMinute)
	refillRate := float64(ratePerMinute) / 60.0

	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:     maxTokens,
			maxTokens:  maxTokens,
			refillRate: refillRate,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = bucket
	}

	if bucket.maxTokens != maxTokens {
		bucket.maxTokens = maxTokens
		bucket.refillRate = refillRate
	}

	if bucket.allow() {
		return true, 0
	}
	return false, bucket.retryAfter()
}

const ContextAuthMethod = "authMethod"

func RateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(ContextUserID)
		if !exists {
			c.Next()
			return
		}

		authMethod, _ := c.Get(ContextAuthMethod)
		method, _ := authMethod.(string)

		var key string
		var rateLimit int

		switch method {
		case "apikey":
			apiKeyID, _ := c.Get("apiKeyID")
			key = fmt.Sprintf("apikey:%v", apiKeyID)
			if override, ok := c.Get("apiKeyRateLimit"); ok {
				if rl, ok := override.(int); ok && rl > 0 {
					rateLimit = rl
				}
			}
			if rateLimit == 0 {
				rateLimit = DefaultAPIKeyRateLimit
			}
		default:
			key = fmt.Sprintf("jwt:%v", userID)
			rateLimit = DefaultJWTRateLimit
		}

		allowed, retryAfter := limiter.Allow(key, rateLimit)
		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%.0f", retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":      "rate limit exceeded",
				"retryAfter": fmt.Sprintf("%.0f", retryAfter),
			})
			return
		}

		c.Next()
	}
}
