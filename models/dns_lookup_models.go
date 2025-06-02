package models

import "github.com/vit0-9/utils_api/pkg/utils"

// DNSLookupRequest defines the input for a DNS lookup.
type DNSLookupRequest struct {
	Domain      string   `json:"domain" binding:"required"`
	RecordTypes []string `json:"record_types,omitempty" binding:"omitempty,dive"`
}

// DNSLookupResponse is the output of a DNS lookup.
type DNSLookupResponse struct {
	Domain  string                       `json:"domain"`
	Records map[string][]utils.DNSRecord `json:"records"`          // Keyed by record type
	Errors  map[string]string            `json:"errors,omitempty"` // Errors for specific record type lookups
}
