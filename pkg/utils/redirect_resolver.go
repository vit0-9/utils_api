package utils

import (
	"fmt"
	"net/http"
	"time"
)

// ResolveRedirect follows HTTP redirects for a given URL and returns the final destination URL.
func ResolveRedirect(initialURL string) (string, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Make a GET request. The client will automatically follow redirects.
	resp, err := client.Get(initialURL)
	if err != nil {
		if resp != nil && resp.Request != nil && resp.Request.URL != nil {
			return resp.Request.URL.String(), fmt.Errorf("failed to get final URL, possibly too many redirects or other error: %w. Last known URL: %s", err, resp.Request.URL.String())
		}
		return "", fmt.Errorf("request failed for %s: %w", initialURL, err)
	}
	defer resp.Body.Close()

	// The final URL after all redirects will be in resp.Request.URL
	finalURL := resp.Request.URL.String()

	// It's good practice to check if the final URL is substantially different from the initial one,
	// especially if no redirects actually occurred.
	if finalURL == initialURL && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// No redirect happened, or it redirected to itself (which is fine)
		// and it's a success status code.
	} else if finalURL == initialURL && resp.StatusCode >= 300 {
		// No redirect, but also not a success code. Could be an error page.
		return finalURL, fmt.Errorf("no redirect from %s, but resulted in status: %s", initialURL, resp.Status)

	}

	return finalURL, nil
}
