package services

import (
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string]*RequestMetadata
	mutex    sync.Mutex
	maxLimit int
	timeLImit int
}

type RequestMetadata struct {
	lastSeen   time.Time
	tokenCount int
}

func NewRateLimiter(maxLimit int, timeLimit int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*RequestMetadata),
		maxLimit: maxLimit,
		timeLImit: timeLimit,
	}
}

func (rl *RateLimiter) Allow(apiKey string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	metadata, exists := rl.requests[apiKey]
	if !exists {
		rl.requests[apiKey] = &RequestMetadata{
			lastSeen:   time.Now(),
			tokenCount: rl.maxLimit - 1,
		}
		return true
	}

	refillRate := float64(rl.maxLimit) / float64(rl.timeLImit)
	timePassed := time.Since(metadata.lastSeen).Seconds()
	tokensToAdd := int(timePassed * refillRate)

	if tokensToAdd > 0 {
		metadata.tokenCount = metadata.tokenCount + tokensToAdd
		metadata.lastSeen = time.Now()
	}

	if metadata.tokenCount > rl.maxLimit {
		metadata.tokenCount = rl.maxLimit
	}

	if metadata.tokenCount > 0 {
		metadata.tokenCount--
		return true
	}

	return false
}