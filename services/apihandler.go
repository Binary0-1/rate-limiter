package services

import (
	"net/http"
	apistore "rate-limiter/api-store"
)

func RateLimiterMiddleware(next http.Handler, limiter *RateLimiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")
		if apiKey == "" {
			http.Error(w, "Missing API key", http.StatusUnauthorized)
			return
		}

		if !isValidApiKey(apiKey) {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		if !limiter.Allow(apiKey) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isValidApiKey(apiKey string) bool {
	apiKeys := apistore.GetApiKeys()
	_, exists := apiKeys[apiKey]
	return exists
}