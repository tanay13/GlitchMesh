package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	glitchmeshURL := os.Getenv("GLITCHMESH_URL")
	if glitchmeshURL == "" {
		glitchmeshURL = "http://glitchmesh:9000"
	}

	feedServiceName := os.Getenv("FEED_SERVICE_NAME")
	if feedServiceName == "" {
		feedServiceName = "feed-service"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/feed", feedHandler(glitchmeshURL, feedServiceName))
	mux.HandleFunc("/api/feed/create", createNewFeed(glitchmeshURL, feedServiceName))

	addr := ":" + port
	log.Printf("[gateway] listening on %s", addr)
	log.Printf("[gateway] upstream via glitchmesh=%s service=%s", glitchmeshURL, feedServiceName)
	log.Fatal(http.ListenAndServe(addr, logRequests(mux)))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","service":"gateway"}`))
}

func createNewFeed(glitchmeshURL, serviceName string) http.HandlerFunc {
	client := &http.Client{Timeout: 30 * time.Second}
	proxyURL := fmt.Sprintf("%s/redirect/%s/feed/create", glitchmeshURL, serviceName)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[gateway] calling glitchmesh proxy url=%s", proxyURL)

		start := time.Now()
		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, proxyURL, r.Body)
		if err != nil {
			log.Printf("[gateway] build request error: %v", err)
			http.Error(w, "failed to build upstream request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[gateway] glitchmesh request error: %v", err)
			http.Error(w, "upstream request failed", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[gateway] read response error: %v", err)
			http.Error(w, "failed to read upstream response", http.StatusBadGateway)
			return
		}

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(body)

		log.Printf("[gateway] glitchmesh response status=%d elapsed=%s", resp.StatusCode, time.Since(start))
	}
}

func feedHandler(glitchmeshURL, serviceName string) http.HandlerFunc {
	client := &http.Client{Timeout: 30 * time.Second}
	proxyURL := fmt.Sprintf("%s/redirect/%s/feed", glitchmeshURL, serviceName)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[gateway] calling glitchmesh proxy url=%s", proxyURL)

		start := time.Now()
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, proxyURL, nil)
		if err != nil {
			log.Printf("[gateway] build request error: %v", err)
			http.Error(w, "failed to build upstream request", http.StatusInternalServerError)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[gateway] glitchmesh request error: %v", err)
			http.Error(w, "upstream request failed", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[gateway] read response error: %v", err)
			http.Error(w, "failed to read upstream response", http.StatusBadGateway)
			return
		}

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(body)

		log.Printf("[gateway] glitchmesh response status=%d elapsed=%s", resp.StatusCode, time.Since(start))
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[gateway] %s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
