package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type feedResponse struct {
	Items []feedItem `json:"items"`
	At    string     `json:"at"`
}

type feedItem struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/feed", feedHandler)

	addr := ":" + port
	log.Printf("[feed-service] listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, logRequests(mux)))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","service":"feed-service"}`))
}

func feedHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[feed-service] feed request from %s", r.RemoteAddr)

	resp := feedResponse{
		At: time.Now().UTC().Format(time.RFC3339),
		Items: []feedItem{
			{ID: 1, Title: "Hello", Content: "First mock post"},
			{ID: 2, Title: "World", Content: "Second mock post"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[feed-service] encode error: %v", err)
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[feed-service] %s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
