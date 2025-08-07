# Token-Based Rate Limiter in Go

This project implements a token-based rate limiting service in Go. It uses a token bucket algorithm to control the number of requests a client can make to an API endpoint based on their API key.

## How it Works

The rate limiter works by assigning a "bucket" of tokens to each unique API key. When a request comes in, the service checks if the corresponding API key has any tokens left in its bucket.

1.  **Token Bucket:** Each API key gets a bucket that is periodically refilled with new tokens up to a maximum capacity.
2.  **Request Handling:** For each incoming request, the system attempts to consume one token from the bucket associated with the request's API key.
3.  **Rate Limiting:**
    *   If a token is available, the request is allowed to proceed to the application handler.
    *   If the bucket is empty, the request is rejected with a `429 Too Many Requests` status code, preventing the client from overwhelming the server.
4.  **Token Refill:** Tokens are added back to the bucket at a fixed rate (e.g., 5 tokens per minute). This allows for bursts of requests and ensures that well-behaved clients can continue to access the service.

## Code Explained

The project is structured into several files, each with a specific responsibility.

### `main.go` - The Entry Point

This file is responsible for initializing and starting the web server.

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"rate-limiter/services"
)

func main() {
	// 1. Initialize a new rate limiter.
	// It's configured to allow 5 requests every 60 seconds.
	rateLimiter := services.NewRateLimiter(5, 60) 

	// 2. Define the actual application handlers.
	// These handlers contain the core application logic.
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World")
	})

	worldHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the World")
	})

	// 3. Register the handlers with the rate-limiting middleware.
	// The middleware intercepts requests before they reach the handlers.
	http.Handle("/hello", services.RateLimiterMiddleware(helloHandler, rateLimiter))
	http.Handle("/world", services.RateLimiterMiddleware(worldHandler, rateLimiter))

	// 4. Start the HTTP server.
	fmt.Println("Server started on :8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}
```

### `services/ratelimiter.go` - The Core Logic

This file contains the core token bucket algorithm.

```go
package services

import (
	"sync"
	"time"
)

// RateLimiter holds the state for all clients.
type RateLimiter struct {
	requests map[string]*RequestMetadata // Stores metadata for each API key.
	mutex    sync.Mutex                  // Ensures thread-safe access to the map.
	maxLimit int                         // Maximum tokens allowed in the bucket.
	timeLImit int                        // The time window in seconds (e.g., 60s).
}

// RequestMetadata stores the token count and last seen time for a single client.
type RequestMetadata struct {
	lastSeen   time.Time
	tokenCount int
}

// NewRateLimiter creates a new RateLimiter instance.
func NewRateLimiter(maxLimit int, timeLimit int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*RequestMetadata),
		maxLimit: maxLimit,
		timeLImit: timeLimit,
	}
}

// Allow checks if a request from a given API key should be processed.
func (rl *RateLimiter) Allow(apiKey string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	metadata, exists := rl.requests[apiKey]
	// 1. If the API key is seen for the first time, create a new bucket for it.
	if !exists {
		rl.requests[apiKey] = &RequestMetadata{
			lastSeen:   time.Now(),
			tokenCount: rl.maxLimit - 1, // Consume one token for the current request.
		}
		return true
	}

	// 2. Calculate how many tokens to refill based on time passed.
	refillRate := float64(rl.maxLimit) / float64(rl.timeLImit) // e.g., 5 tokens / 60s
	timePassed := time.Since(metadata.lastSeen).Seconds()
	tokensToAdd := int(timePassed * refillRate)

	// 3. Add the new tokens and update the last seen time.
	if tokensToAdd > 0 {
		metadata.tokenCount += tokensToAdd
		metadata.lastSeen = time.Now()
	}

	// 4. Ensure the bucket doesn't exceed its maximum capacity.
	if metadata.tokenCount > rl.maxLimit {
		metadata.tokenCount = rl.maxLimit
	}

	// 5. If tokens are available, consume one and allow the request.
	if metadata.tokenCount > 0 {
		metadata.tokenCount--
		return true
	}

	// 6. If no tokens are left, deny the request.
	return false
}
```

### `services/apihandler.go` - The Middleware

This file contains the HTTP middleware that enforces the rate limit.

```go
package services

import (
	"net/http"
	apistore "rate-limiter/api-store"
)

// RateLimiterMiddleware wraps an HTTP handler to enforce rate limiting.
func RateLimiterMiddleware(next http.Handler, limiter *RateLimiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Get the API key from the request header.
		apiKey := r.Header.Get("X-API-KEY")
		if apiKey == "" {
			http.Error(w, "Missing API key", http.StatusUnauthorized)
			return
		}

		// 2. Validate the API key.
		if !isValidApiKey(apiKey) {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		// 3. Check if the request is allowed by the rate limiter.
		if !limiter.Allow(apiKey) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// 4. If allowed, forward the request to the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// isValidApiKey checks if the provided key exists in our store.
func isValidApiKey(apiKey string) bool {
	apiKeys := apistore.GetApiKeys()
	_, exists := apiKeys[apiKey]
	return exists
}
```

### `api-store/store.go` - API Key Storage

This file simulates a database or a key store. In a real-world application, you would fetch these keys from a database or a secure configuration service.

```go
package apistore

// GetApiKeys returns a map of valid API keys.
func GetApiKeys() map[string]bool {
	// In a real application, this would come from a database.
	KEYS := map[string]bool{
		"apikey123": true,
		"apikey124": true,
	}

	return KEYS
}
```

## How to Run

1.  **Start the server:**
    ```sh
    go run main.go
    ```

2.  **Send requests using cURL:**
    You can test the rate limiter by sending requests to the `/hello` or `/world` endpoints. Make sure to include a valid API key in the `X-API-KEY` header.

    ```sh
    # This request will succeed
    curl -H "X-API-KEY: apikey123" http://localhost:8083/hello

    # If you send more than 5 requests in 60 seconds, you will get an error
    curl -H "X-API-KEY: apikey123" http://localhost:8083/hello
    # Output: Rate limit exceeded
    ```
