package utils

import (
	"net/url"
	"strings"
)

// UTMGeneratorOptions defines formatting assistance options.
// THIS IS NOW DEFINED IN THE UTILS PACKAGE.
type UTMGeneratorOptions struct {
	ForceLowercase   bool   `json:"force_lowercase"`
	SpaceReplacement string `json:"space_replacement,omitempty"` // e.g., "_" or "-"
}

// FullUTMParams holds all potential UTM parameters for a single link.
// THIS IS ALSO DEFINED HERE (OR REMAINS AS IS IF IT WAS ALREADY HERE).
type FullUTMParams struct {
	Source   string
	Medium   string
	Campaign string
	Term     string
	Content  string
}

// FormatUTMValue applies formatting options to a UTM parameter value.
// It now uses the locally defined UTMGeneratorOptions.
func FormatUTMValue(value string, options *UTMGeneratorOptions) string {
	if value == "" || options == nil {
		return value
	}
	processedValue := value
	if options.ForceLowercase {
		processedValue = strings.ToLower(processedValue)
	}
	if options.SpaceReplacement != "" {
		processedValue = strings.ReplaceAll(processedValue, " ", options.SpaceReplacement)
	}
	return processedValue
}

// GenerateUTMLink constructs a single URL with UTM parameters.
// It now uses the locally defined UTMGeneratorOptions and FullUTMParams.
func GenerateUTMLink(baseURL string, params FullUTMParams, options *UTMGeneratorOptions) (string, error) {
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	query := parsedBaseURL.Query()

	if p := FormatUTMValue(params.Source, options); p != "" {
		query.Set("utm_source", p)
	}
	if p := FormatUTMValue(params.Medium, options); p != "" {
		query.Set("utm_medium", p)
	}
	if p := FormatUTMValue(params.Campaign, options); p != "" {
		query.Set("utm_campaign", p)
	}
	if p := FormatUTMValue(params.Term, options); p != "" {
		query.Set("utm_term", p)
	}
	if p := FormatUTMValue(params.Content, options); p != "" {
		query.Set("utm_content", p)
	}

	parsedBaseURL.RawQuery = query.Encode()
	return parsedBaseURL.String(), nil
}
