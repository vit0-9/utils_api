
package domain

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

type WhoisInfo struct {
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
	RawData         string    `json:"raw_data,omitempty"`
	WhoisServer     string    `json:"whois_server"`
	QueryTime       time.Time `json:"query_time"`
}

type WhoisError struct {
	Domain string
	Err    error
	Server string
}

func (e *WhoisError) Error() string {
	return fmt.Sprintf("whois lookup failed for %s via %s: %v", e.Domain, e.Server, e.Err)
}

// WhoisServers defines fallback servers for different TLDs
var WhoisServers = map[string][]string{
	"com":     {"whois.verisign-grs.com", "whois.markmonitor.com"},
	"net":     {"whois.verisign-grs.com"},
	"org":     {"whois.pir.org"},
	"info":    {"whois.afilias.net"},
	"biz":     {"whois.neulevel.biz"},
	"default": {"whois.iana.org", "whois.internic.net"},
}

// GetWhoisInfo performs WHOIS lookup with fallback servers
func GetWhoisInfo(ctx context.Context, domain string) (*WhoisInfo, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	// Extract TLD for server selection
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid domain format: %s", domain)
	}
	tld := parts[len(parts)-1]

	// Get servers for this TLD
	servers := WhoisServers[tld]
	if len(servers) == 0 {
		servers = WhoisServers["default"]
	}

	var lastErr error
	for _, server := range servers {
		result, err := queryWhoisServer(ctx, domain, server)
		if err != nil {
			lastErr = &WhoisError{Domain: domain, Err: err, Server: server}
			continue
		}
		return result, nil
	}

	return nil, lastErr
}

// queryWhoisServer performs the actual WHOIS query
func queryWhoisServer(ctx context.Context, domain, server string) (*WhoisInfo, error) {
	conn, err := net.DialTimeout("tcp", server+":43", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	// Set deadline for the entire operation
	conn.SetDeadline(time.Now().Add(15 * time.Second))

	// Send query
	query := domain + "\r\n"
	if _, err := conn.Write([]byte(query)); err != nil {
		return nil, fmt.Errorf("write failed: %w", err)
	}

	// Read response
	var response strings.Builder
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		response.WriteString(scanner.Text() + "\n")
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	rawData := response.String()
	if rawData == "" {
		return nil, fmt.Errorf("empty response from server")
	}

	// Parse the response
	whoisInfo := parseWhoisResponse(domain, rawData, server)
	whoisInfo.QueryTime = time.Now()

	return whoisInfo, nil
}

// parseWhoisResponse extracts structured data from raw WHOIS response
func parseWhoisResponse(domain, rawData, server string) *WhoisInfo {
	info := &WhoisInfo{
		Domain:      domain,
		RawData:     rawData,
		WhoisServer: server,
	}

	lines := strings.Split(rawData, "\n")

	// Common patterns for different WHOIS formats
	patterns := map[string]*regexp.Regexp{
		"registrar":        regexp.MustCompile(`(?i)registrar:\s*(.+)`),
		"creation_date":    regexp.MustCompile(`(?i)(creation date|created|registered):\s*(.+)`),
		"expiration_date":  regexp.MustCompile(`(?i)(expir|expires).*:\s*(.+)`),
		"updated_date":     regexp.MustCompile(`(?i)(updated|last updated|modified).*:\s*(.+)`),
		"name_server":      regexp.MustCompile(`(?i)name server:\s*(.+)`),
		"status":           regexp.MustCompile(`(?i)(domain )?status:\s*(.+)`),
		"registrant_org":   regexp.MustCompile(`(?i)registrant.*organization:\s*(.+)`),
		"registrant_email": regexp.MustCompile(`(?i)registrant.*email:\s*(.+)`),
		"admin_email":      regexp.MustCompile(`(?i)admin.*email:\s*(.+)`),
		"tech_email":       regexp.MustCompile(`(?i)tech.*email:\s*(.+)`),
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse different fields
		if match := patterns["registrar"].FindStringSubmatch(line); len(match) > 1 {
			info.Registrar = strings.TrimSpace(match[1])
		}

		if match := patterns["creation_date"].FindStringSubmatch(line); len(match) > 2 {
			if date := parseDate(match[2]); !date.IsZero() {
				info.CreationDate = date
			}
		}

		if match := patterns["expiration_date"].FindStringSubmatch(line); len(match) > 2 {
			if date := parseDate(match[2]); !date.IsZero() {
				info.ExpirationDate = date
			}
		}

		if match := patterns["updated_date"].FindStringSubmatch(line); len(match) > 2 {
			if date := parseDate(match[2]); !date.IsZero() {
				info.UpdatedDate = date
			}
		}

		if match := patterns["name_server"].FindStringSubmatch(line); len(match) > 1 {
			ns := strings.ToLower(strings.TrimSpace(match[1]))
			info.NameServers = append(info.NameServers, ns)
		}

		if match := patterns["status"].FindStringSubmatch(line); len(match) > 2 {
			status := strings.TrimSpace(match[2])
			info.Status = append(info.Status, status)
		}

		if match := patterns["registrant_org"].FindStringSubmatch(line); len(match) > 1 {
			info.RegistrantOrg = strings.TrimSpace(match[1])
		}

		if match := patterns["registrant_email"].FindStringSubmatch(line); len(match) > 1 {
			info.RegistrantEmail = strings.TrimSpace(match[1])
		}

		if match := patterns["admin_email"].FindStringSubmatch(line); len(match) > 1 {
			info.AdminEmail = strings.TrimSpace(match[1])
		}

		if match := patterns["tech_email"].FindStringSubmatch(line); len(match) > 1 {
			info.TechEmail = strings.TrimSpace(match[1])
		}
	}

	// Remove duplicates from slices
	info.NameServers = removeDuplicates(info.NameServers)
	info.Status = removeDuplicates(info.Status)

	return info
}

// parseDate attempts to parse various date formats found in WHOIS data
func parseDate(dateStr string) time.Time {
	dateStr = strings.TrimSpace(dateStr)

	// Common WHOIS date formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00", // RFC3339
		"2006-01-02T15:04:05Z",      // RFC3339 UTC
		"2006-01-02 15:04:05",       // MySQL datetime
		"2006-01-02",                // Date only
		"02-Jan-2006",               // Some registrars
		"January 02 2006",           // Some registrars
		"2-Jan-2006",                // Some registrars
		"2006/01/02",                // Some registrars
	}

	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date
		}
	}

	return time.Time{}
}

// removeDuplicates removes duplicate strings from slice
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}