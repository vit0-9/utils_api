// File: internal/utils/domain/ssl.go
package domain

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"time"
)

type SSLInfo struct {
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
}

type CertificateInfo struct {
	Subject   string    `json:"subject"`
	Issuer    string    `json:"issuer"`
	NotBefore time.Time `json:"not_before"`
	NotAfter  time.Time `json:"not_after"`
	IsCA      bool      `json:"is_ca"`
	KeyUsage  []string  `json:"key_usage"`
}

type SSLError struct {
	Domain string
	Err    error
}

func (e *SSLError) Error() string {
	return fmt.Sprintf("SSL check failed for %s: %v", e.Domain, e.Err)
}

// GetSSLInfo retrieves SSL certificate information for a domain
func GetSSLInfo(ctx context.Context, domain string, port ...int) (*SSLInfo, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	// Default to HTTPS port
	targetPort := 443
	if len(port) > 0 && port[0] > 0 {
		targetPort = port[0]
	}

	address := fmt.Sprintf("%s:%d", domain, targetPort)

	// Create TLS connection with timeout
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", address, &tls.Config{
		ServerName:         domain,
		InsecureSkipVerify: true, // We want to analyze even invalid certs
	})
	if err != nil {
		return nil, &SSLError{Domain: domain, Err: err}
	}
	defer conn.Close()

	// Get connection state
	state := conn.ConnectionState()

	if len(state.PeerCertificates) == 0 {
		return nil, &SSLError{Domain: domain, Err: fmt.Errorf("no certificates found")}
	}

	cert := state.PeerCertificates[0]

	// Build SSL info
	sslInfo := &SSLInfo{
		Domain:             domain,
		Issuer:             cert.Issuer.String(),
		Subject:            cert.Subject.String(),
		SerialNumber:       cert.SerialNumber.String(),
		NotBefore:          cert.NotBefore,
		NotAfter:           cert.NotAfter,
		SubjectAltNames:    cert.DNSNames,
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
		Version:            cert.Version,
		TLSVersion:         getTLSVersion(state.Version),
		CipherSuite:        tls.CipherSuiteName(state.CipherSuite),
		QueryTime:          time.Now(),
	}

	// Calculate days until expiry
	daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)
	sslInfo.DaysUntilExpiry = daysUntilExpiry

	// Check if certificate is valid
	sslInfo.IsValid = daysUntilExpiry > 0 && time.Now().After(cert.NotBefore)

	// Determine key size
	sslInfo.KeySize = getKeySize(cert)

	// Check if self-signed
	sslInfo.IsSelfSigned = cert.Issuer.String() == cert.Subject.String()

	// Check if wildcard
	for _, name := range cert.DNSNames {
		if strings.HasPrefix(name, "*.") {
			sslInfo.IsWildcard = true
			break
		}
	}

	// Validate certificate chain
	sslInfo.ValidationErrors = validateCertificate(cert, domain)

	// Process certificate chain
	for _, peerCert := range state.PeerCertificates {
		certInfo := CertificateInfo{
			Subject:   peerCert.Subject.String(),
			Issuer:    peerCert.Issuer.String(),
			NotBefore: peerCert.NotBefore,
			NotAfter:  peerCert.NotAfter,
			IsCA:      peerCert.IsCA,
			KeyUsage:  getKeyUsage(peerCert),
		}
		sslInfo.CertificateChain = append(sslInfo.CertificateChain, certInfo)
	}

	return sslInfo, nil
}

// validateCertificate performs basic certificate validation
func validateCertificate(cert *x509.Certificate, domain string) []string {
	var errors []string

	// Check expiry
	if time.Now().After(cert.NotAfter) {
		errors = append(errors, "certificate has expired")
	}

	// Check not yet valid
	if time.Now().Before(cert.NotBefore) {
		errors = append(errors, "certificate is not yet valid")
	}

	// Check domain match
	if !matchesDomain(cert, domain) {
		errors = append(errors, "certificate does not match domain")
	}

	// Check if certificate is revoked (basic check)
	if len(cert.CRLDistributionPoints) == 0 && len(cert.OCSPServer) == 0 {
		errors = append(errors, "no revocation checking mechanism available")
	}

	return errors
}

// matchesDomain checks if certificate matches the domain
func matchesDomain(cert *x509.Certificate, domain string) bool {
	// Check subject common name
	if strings.EqualFold(cert.Subject.CommonName, domain) {
		return true
	}

	// Check subject alternative names
	for _, name := range cert.DNSNames {
		if strings.EqualFold(name, domain) {
			return true
		}
		// Check wildcard match
		if strings.HasPrefix(name, "*.") {
			wildcard := name[2:]
			if strings.HasSuffix(domain, "."+wildcard) || strings.EqualFold(domain, wildcard) {
				return true
			}
		}
	}

	return false
}

// getKeySize determines the key size based on public key type
func getKeySize(cert *x509.Certificate) int {
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return pub.N.BitLen()
	case *ecdsa.PublicKey:
		return pub.Curve.Params().BitSize
	case *ed25519.PublicKey:
		return 256 // Ed25519 is equivalent to 256-bit
	default:
		return 0
	}
}

// getTLSVersion converts TLS version constant to string
func getTLSVersion(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (%d)", version)
	}
}

// getKeyUsage extracts key usage information
func getKeyUsage(cert *x509.Certificate) []string {
	var usage []string

	if cert.KeyUsage&x509.KeyUsageDigitalSignature != 0 {
		usage = append(usage, "Digital Signature")
	}
	if cert.KeyUsage&x509.KeyUsageContentCommitment != 0 {
		usage = append(usage, "Content Commitment")
	}
	if cert.KeyUsage&x509.KeyUsageKeyEncipherment != 0 {
		usage = append(usage, "Key Encipherment")
	}
	if cert.KeyUsage&x509.KeyUsageDataEncipherment != 0 {
		usage = append(usage, "Data Encipherment")
	}
	if cert.KeyUsage&x509.KeyUsageKeyAgreement != 0 {
		usage = append(usage, "Key Agreement")
	}
	if cert.KeyUsage&x509.KeyUsageCertSign != 0 {
		usage = append(usage, "Certificate Signing")
	}
	if cert.KeyUsage&x509.KeyUsageCRLSign != 0 {
		usage = append(usage, "CRL Signing")
	}

	return usage
}
