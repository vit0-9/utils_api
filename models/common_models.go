// models/common_models.go
package models

// APIErrorResponse represents a standard error response format.
type APIErrorResponse struct {
	StatusCode int    `json:"status_code"`       // HTTP status code
	ErrorCode  string `json:"error_code"`        // Application-specific error code (optional)
	Message    string `json:"message"`           // User-friendly error message
	Details    string `json:"details,omitempty"` // More detailed information, if available
}
