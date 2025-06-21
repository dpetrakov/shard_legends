package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Response struct {
	Message string `json:"message"`
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	// Get auth-service URL from environment or use default
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://auth-service:8080" // Default for Docker
	}

	// Create URL for auth endpoint
	targetURL, err := url.Parse(authServiceURL + "/auth")
	if err != nil {
		log.Printf("Error parsing auth service URL: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create new request to auth-service - always use POST for auth endpoint
	proxyReq, err := http.NewRequest(http.MethodPost, targetURL.String(), r.Body)
	if err != nil {
		log.Printf("Error creating proxy request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Copy all headers from original request
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Copy query parameters
	proxyReq.URL.RawQuery = r.URL.RawQuery

	// Add some debug headers to identify the proxy
	proxyReq.Header.Set("X-Proxied-By", "ping-service")
	proxyReq.Header.Set("X-Original-Path", r.URL.Path)

	log.Printf("Proxying %s %s to %s", r.Method, r.URL.Path, targetURL.String())
	log.Printf("Headers: %v", r.Header)

	// Make request to auth-service
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("Error calling auth service: %v", err)
		http.Error(w, "Auth service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}

	log.Printf("Proxied response: %d", resp.StatusCode)
}

func main() {
	http.HandleFunc("/ping", pingHandler)

	port := "8080"
	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
