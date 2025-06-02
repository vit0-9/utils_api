package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vit0-9/utils_api/models"           // Your models package
	"github.com/vit0-9/utils_api/pkg/utils"        // Your general utils
	"github.com/vit0-9/utils_api/pkg/utils/domain" // Your domain specific utils
)

// NetworkIntelligenceHandlers groups network and domain related utilities
type NetworkIntelligenceHandlers struct {
	// Dependencies can be injected here if needed
}

func NewNetworkIntelligenceHandlers() *NetworkIntelligenceHandlers {
	return &NetworkIntelligenceHandlers{}
}

var defaultDNSRecordTypes = []string{"A", "AAAA", "MX", "CNAME", "TXT", "NS"}

// DNSLookupHandler godoc
// @Summary      Perform DNS lookups for a domain
// @Description  Retrieves DNS records for a given domain. If 'record_types' is omitted or empty, a default set (A, AAAA, MX, CNAME, TXT, NS) will be queried.
// @Tags         Network & Domain Intelligence
// @Produce      json
// @Param        domain query string true "Domain to lookup"
// @Param        record_types query []string false "DNS record types to query (e.g., A, MX, TXT). Defaults to common set if omitted." collectionFormat(csv)
// @Success      200 {object} models.DNSLookupResponse "Successfully retrieved DNS records or errors for specific types"
// @Failure      400 {object} map[string]string "Error: Invalid input (e.g., missing domain)"
// @Router       /net/dns-lookup [get]
func (h *NetworkIntelligenceHandlers) DNSLookupHandler(c *gin.Context) {
	domainQuery := c.Query("domain")
	if domainQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain query parameter is required"})
		return
	}

	recordTypesQuery := c.QueryArray("record_types") // For GET, query params are often like ?record_types=A&record_types=MX
	// Or if using a single comma-separated string:
	// recordTypesStr := c.Query("record_types")
	// if recordTypesStr != "" { recordTypesQuery = strings.Split(recordTypesStr, ",") }

	typesToLookup := recordTypesQuery
	if len(typesToLookup) == 0 {
		// Check if the parameter was provided but empty vs not provided at all
		// For simplicity, if query array is empty, use default.
		// More robust: check c.Request.URL.Query().Has("record_types")
		if _, ok := c.Request.URL.Query()["record_types"]; !ok {
			typesToLookup = defaultDNSRecordTypes
		}
	}

	for i, rt := range typesToLookup {
		typesToLookup[i] = strings.ToUpper(strings.TrimSpace(rt))
	}

	utilRecords, lookupErrors := utils.LookupDNSRecords(domainQuery, typesToLookup) // Assuming this is in general utils now

	responseRecords := make(map[string][]utils.DNSRecord)
	for recordType, localRecs := range utilRecords {
		modelRecs := make([]utils.DNSRecord, len(localRecs))
		for i, lr := range localRecs {
			modelRecs[i] = utils.DNSRecord{
				Type:     lr.Type,
				Value:    lr.Value,
				Priority: lr.Priority,
			}
		}
		responseRecords[recordType] = modelRecs
	}

	response := models.DNSLookupResponse{
		Domain:  domainQuery,
		Records: responseRecords,
		Errors:  lookupErrors,
	}
	c.JSON(http.StatusOK, response) // Changed from PureJSON if SafeURLString handles escaping
}

// IPInfoHandler godoc
// @Summary      Get detailed information about an IP address
// @Description  Provides validation, type classification, reverse DNS, and GeoIP/ASN information for an IP.
// @Tags         Network & Domain Intelligence
// @Produce      json
// @Param        ip query string true "IP Address to get info for"
// @Success      200 {object} models.IPInfoResponse "Successfully retrieved IP information"
// @Failure      400 {object} map[string]string "Error: Invalid input (e.g., missing IP address)"
// @Router       /net/ip-info [get]
func (h *NetworkIntelligenceHandlers) IPInfoHandler(c *gin.Context) {
	ipAddress := c.Query("ip")
	if ipAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip query parameter is required"})
		return
	}

	utilData := utils.GetBasicIPInfo(ipAddress) // Assuming this is in general utils now

	response := models.IPInfoResponse{
		IPAddress:          utilData.IPAddress,
		IsValid:            utilData.IsValid,
		Version:            utilData.Version,
		IsLoopback:         utilData.IsLoopback,
		IsPrivate:          utilData.IsPrivate,
		IsMulticast:        utilData.IsMulticast,
		IsLinkLocalUnicast: utilData.IsLinkLocalUnicast,
		IsGlobalUnicast:    utilData.IsGlobalUnicast,
		ReverseDNSNames:    utilData.ReverseDNSNames,
		Error:              utilData.Error,
		CountryCode:        utilData.CountryCode,
		CountryName:        utilData.CountryName,
		CityName:           utilData.CityName,
		PostalCode:         utilData.PostalCode,
		Latitude:           utilData.Latitude,
		Longitude:          utilData.Longitude,
		TimeZone:           utilData.TimeZone,
		ASN:                utilData.ASN,
		ASOrganization:     utilData.ASOrganization,
		GeoError:           utilData.GeoError,
	}
	c.JSON(http.StatusOK, response)
}

// WhoisLookupHandler godoc
// @Summary      Perform WHOIS lookup for a domain
// @Description  Retrieves WHOIS information for a given domain.
// @Tags         Network & Domain Intelligence
// @Produce      json
// @Param        domain query string true "Domain for WHOIS lookup"
// @Success      200 {object} models.WhoisLookupResponse "Successfully retrieved WHOIS information or error during lookup"
// @Failure      400 {object} map[string]string "Error: Invalid input (e.g., missing domain)"
// @Router       /net/whois-lookup [get]
func (h *NetworkIntelligenceHandlers) WhoisLookupHandler(c *gin.Context) {
	domainQuery := c.Query("domain")
	if domainQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain query parameter is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	whoisInfo, err := domain.GetWhoisInfo(ctx, domainQuery) // domain.GetWhoisInfo
	if err != nil {
		c.JSON(http.StatusOK, models.WhoisLookupResponse{ // Still 200 but with error in body
			Domain:    domainQuery,
			QueryTime: time.Now(),
			Error:     err.Error(),
		})
		return
	}
	response := models.WhoisLookupResponse{
		Domain:          whoisInfo.Domain,
		Registrar:       whoisInfo.Registrar,
		CreationDate:    whoisInfo.CreationDate,
		ExpirationDate:  whoisInfo.ExpirationDate,
		UpdatedDate:     whoisInfo.UpdatedDate,
		NameServers:     whoisInfo.NameServers,
		Status:          whoisInfo.Status,
		RegistrantOrg:   whoisInfo.RegistrantOrg,
		RegistrantEmail: whoisInfo.RegistrantEmail,
		AdminEmail:      whoisInfo.AdminEmail,
		TechEmail:       whoisInfo.TechEmail,
		WhoisServer:     whoisInfo.WhoisServer,
		QueryTime:       whoisInfo.QueryTime,
	}
	c.JSON(http.StatusOK, response)
}

// SSLCheckHandler godoc
// @Summary      Check SSL certificate information for a domain/host
// @Description  Retrieves SSL certificate details for a given host and optional port (defaults to 443).
// @Tags         Network & Domain Intelligence
// @Produce      json
// @Param        host query string true "Host (domain or IP) for SSL check"
// @Param        port query int false "Port for SSL check (defaults to 443)"
// @Success      200 {object} models.SSLCheckResponse "Successfully retrieved SSL certificate information or error during check"
// @Failure      400 {object} map[string]string "Error: Invalid input (e.g., missing host)"
// @Router       /net/ssl-check [get]
func (h *NetworkIntelligenceHandlers) SSLCheckHandler(c *gin.Context) {
	hostQuery := c.Query("host")
	if hostQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "host query parameter is required"})
		return
	}

	portQueryStr := c.Query("port")
	port := 0 // Default will be handled by GetSSLInfo (usually 443)
	var err error
	if portQueryStr != "" {
		port, err = strconv.Atoi(portQueryStr)
		if err != nil || port <= 0 || port > 65535 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid port number"})
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second) // Adjusted timeout
	defer cancel()

	var sslInfo *domain.SSLInfo // Assuming domain.SSLInfo is the struct from your util

	if port > 0 {
		sslInfo, err = domain.GetSSLInfo(ctx, hostQuery, port)
	} else {
		sslInfo, err = domain.GetSSLInfo(ctx, hostQuery) // Util defaults port to 443
	}

	if err != nil {
		c.JSON(http.StatusOK, models.SSLCheckResponse{ // Still 200 but with error in body
			Domain:    hostQuery, // Use hostQuery as Domain for response consistency
			QueryTime: time.Now(),
			Error:     err.Error(),
		})
		return
	}

	certificateChain := make([]models.CertificateInfo, len(sslInfo.CertificateChain))
	for i, cert := range sslInfo.CertificateChain {
		certificateChain[i] = models.CertificateInfo{
			Subject:   cert.Subject,
			Issuer:    cert.Issuer,
			NotBefore: cert.NotBefore,
			NotAfter:  cert.NotAfter,
			IsCA:      cert.IsCA,
			KeyUsage:  cert.KeyUsage,
		}
	}

	response := models.SSLCheckResponse{
		Domain:             sslInfo.Domain,
		IsValid:            sslInfo.IsValid,
		Issuer:             sslInfo.Issuer,
		Subject:            sslInfo.Subject,
		SerialNumber:       sslInfo.SerialNumber,
		NotBefore:          sslInfo.NotBefore,
		NotAfter:           sslInfo.NotAfter,
		DaysUntilExpiry:    sslInfo.DaysUntilExpiry,
		SubjectAltNames:    sslInfo.SubjectAltNames,
		SignatureAlgorithm: sslInfo.SignatureAlgorithm,
		PublicKeyAlgorithm: sslInfo.PublicKeyAlgorithm,
		KeySize:            sslInfo.KeySize,
		Version:            sslInfo.Version,
		IsSelfSigned:       sslInfo.IsSelfSigned,
		IsWildcard:         sslInfo.IsWildcard,
		CertificateChain:   certificateChain,
		TLSVersion:         sslInfo.TLSVersion,
		CipherSuite:        sslInfo.CipherSuite,
		ValidationErrors:   sslInfo.ValidationErrors,
		QueryTime:          sslInfo.QueryTime,
	}
	c.JSON(http.StatusOK, response)
}
