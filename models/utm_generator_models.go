package models

import "github.com/vit0-9/utils_api/pkg/utils"

// UTMParameterSet represents a single set of UTM parameters for one generated URL.
type UTMParameterSet struct {
	Source   string `json:"utm_source" binding:"required"`
	Medium   string `json:"utm_medium" binding:"required"`
	Campaign string `json:"utm_campaign,omitempty"`
	Term     string `json:"utm_term,omitempty"`
	Content  string `json:"utm_content,omitempty"`
}

// UTMGeneratorRequest is the structure for the UTM generation request.
// It now uses utils.UTMGeneratorOptions.
type UTMGeneratorRequest struct {
	BaseURL      string `json:"base_url" binding:"required,url"`
	CommonParams struct {
		Campaign string `json:"utm_campaign" binding:"required"`
		Term     string `json:"utm_term,omitempty"`
		Content  string `json:"utm_content,omitempty"`
	} `json:"common_params" binding:"required"`
	VariableSets []UTMParameterSet          `json:"variable_sets" binding:"required,dive"`
	Options      *utils.UTMGeneratorOptions `json:"options,omitempty"`
}

// GeneratedUTMLink holds a single generated URL and its parameters.
type GeneratedUTMLink struct {
	Source   string        `json:"source"`
	Medium   string        `json:"medium"`
	Campaign string        `json:"campaign"`
	Term     string        `json:"term,omitempty"`
	Content  string        `json:"content,omitempty"`
	FullURL  SafeURLString `json:"full_url"`
}

// UTMGeneratorResponse is the structure for the API response.
// It now uses utils.UTMGeneratorOptions.
type UTMGeneratorResponse struct {
	BaseURL        string                     `json:"base_url"`
	GeneratedURLs  []GeneratedUTMLink         `json:"generated_urls"`
	OptionsApplied *utils.UTMGeneratorOptions `json:"options_applied,omitempty"`
}
