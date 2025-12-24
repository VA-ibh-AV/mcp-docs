package indexer

import (
	"context"
	"sync"
	"time"
)

// RateLimiter provides per-domain rate limiting
type RateLimiter struct {
	mu            sync.Mutex
	domains       map[string]*domainLimiter
	defaultRate   float64       // requests per second
	defaultBurst  int           // burst size
	cleanupTicker *time.Ticker
	done          chan struct{}
}

// domainLimiter tracks rate limiting state for a single domain
type domainLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		domains:      make(map[string]*domainLimiter),
		defaultRate:  requestsPerSecond,
		defaultBurst: burst,
		done:         make(chan struct{}),
	}

	// Start cleanup goroutine to remove stale domain limiters
	rl.cleanupTicker = time.NewTicker(5 * time.Minute)
	go rl.cleanup()

	return rl
}

// Wait blocks until a request can be made to the given domain
// Returns error if context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context, domain string) error {
	limiter := rl.getLimiter(domain)

	for {
		limiter.mu.Lock()

		// Refill tokens based on time elapsed
		now := time.Now()
		elapsed := now.Sub(limiter.lastRefill).Seconds()
		limiter.tokens += elapsed * limiter.refillRate
		if limiter.tokens > limiter.maxTokens {
			limiter.tokens = limiter.maxTokens
		}
		limiter.lastRefill = now

		// Try to consume a token
		if limiter.tokens >= 1 {
			limiter.tokens--
			limiter.mu.Unlock()
			return nil
		}

		// Calculate wait time
		waitTime := time.Duration((1 - limiter.tokens) / limiter.refillRate * float64(time.Second))
		limiter.mu.Unlock()

		// Wait with context
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Continue loop to try again
		}
	}
}

// TryAcquire attempts to acquire permission without blocking
// Returns true if allowed, false if rate limited
func (rl *RateLimiter) TryAcquire(domain string) bool {
	limiter := rl.getLimiter(domain)

	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(limiter.lastRefill).Seconds()
	limiter.tokens += elapsed * limiter.refillRate
	if limiter.tokens > limiter.maxTokens {
		limiter.tokens = limiter.maxTokens
	}
	limiter.lastRefill = now

	// Try to consume
	if limiter.tokens >= 1 {
		limiter.tokens--
		return true
	}

	return false
}

// getLimiter returns the limiter for a domain, creating if needed
func (rl *RateLimiter) getLimiter(domain string) *domainLimiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.domains[domain]
	if !exists {
		limiter = &domainLimiter{
			tokens:     float64(rl.defaultBurst), // Start with burst capacity
			maxTokens:  float64(rl.defaultBurst),
			refillRate: rl.defaultRate,
			lastRefill: time.Now(),
		}
		rl.domains[domain] = limiter
	}

	return limiter
}

// SetDomainRate sets a custom rate for a specific domain
func (rl *RateLimiter) SetDomainRate(domain string, requestsPerSecond float64, burst int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.domains[domain] = &domainLimiter{
		tokens:     float64(burst),
		maxTokens:  float64(burst),
		refillRate: requestsPerSecond,
		lastRefill: time.Now(),
	}
}

// cleanup removes stale domain limiters periodically
func (rl *RateLimiter) cleanup() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.mu.Lock()
			// Remove domains that haven't been used recently
			// (This is a simple approach; in production you might
			// want to track last access time)
			if len(rl.domains) > 1000 {
				// Clear all if too many (simple approach)
				rl.domains = make(map[string]*domainLimiter)
			}
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}

// Close stops the rate limiter
func (rl *RateLimiter) Close() {
	rl.cleanupTicker.Stop()
	close(rl.done)
}

// Stats returns current stats about the rate limiter
func (rl *RateLimiter) Stats() map[string]interface{} {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return map[string]interface{}{
		"domains_tracked": len(rl.domains),
		"default_rate":    rl.defaultRate,
		"default_burst":   rl.defaultBurst,
	}
}
