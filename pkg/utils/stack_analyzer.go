package utils

import (
	"bytes"         // Added for bytes.NewReader
	"compress/gzip" // Added for gzip decompression
	"compress/zlib" // Added for deflate decompression
	"fmt"
	"io" // Added for io.ReadAll
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	wappalyze "github.com/projectdiscovery/wappalyzergo"
)

// Global Wappalyzer client
var (
	wappalyzerClient   *wappalyze.Wappalyze
	wappalyzerInitOnce sync.Once
	wappalyzerInitErr  error
)

const versionSeparator = ":"
const htmlResponsesDir = "html_responses"

func initializeWappalyzer() {
	wappalyzerInitOnce.Do(func() {
		var err error
		wappalyzerClient, err = wappalyze.New()
		if err != nil {
			wappalyzerInitErr = fmt.Errorf("failed to initialize wappalyzer client: %w", err)
			log.Println(wappalyzerInitErr)
			return
		}
		log.Println("Wappalyzer client initialized successfully.")
	})
}

type DetectedTechnologyInfo struct {
	Name        string
	Version     string
	Categories  []string
	Description string
	Website     string
	Icon        string
	CPE         string
}

func sanitizeFilename(input string) string {
	s := regexp.MustCompile(`^https?://`).ReplaceAllString(input, "")
	s = regexp.MustCompile(`[^a-zA-Z0-9_.-]+`).ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	s = regexp.MustCompile(`_+`).ReplaceAllString(s, "_")
	if len(s) > 100 {
		s = s[:100]
	}
	if s == "" {
		return "unnamed_response"
	}
	return s
}

func saveHTMLToFile(filePath string, content []byte) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write HTML to file %s: %w", filePath, err)
	}
	log.Printf("Successfully saved HTML to %s\n", filePath)
	return nil
}

// AnalyzeStack fetches a URL, decompress its body if needed,
// analyzes its technology stack, and saves the HTML response.
func AnalyzeStack(targetURL string) ([]DetectedTechnologyInfo, string, error) {
	initializeWappalyzer()
	if wappalyzerInitErr != nil {
		return nil, targetURL, wappalyzerInitErr
	}
	if wappalyzerClient == nil {
		return nil, targetURL, fmt.Errorf("wappalyzer client not available")
	}

	// Assume FetchURL returns a struct FetchResult with fields:
	// Body []byte, Headers http.Header, FinalURL string, StatusCode int, Status string
	// If FetchURL is not defined here, this is a placeholder for its expected behavior.
	// type FetchResult struct {
	// 	Body       []byte
	// 	Headers    http.Header // Assuming http.Header or similar map[string][]string
	// 	FinalURL   string
	// 	StatusCode int
	// 	Status     string
	// }
	fetchResult, err := FetchURL(targetURL) // This is your existing call
	if err != nil {
		finalErrURL := targetURL
		if fetchResult != nil && fetchResult.FinalURL != "" {
			finalErrURL = fetchResult.FinalURL
		}
		return nil, finalErrURL, err
	}

	finalURL := fetchResult.FinalURL

	if fetchResult.StatusCode != 200 { // http.StatusOK
		return nil, finalURL, fmt.Errorf("failed to fetch %s: received status code %d (%s)", targetURL, fetchResult.StatusCode, fetchResult.Status)
	}

	// --- Decompression and Body Processing ---
	var bodyToProcess []byte = fetchResult.Body
	var errDecompress error

	// Check Content-Encoding header. Note: fetchResult.Headers needs to be accessible
	// and behave like http.Header or provide a Get method.
	// If fetchResult.Headers is map[string][]string, access might be:
	//   contentEncodingValues := fetchResult.Headers["Content-Encoding"]
	//   if len(contentEncodingValues) > 0 { contentEncoding = contentEncodingValues[0] }
	// For simplicity, assuming a .Get() method similar to http.Header:
	contentEncoding := ""
	if fetchResult.Headers != nil { // Ensure headers map/object is not nil
		// If fetchResult.Headers is map[string][]string:
		// if ceValues, ok := fetchResult.Headers["Content-Encoding"]; ok && len(ceValues) > 0 {
		// 	contentEncoding = strings.ToLower(strings.TrimSpace(ceValues[0]))
		// }
		// If fetchResult.Headers has a Get method (like http.Header):
		contentEncoding = strings.ToLower(strings.TrimSpace(fetchResult.Headers.Get("Content-Encoding")))
	}

	log.Printf("Response from %s - Content-Encoding: '%s', Content-Type: '%s'", finalURL, contentEncoding, fetchResult.Headers.Get("Content-Type"))

	switch contentEncoding {
	case "gzip":
		gzReader, errGzip := gzip.NewReader(bytes.NewReader(fetchResult.Body))
		if errGzip == nil {
			defer gzReader.Close()
			decompressedBody, errRead := io.ReadAll(gzReader)
			if errRead == nil {
				bodyToProcess = decompressedBody
				log.Printf("Successfully decompressed GZIP body for %s", finalURL)
			} else {
				errDecompress = fmt.Errorf("failed to read gzip decompressed body: %w", errRead)
			}
		} else {
			errDecompress = fmt.Errorf("failed to create gzip reader: %w", errGzip)
		}
	case "deflate":
		zlibReader, errZlib := zlib.NewReader(bytes.NewReader(fetchResult.Body))
		if errZlib == nil {
			defer zlibReader.Close()
			decompressedBody, errRead := io.ReadAll(zlibReader)
			if errRead == nil {
				bodyToProcess = decompressedBody
				log.Printf("Successfully decompressed DEFLATE body for %s", finalURL)
			} else {
				errDecompress = fmt.Errorf("failed to read deflate decompressed body: %w", errRead)
			}
		} else {
			errDecompress = fmt.Errorf("failed to create deflate reader: %w", errZlib)
		}
	case "br":
		log.Printf("Warning: Brotli (br) Content-Encoding detected for %s. Standard library does not support Brotli. Body might remain compressed.", finalURL)
		// Brotli decompression would require an external library, e.g., github.com/andybalholm/brotli
		// For now, we'll proceed with the original body if it's brotli.
	case "":
		// No Content-Encoding or it's an identity encoding. Body is likely plain.
		log.Printf("No specific Content-Encoding or identity encoding for %s. Assuming plain body.", finalURL)
	default:
		log.Printf("Warning: Unsupported Content-Encoding '%s' for %s. Using original body.", contentEncoding, finalURL)
	}

	if errDecompress != nil {
		log.Printf("Warning: Decompression error for %s (encoding: %s): %v. Attempting to use original body for analysis and saving.", finalURL, contentEncoding, errDecompress)
		// bodyToProcess is already fetchResult.Body by default, so we proceed with it.
	}
	// --- End Decompression and Body Processing ---

	// // --- Save HTML content for inspection (using the potentially decompressed body) ---
	// if len(bodyToProcess) > 0 {
	// 	sanitizedBaseName := sanitizeFilename(finalURL)
	// 	timestamp := time.Now().Format("20060102_150405_000")
	// 	htmlFilename := filepath.Join(htmlResponsesDir, fmt.Sprintf("%s_%s.html", sanitizedBaseName, timestamp))

	// 	if errSave := saveHTMLToFile(htmlFilename, bodyToProcess); errSave != nil {
	// 		log.Printf("Warning: Failed to save HTML for %s (to %s): %v\n", finalURL, htmlFilename, errSave)
	// 	}
	// } else {
	// 	log.Printf("Info: No HTML body content to save for %s (original length: %d)\n", finalURL, len(fetchResult.Body))
	// }
	// // --- End HTML saving ---

	// Use the processed (ideally decompressed) body for Wappalyzer
	detectedAppsWithInfo := wappalyzerClient.FingerprintWithInfo(fetchResult.Headers, bodyToProcess)

	var results []DetectedTechnologyInfo
	for appKey, appInfo := range detectedAppsWithInfo {
		name := appKey
		version := ""

		if strings.Contains(appKey, versionSeparator) {
			parts := strings.SplitN(appKey, versionSeparator, 2)
			name = parts[0]
			if len(parts) > 1 {
				version = parts[1]
			}
		}

		results = append(results, DetectedTechnologyInfo{
			Name:        name,
			Version:     version,
			Categories:  appInfo.Categories,
			Description: appInfo.Description,
			Website:     appInfo.Website,
			Icon:        appInfo.Icon,
			CPE:         appInfo.CPE,
		})
	}

	return results, finalURL, nil
}
