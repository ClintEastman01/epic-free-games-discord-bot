package ratelimit

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	tokens    chan struct{}
	ticker    *time.Ticker
	mu        sync.Mutex
	closed    bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewRateLimiter creates a new rate limiter
// rate: number of operations per second
// burst: maximum number of operations that can be performed at once
func NewRateLimiter(rate int, burst int) *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	
	rl := &RateLimiter{
		tokens: make(chan struct{}, burst),
		ticker: time.NewTicker(time.Second / time.Duration(rate)),
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Fill the bucket initially
	for i := 0; i < burst; i++ {
		select {
		case rl.tokens <- struct{}{}:
		default:
			break
		}
	}
	
	// Start the token refill goroutine
	go rl.refill()
	
	return rl
}

// Wait waits for a token to become available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-rl.ctx.Done():
		return rl.ctx.Err()
	}
}

// TryWait attempts to get a token without blocking
func (rl *RateLimiter) TryWait() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// refill adds tokens to the bucket at the specified rate
func (rl *RateLimiter) refill() {
	defer rl.ticker.Stop()
	
	for {
		select {
		case <-rl.ticker.C:
			select {
			case rl.tokens <- struct{}{}:
			default:
				// Bucket is full, skip this token
			}
		case <-rl.ctx.Done():
			return
		}
	}
}

// Close stops the rate limiter
func (rl *RateLimiter) Close() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	if !rl.closed {
		rl.closed = true
		rl.cancel()
		close(rl.tokens)
	}
}

// DiscordRateLimiter provides Discord-specific rate limiting
type DiscordRateLimiter struct {
	global   *RateLimiter
	channels map[string]*RateLimiter
	mu       sync.RWMutex
}

// NewDiscordRateLimiter creates a Discord-specific rate limiter
func NewDiscordRateLimiter() *DiscordRateLimiter {
	return &DiscordRateLimiter{
		global:   NewRateLimiter(50, 1),  // Discord global rate limit
		channels: make(map[string]*RateLimiter),
	}
}

// WaitForChannel waits for permission to send a message to a specific channel
func (drl *DiscordRateLimiter) WaitForChannel(ctx context.Context, channelID string) error {
	// Wait for global rate limit
	if err := drl.global.Wait(ctx); err != nil {
		return err
	}
	
	// Wait for channel-specific rate limit
	drl.mu.RLock()
	channelLimiter, exists := drl.channels[channelID]
	drl.mu.RUnlock()
	
	if !exists {
		drl.mu.Lock()
		// Double-check after acquiring write lock
		if channelLimiter, exists = drl.channels[channelID]; !exists {
			channelLimiter = NewRateLimiter(5, 1) // 5 messages per second per channel
			drl.channels[channelID] = channelLimiter
		}
		drl.mu.Unlock()
	}
	
	return channelLimiter.Wait(ctx)
}

// Close closes all rate limiters
func (drl *DiscordRateLimiter) Close() {
	drl.mu.Lock()
	defer drl.mu.Unlock()
	
	drl.global.Close()
	for _, limiter := range drl.channels {
		limiter.Close()
	}
}