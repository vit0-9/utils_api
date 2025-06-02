package utils

import (
	"fmt"
	"log"
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

func initializeWappalyzer() {
	wappalyzerInitOnce.Do(func() {
		var err error
		// Use wappalyzergo.New() as per the user's provided library structure
		wappalyzerClient, err = wappalyze.New()
		if err != nil {
			wappalyzerInitErr = fmt.Errorf("failed to initialize wappalyzer client: %w", err)
			log.Println(wappalyzerInitErr)
			return
		}
		log.Println("Wappalyzer client initialized successfully.")
	})
}

// DetectedTechnologyInfo is the internal struct used by the utility, mirroring models.DetectedTechnology
type DetectedTechnologyInfo struct {
	Name        string
	Version     string
	Categories  []string
	Description string
	Website     string
	Icon        string
	CPE         string
}

// AnalyzeStack fetches a URL and analyzes its technology stack.
func AnalyzeStack(targetURL string) ([]DetectedTechnologyInfo, string, error) {
	initializeWappalyzer()
	if wappalyzerInitErr != nil {
		return nil, targetURL, wappalyzerInitErr
	}
	if wappalyzerClient == nil {
		return nil, targetURL, fmt.Errorf("wappalyzer client not available")
	}

	// Use the new FetchURL utility
	fetchResult, err := FetchURL(targetURL)
	if err != nil {
		// FetchURL already formats the error well, including the original targetURL
		// If FetchResult is nil, finalURL might not be available, so use targetURL
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

	// Use FingerprintWithInfo to get rich details
	// The map key is `appName` or `appName:version`
	// The value is `wappalyzergo.AppInfo` which contains Categories as []string
	detectedAppsWithInfo := wappalyzerClient.FingerprintWithInfo(fetchResult.Headers, fetchResult.Body)

	var results []DetectedTechnologyInfo
	for appKey, appInfo := range detectedAppsWithInfo {
		name := appKey
		version := ""

		// Wappalyzergo's FingerprintWithInfo returns keys that might include version
		// e.g. "Apache:2.4.41"
		// The UniqueFingerprints.GetValues() also formats it this way.
		if strings.Contains(appKey, versionSeparator) {
			parts := strings.SplitN(appKey, versionSeparator, 2)
			name = parts[0]
			if len(parts) > 1 {
				version = parts[1]
			}
		}

		// If appInfo itself contains a version (sometimes Wappalyzer fingerprints have version capture groups)
		// we might prefer that. However, the primary version indication from wappalyzergo
		// often comes from the formatted key if confidence is high enough for a versioned match.
		// For now, we'll rely on parsing the key. The `AppInfo` struct itself from `wappalyzergo`
		// does not seem to have a `Version` field directly based on `AppInfoFromFingerprint`.

		results = append(results, DetectedTechnologyInfo{
			Name:        name,
			Version:     version,
			Categories:  appInfo.Categories, // This is already []string
			Description: appInfo.Description,
			Website:     appInfo.Website,
			Icon:        appInfo.Icon,
			CPE:         appInfo.CPE,
		})
	}

	return results, finalURL, nil
}
