package models

// HTTPHeadersRequest defines the input for fetching HTTP headers.
type HTTPHeadersRequest struct {
	URL    string `json:"url" binding:"required,url"`
	Method string `json:"method,omitempty"`
}

// HTTPHeadersResponse is the output for HTTP headers.
type HTTPHeadersResponse struct {
	RequestURL string              `json:"request_url"`
	FinalURL   string              `json:"final_url,omitempty"`
	StatusCode int                 `json:"status_code,omitempty"`
	Status     string              `json:"status,omitempty"`
	Headers    map[string][]string `json:"headers,omitempty"`
	Error      string              `json:"error,omitempty"`
}
