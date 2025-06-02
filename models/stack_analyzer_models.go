package models

// StackAnalyzerRequest remains the same
type StackAnalyzerRequest struct {
	URL string `json:"url" binding:"required,url"`
}

// DetectedTechnology holds information about a single detected technology.
type DetectedTechnology struct {
	Name        string   `json:"name"`
	Version     string   `json:"version,omitempty"`     // Version might be part of the map key from wappalyzergo
	Categories  []string `json:"categories,omitempty"`  // Provided by AppInfo
	Description string   `json:"description,omitempty"` // Provided by AppInfo
	Website     string   `json:"website,omitempty"`     // Provided by AppInfo
	Icon        string   `json:"icon,omitempty"`        // Provided by AppInfo
	CPE         string   `json:"cpe,omitempty"`         // Provided by AppInfo
}

// StackAnalyzerResponse remains the same structure but will be populated from the new util output.
type StackAnalyzerResponse struct {
	RequestURL   string               `json:"request_url"`
	FinalURL     string               `json:"final_url"`
	Technologies []DetectedTechnology `json:"technologies"`
	Error        string               `json:"error,omitempty"`
}
