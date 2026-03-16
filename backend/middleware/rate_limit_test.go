package middleware

import (
	"testing"
)

func TestRateLimiter(t *testing.T) {
	t.Run("AllowsWithinLimit", func(t *testing.T) {
		rl := NewRateLimiter()
		for i := 0; i < 60; i++ {
			allowed, _ := rl.Allow("test-key", 60)
			if !allowed {
				t.Fatalf("expected request %d to be allowed", i+1)
			}
		}
	})

	t.Run("BlocksOverLimit", func(t *testing.T) {
		rl := NewRateLimiter()
		// Exhaust the bucket
		for i := 0; i < 60; i++ {
			rl.Allow("test-key-2", 60)
		}
		allowed, retryAfter := rl.Allow("test-key-2", 60)
		if allowed {
			t.Fatal("expected request to be blocked after exceeding limit")
		}
		if retryAfter <= 0 {
			t.Errorf("expected positive retryAfter, got %f", retryAfter)
		}
	})

	t.Run("SeparateKeysIndependent", func(t *testing.T) {
		rl := NewRateLimiter()
		// Exhaust key A
		for i := 0; i < 60; i++ {
			rl.Allow("key-a", 60)
		}
		// Key B should still work
		allowed, _ := rl.Allow("key-b", 60)
		if !allowed {
			t.Fatal("expected different key to be independent")
		}
	})

	t.Run("CustomRateLimit", func(t *testing.T) {
		rl := NewRateLimiter()
		// Use a rate of 10/min
		for i := 0; i < 10; i++ {
			allowed, _ := rl.Allow("custom-key", 10)
			if !allowed {
				t.Fatalf("expected request %d to be allowed with limit 10", i+1)
			}
		}
		allowed, _ := rl.Allow("custom-key", 10)
		if allowed {
			t.Fatal("expected request to be blocked after exceeding custom limit")
		}
	})
}
