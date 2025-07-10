package ratelimiter

import (
	"sync"
	"time"
)

// RateLimiter manages rate limiting for users
type RateLimiter struct {
	mu       sync.RWMutex
	lastSent map[int64]time.Time
	cooldown time.Duration
}

// New creates a new rate limiter with the specified cooldown
func New(cooldown time.Duration) *RateLimiter {
	return &RateLimiter{
		lastSent: make(map[int64]time.Time),
		cooldown: cooldown,
	}
}

// Record records that a message was sent to the user
func (rl *RateLimiter) Record(chatID int64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.lastSent[chatID] = time.Now()
}

// CanSendInSeconds returns how many seconds until the user can receive another message
func (rl *RateLimiter) CanSendInSeconds(chatID int64) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	lastSent, exists := rl.lastSent[chatID]
	if !exists {
		return 0
	}

	elapsed := time.Since(lastSent)
	if elapsed >= rl.cooldown {
		return 0
	}

	remaining := rl.cooldown - elapsed
	return int(remaining.Seconds())
}
