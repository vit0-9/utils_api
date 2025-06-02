package utils

import (
	"fmt"
	"net"
	"strings"
)

type DNSRecord struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	Priority uint16 `json:"priority,omitempty"` // For MX records
	TTL      uint32 `json:"ttl,omitempty"`      // Often not directly available from simple lookups
}

// LookupDNSRecords performs DNS lookups for various record types.
func LookupDNSRecords(domain string, recordTypes []string) (map[string][]DNSRecord, map[string]string) {
	results := make(map[string][]DNSRecord)
	errors := make(map[string]string)

	for _, recordType := range recordTypes {
		var records []DNSRecord
		var err error

		normalizedType := strings.ToUpper(strings.TrimSpace(recordType))

		switch normalizedType {
		case "A":
			ips, e := net.LookupIP(domain)
			err = e
			for _, ip := range ips {
				if ip.To4() != nil { // Ensure it's an IPv4 address
					records = append(records, DNSRecord{Type: "A", Value: ip.String()})
				}
			}
		case "AAAA":
			ips, e := net.LookupIP(domain)
			err = e
			for _, ip := range ips {
				if ip.To16() != nil && ip.To4() == nil { // Ensure it's an IPv6 address and not an IPv4-mapped IPv6
					records = append(records, DNSRecord{Type: "AAAA", Value: ip.String()})
				}
			}
		case "MX":
			mxs, e := net.LookupMX(domain)
			err = e
			for _, mx := range mxs {
				records = append(records, DNSRecord{Type: "MX", Value: mx.Host, Priority: mx.Pref})
			}
		case "TXT":
			txts, e := net.LookupTXT(domain)
			err = e
			for _, txt := range txts {
				records = append(records, DNSRecord{Type: "TXT", Value: txt})
			}
		case "CNAME":
			cname, e := net.LookupCNAME(domain)
			err = e
			if cname != "" { // LookupCNAME returns empty string if no CNAME or multiple CNAMEs (which is invalid)
				records = append(records, DNSRecord{Type: "CNAME", Value: cname})
			}
		case "NS":
			nss, e := net.LookupNS(domain)
			err = e
			for _, ns := range nss {
				records = append(records, DNSRecord{Type: "NS", Value: ns.Host})
			}
		default:
			errors[recordType] = fmt.Sprintf("Unsupported record type: %s", recordType)
			continue
		}

		if err != nil {
			errors[recordType] = err.Error()
		}
		if len(records) > 0 {
			results[normalizedType] = records
		} else if err == nil { // No error but no records (e.g. for specific IP versions in A/AAAA)
			// results[normalizedType] = []DNSRecord{} // Optionally indicate no records found vs. error
		}

	}
	return results, errors
}
