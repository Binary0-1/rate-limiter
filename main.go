package main

import (
	"fmt"
	"log"
	"net/http"
	"rate-limiter/services"
)

func main() {
	rateLimiter := services.NewRateLimiter(5, 60) // 5 requests per 60 seconds

	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World")
	})

	worldHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the World")
	})

	http.Handle("/hello", services.RateLimiterMiddleware(helloHandler, rateLimiter))
	http.Handle("/world", services.RateLimiterMiddleware(worldHandler, rateLimiter))

	fmt.Println("Server started on :8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}