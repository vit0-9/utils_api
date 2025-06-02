package utils

import (
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"
)

// defaultUserAgents is a list of common browser User-Agent strings.
// It's good to keep this list updated or even load it from a config.
var defaultUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:99.0) Gecko/20100101 Firefox/99.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:99.0) Gecko/20100101 Firefox/99.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/100.0.1185.36", // Microsoft Edge (Chromium)
	"Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 11; SM-G991U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Mobile Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.88 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.3 Safari/605.1.15", // Safari on macOS
}

var (
	httpClient     *http.Client
	httpClientOnce sync.Once
	randSource     rand.Source
)

func init() {
	// Initialize random source for user agent selection
	randSource = rand.NewSource(time.Now().UnixNano())
}

// initializeHTTPClient creates a shared HTTP client with good defaults.
func initializeHTTPClient() {
	httpClientOnce.Do(func() {
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			// Log or handle error, for now, we'll proceed without a cookie jar if it fails
			// This should ideally not fail with publicsuffix.List unless there's a major issue.
			// Consider logging this error.
		}

		// Custom transport with more browser-like settings
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12, // Enforce modern TLS
				// CipherSuites: you can specify a list of cipher suites if needed for very specific targets
			},
			DialContext: (&net.Dialer{
				Timeout:   15 * time.Second, // Connection timeout
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10, // More realistic than default 2 for browsers
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ForceAttemptHTTP2:     true, // Attempt HTTP/2
		}

		httpClient = &http.Client{
			Timeout:   30 * time.Second, // Overall request timeout
			Jar:       jar,
			Transport: transport,
			// Default redirect policy: follow up to 10 redirects.
			// If you need to prevent redirects for specific utilities,
			// you'd create a request and use client.Do(req) with a client
			// that has a CheckRedirect policy, or a new client.
			// For general fetching (like for Wappalyzer), following is good.
		}
	})
}

// GetRandomUserAgent selects a User-Agent string randomly from the predefined list.
func GetRandomUserAgent() string {
	r := rand.New(randSource) // Create a new rand.Rand for thread-safety if this func is called concurrently often
	return defaultUserAgents[r.Intn(len(defaultUserAgents))]
}

// FetchResult encapsulates the results of an HTTP fetch operation.
type FetchResult struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       []byte
	FinalURL   string // URL after all redirects
}

// FetchURL performs an HTTP GET request to the targetURL with browser-like headers
// and returns the response details.
func FetchURL(targetURL string) (*FetchResult, error) {
	initializeHTTPClient() // Ensure our shared client is initialized

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", targetURL, err)
	}

	// Set common browser headers
	req.Header.Set("User-Agent", GetRandomUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("DNT", "1") // Do Not Track
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", targetURL, err)
	}

	result := &FetchResult{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
		Body:       bodyBytes,
		FinalURL:   resp.Request.URL.String(), // URL after redirects
	}

	return result, nil
}
