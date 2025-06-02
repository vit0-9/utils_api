package models

// ResolveRedirectRequest defines the expected JSON input
type ResolveRedirectRequest struct {
	URL string `json:"url" binding:"required,url"`
}

// ResolveRedirectResponse defines the JSON output
type ResolveRedirectResponse struct {
	OriginalURL SafeURLString `json:"original_url"`
	FinalURL    SafeURLString `json:"final_url,omitempty"`
	Error       string        `json:"error,omitempty"`
}
