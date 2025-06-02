package models

// IPInfoRequest defines the input for IP information.
type IPInfoRequest struct {
	IPAddress string `json:"ip_address" binding:"required"`
}

// IPInfoResponse is the output for IP information.
type IPInfoResponse struct {
	IPAddress          string   `json:"ip_address"`
	IsValid            bool     `json:"is_valid"`
	Version            string   `json:"version,omitempty"`
	IsLoopback         bool     `json:"is_loopback"`
	IsPrivate          bool     `json:"is_private"`
	IsMulticast        bool     `json:"is_multicast"`
	IsLinkLocalUnicast bool     `json:"is_link_local_unicast"`
	IsGlobalUnicast    bool     `json:"is_global_unicast"`
	ReverseDNSNames    []string `json:"reverse_dns_names,omitempty"`
	Error              string   `json:"error,omitempty"`

	// GeoIP Information
	CountryCode    string  `json:"country_code,omitempty"`
	CountryName    string  `json:"country_name,omitempty"`
	CityName       string  `json:"city_name,omitempty"`
	PostalCode     string  `json:"postal_code,omitempty"`
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
	TimeZone       string  `json:"time_zone,omitempty"`
	ASN            uint    `json:"asn,omitempty"`
	ASOrganization string  `json:"as_organization,omitempty"`
	GeoError       string  `json:"geo_error,omitempty"` // Errors specific to GeoIP lookup
}
