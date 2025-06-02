// File: models/domain.go
package models

import "time"

// WhoisLookupRequest represents the request for WHOIS lookup
type WhoisLookupRequest struct {
	Domain string `json:"domain" binding:"required" example:"example.com"`
}

// WhoisLookupResponse represents the response from WHOIS lookup
type WhoisLookupResponse struct {
	Domain          string    `json:"domain"`
	Registrar       string    `json:"registrar"`
	CreationDate    time.Time `json:"creation_date"`
	ExpirationDate  time.Time `json:"expiration_date"`
	UpdatedDate     time.Time `json:"updated_date"`
	NameServers     []string  `json:"name_servers"`
	Status          []string  `json:"status"`
	RegistrantOrg   string    `json:"registrant_org,omitempty"`
	RegistrantEmail string    `json:"registrant_email,omitempty"`
	AdminEmail      string    `json:"admin_email,omitempty"`
	TechEmail       string    `json:"tech_email,omitempty"`
	WhoisServer     string    `json:"whois_server"`
	QueryTime       time.Time `json:"query_time"`
	Error           string    `json:"error,omitempty"`
}
