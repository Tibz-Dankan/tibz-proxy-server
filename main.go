package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Middleware to log every incoming request, including the time taken
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Record the start time
		startTime := time.Now()

		// Log the incoming request
		currentTime := startTime.Format("2006-01-02 15:04:05")
		log.Printf("[%s] %s %s %s\n", currentTime, req.Method, req.URL, req.RemoteAddr)

		// Pass the request to the next handler
		next.ServeHTTP(w, req)

		// Calculate the time taken for the request
		duration := time.Since(startTime)

		// Log the time taken for the request
		log.Printf("[%s] Request took %s\n", currentTime, duration)
	})
}

// Logger to log requests and responses with timestamps
func logRequestResponse(req *http.Request, resp *http.Response) {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[%s] %s %s %s\n", currentTime, req.Method, req.URL, req.RemoteAddr)

	if resp != nil {
		log.Printf("[%s] Response Status: %s\n", currentTime, resp.Status)
		for key, value := range resp.Header {
			log.Printf("[%s] Response Header: %s: %s\n", currentTime, key, value)
		}
	}
}

// Handle incoming requests and forward them
func handleRequestAndForward(w http.ResponseWriter, req *http.Request) {
	// Clone the original request to forward it
	originalURL := req.URL
	targetURL, err := url.Parse(originalURL.String())
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		log.Printf("Error parsing URL: %v", err)
		return
	}

	// Create a new HTTP client with timeouts
	client := &http.Client{
		Timeout: 10 * time.Second, // Set a global timeout for the client
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second, // Set timeout for the connection attempt
			}).Dial,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Optionally, log redirects
			log.Printf("Redirecting to %s", req.URL)
			return nil
		},
	}

	req.RequestURI = ""
	req.URL = targetURL
	req.Host = targetURL.Host

	// Forward the request to the destination server
	resp, err := client.Do(req)
	if err != nil {
		// Handle various types of errors with more context
		if netErr, ok := err.(*net.OpError); ok {
			log.Printf("Network error: %v\n", netErr)
			http.Error(w, "Network error occurred", http.StatusBadGateway)
		} else if urlErr, ok := err.(*url.Error); ok {
			log.Printf("URL error: %v\n", urlErr)
			http.Error(w, "Invalid URL", http.StatusBadRequest)
		} else {
			log.Printf("Error forwarding request: %v\n", err)
			http.Error(w, "Error forwarding request", http.StatusInternalServerError)
		}
		return
	}
	defer resp.Body.Close()

	// Log the request and response
	logRequestResponse(req, resp)

	// Copy the response headers to the original response
	for key, value := range resp.Header {
		w.Header()[key] = value
	}

	// Write the status code and body back to the client
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	// Open log file for appending logs
	logFile, err := os.OpenFile("proxy.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()

	// Create a multi-writer to log to both terminal and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Set log output to the multi-writer (both terminal and file)
	log.SetOutput(multiWriter)

	// Set up the HTTP handler
	http.HandleFunc("/", handleRequestAndForward)

	// Wrap the handler with the logging middleware
	loggedHandler := loggingMiddleware(http.DefaultServeMux)

	// Start the proxy server
	port := "8088"
	fmt.Printf("Proxy server is running on port %s\n", port)
	if err := http.ListenAndServe(":"+port, loggedHandler); err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}
