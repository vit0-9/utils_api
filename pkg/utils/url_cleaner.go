package utils

import (
	"embed"
	"encoding/json"
	"log"
	"net/url"
	"sort"
	"strings"
	"sync"
)

//go:embed tracking_params.json
var trackingParamsJSON embed.FS

// TrackingParamDetail defines the structure for each tracking parameter's metadata.
type TrackingParamDetail struct {
	Key         string `json:"key"`
	MatchType   string `json:"match_type,omitempty"` // "exact" or "prefix"
	Company     string `json:"company"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// RemovedParamInfo holds information about a removed tracking parameter.
type RemovedParamInfo struct {
	Parameter   string `json:"parameter"`
	Value       string `json:"value"`
	Company     string `json:"company"`
	Type        string `json:"type"`
	Description string `json:"description"`
	MatchedRule string `json:"matched_rule"` // The key from tracking_params.json that matched
}

var (
	exactMatchParams  map[string]TrackingParamDetail
	prefixMatchParams []TrackingParamDetail // Slice for prefix rules
	loadOnce          sync.Once
	loadErr           error
)

func loadTrackingDefinitions() {
	loadOnce.Do(func() {
		fileData, err := trackingParamsJSON.ReadFile("tracking_params.json")
		if err != nil {
			loadErr = err
			log.Printf("Error reading embedded tracking_params.json: %v", err)
			return
		}

		var params []TrackingParamDetail
		if err = json.Unmarshal(fileData, &params); err != nil {
			loadErr = err
			log.Printf("Error unmarshalling tracking_params.json: %v", err)
			return
		}

		exactMatchParams = make(map[string]TrackingParamDetail)
		prefixMatchParams = []TrackingParamDetail{} // Initialize slice

		for _, p := range params {
			// Default to "exact" if MatchType is empty
			if p.MatchType == "" {
				p.MatchType = "exact"
			}

			lcKey := strings.ToLower(p.Key) // Store definition keys in lowercase
			p.Key = lcKey                   // Ensure the stored key is lowercase for matching

			if p.MatchType == "prefix" {
				prefixMatchParams = append(prefixMatchParams, p)
			} else { // "exact"
				exactMatchParams[lcKey] = p
			}
		}
		log.Printf("Successfully loaded tracking parameter definitions. Exact: %d, Prefix: %d", len(exactMatchParams), len(prefixMatchParams))
	})
}

// CleanURLResult holds the result of the cleaning operation.
type CleanURLResult struct {
	CleanedURL    string
	RemovedParams []RemovedParamInfo
}

// CleanURL removes tracking parameters and provides details about what was removed.
func CleanURL(rawURL string) (CleanURLResult, error) {
	loadTrackingDefinitions()
	if loadErr != nil {
		return CleanURLResult{}, loadErr
	}
	if (exactMatchParams == nil || len(exactMatchParams) == 0) && (prefixMatchParams == nil || len(prefixMatchParams) == 0) {
		log.Println("Warning: Tracking parameter definitions are empty. No parameters will be removed based on definitions.")
	}

	result := CleanURLResult{
		RemovedParams: []RemovedParamInfo{},
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return result, err
	}

	query := parsedURL.Query()
	if len(query) == 0 {
		result.CleanedURL = parsedURL.String()
		return result, nil
	}

	cleanedQuery := url.Values{}
	var keptKeys []string
	removedThisRound := make(map[string]bool) // To avoid double-removing if a param matches multiple rules (though unlikely with current logic)

	for key, values := range query {
		if removedThisRound[key] {
			continue
		}
		lowercaseKey := strings.ToLower(key)
		var matchedDetail TrackingParamDetail
		var foundMatch bool
		var matchedRuleKey string

		// 1. Check exact matches (more specific)
		if detail, ok := exactMatchParams[lowercaseKey]; ok {
			matchedDetail = detail
			foundMatch = true
			matchedRuleKey = detail.Key
		}

		// 2. If no exact match, check prefix matches
		if !foundMatch {
			for _, prefixDetail := range prefixMatchParams {
				if strings.HasPrefix(lowercaseKey, prefixDetail.Key) {
					matchedDetail = prefixDetail
					foundMatch = true
					matchedRuleKey = prefixDetail.Key // This is the prefix rule key
					break                             // Take the first prefix match
				}
			}
		}

		if foundMatch {
			removedThisRound[key] = true // Mark original key as processed
			for _, value := range values {
				result.RemovedParams = append(result.RemovedParams, RemovedParamInfo{
					Parameter:   key, // Report original key
					Value:       value,
					Company:     matchedDetail.Company,
					Type:        matchedDetail.Type,
					Description: matchedDetail.Description,
					MatchedRule: matchedRuleKey,
				})
			}
			continue // Skip adding to cleanedQuery
		}

		// Keep non-tracking parameters
		for _, value := range values {
			cleanedQuery.Add(key, value)
		}
		if len(values) > 0 {
			keptKeys = append(keptKeys, key)
		}
	}

	sort.Strings(keptKeys)

	finalQueryValues := url.Values{}
	for _, k := range keptKeys {
		if vals, ok := cleanedQuery[k]; ok {
			for _, v := range vals {
				finalQueryValues.Add(k, v)
			}
		}
	}
	parsedURL.RawQuery = finalQueryValues.Encode()
	result.CleanedURL = parsedURL.String()

	return result, nil
}
