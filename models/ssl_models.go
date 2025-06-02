package models

import "time"

// SSLCheckRequest represents the request for SSL certificate check
type SSLCheckRequest struct {
	Domain string `json:"domain" binding:"required" example:"example.com"`
	Port   int    `json:"port,omitempty" example:"443"`
}

// SSLCheckResponse represents the response from SSL certificate check
type SSLCheckResponse struct {
	Domain             string            `json:"domain"`
	IsValid            bool              `json:"is_valid"`
	Issuer             string            `json:"issuer"`
	Subject            string            `json:"subject"`
	SerialNumber       string            `json:"serial_number"`
	NotBefore          time.Time         `json:"not_before"`
	NotAfter           time.Time         `json:"not_after"`
	DaysUntilExpiry    int               `json:"days_until_expiry"`
	SubjectAltNames    []string          `json:"subject_alt_names"`
	SignatureAlgorithm string            `json:"signature_algorithm"`
	PublicKeyAlgorithm string            `json:"public_key_algorithm"`
	KeySize            int               `json:"key_size"`
	Version            int               `json:"version"`
	IsSelfSigned       bool              `json:"is_self_signed"`
	IsWildcard         bool              `json:"is_wildcard"`
	CertificateChain   []CertificateInfo `json:"certificate_chain"`
	TLSVersion         string            `json:"tls_version"`
	CipherSuite        string            `json:"cipher_suite"`
	ValidationErrors   []string          `json:"validation_errors,omitempty"`
	QueryTime          time.Time         `json:"query_time"`
	Error              string            `json:"error,omitempty"`
}

// CertificateInfo represents information about a certificate in the chain
type CertificateInfo struct {
	Subject   string    `json:"subject"`
	Issuer    string    `json:"issuer"`
	NotBefore time.Time `json:"not_before"`
	NotAfter  time.Time `json:"not_after"`
	IsCA      bool      `json:"is_ca"`
	KeyUsage  []string  `json:"key_usage"`
}
