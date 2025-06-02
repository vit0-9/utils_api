// models/url_cleaner_models.go
package models

import "github.com/vit0-9/utils_api/pkg/utils"

// CleanURLRequest defines the expected JSON input for the clean URL endpoint.
// It contains the URL that needs to be cleaned.
type CleanURLRequest struct {
	URL string `json:"url" binding:"required,url" example:"https://example.com?utm_source=google"` // Add example
}

// DetailedCleanURLResponse defines the JSON output with details of removed params
type DetailedCleanURLResponse struct {
	OriginalURL   SafeURLString            `json:"original_url" example:"https://example.com?utm_source=google"`
	CleanedURL    SafeURLString            `json:"cleaned_url" example:"https://example.com/"`
	RemovedParams []utils.RemovedParamInfo `json:"removed_params,omitempty"`
	Message       string                   `json:"message,omitempty" example:"Tracking parameters removed."`
}
