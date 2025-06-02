package utils

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

// IPInfoData struct remains the same (already has ASN fields)
type IPInfoData struct {
	IPAddress          string
	IsValid            bool
	Version            string
	IsLoopback         bool
	IsPrivate          bool
	IsMulticast        bool
	IsLinkLocalUnicast bool
	IsGlobalUnicast    bool
	ReverseDNSNames    []string
	Error              string

	CountryCode    string  `json:"country_code,omitempty"`
	CountryName    string  `json:"country_name,omitempty"`
	CityName       string  `json:"city_name,omitempty"`
	PostalCode     string  `json:"postal_code,omitempty"`
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
	TimeZone       string  `json:"time_zone,omitempty"`
	ASN            uint    `json:"asn,omitempty"`
	ASOrganization string  `json:"as_organization,omitempty"`
	GeoError       string  `json:"geo_error,omitempty"`
}

var (
	cityDB       *geoip2.Reader
	asnDB        *geoip2.Reader // Reader for the ASN database
	cityLoadOnce sync.Once
	asnLoadOnce  sync.Once
	cityLoadErr  error
	asnLoadErr   error
)

// LoadMaxMindDBs initializes the GeoIP2 readers.
// dbPaths should be a map like: {"city": "/path/to/city.mmdb", "asn": "/path/to/asn.mmdb"}
func LoadMaxMindDBs(cityDBPath string, asnDBPath string) {
	if cityDBPath != "" {
		cityLoadOnce.Do(func() {
			db, err := geoip2.Open(cityDBPath)
			if err != nil {
				log.Printf("ERROR: Could not open GeoLite2-City database at %s: %v. City GeoIP lookups will be disabled.", cityDBPath, err)
				cityLoadErr = err
				return
			}
			cityDB = db
			log.Printf("Successfully loaded GeoLite2-City database from %s", cityDBPath)
		})
	} else {
		log.Println("WARN: City MMDB path not provided. City GeoIP lookups will be disabled.")
		cityLoadErr = fmt.Errorf("city MMDB path not provided")
	}

	if asnDBPath != "" {
		asnLoadOnce.Do(func() {
			db, err := geoip2.Open(asnDBPath)
			if err != nil {
				log.Printf("ERROR: Could not open GeoLite2-ASN database at %s: %v. ASN GeoIP lookups will be disabled.", asnDBPath, err)
				asnLoadErr = err
				return
			}
			asnDB = db
			log.Printf("Successfully loaded GeoLite2-ASN database from %s", asnDBPath)
		})
	} else {
		log.Println("WARN: ASN MMDB path not provided. ASN GeoIP lookups will be disabled.")
		asnLoadErr = fmt.Errorf("ASN MMDB path not provided")
	}
}

// CloseMaxMindDBs closes all GeoIP2 readers.
func CloseMaxMindDBs() {
	if cityDB != nil {
		if err := cityDB.Close(); err != nil {
			log.Printf("Error closing GeoLite2-City database: %v", err)
		} else {
			log.Println("GeoLite2-City database closed.")
		}
	}
	if asnDB != nil {
		if err := asnDB.Close(); err != nil {
			log.Printf("Error closing GeoLite2-ASN database: %v", err)
		} else {
			log.Println("GeoLite2-ASN database closed.")
		}
	}
}

// GetBasicIPInfo retrieves basic and GeoIP information about an IP address.
func GetBasicIPInfo(ipStr string) IPInfoData {
	data := IPInfoData{IPAddress: ipStr}
	parsedIP := net.ParseIP(ipStr)

	if parsedIP == nil {
		data.IsValid = false
		data.Error = "Invalid IP address format"
		return data
	}

	data.IsValid = true
	if parsedIP.To4() != nil {
		data.Version = "IPv4"
	} else if parsedIP.To16() != nil {
		data.Version = "IPv6"
	} else {
		data.IsValid = false
		data.Error = "Unrecognized IP address format after parsing"
		return data
	}

	data.IsLoopback = parsedIP.IsLoopback()
	data.IsPrivate = parsedIP.IsPrivate()
	// ... (other basic IP checks as before) ...
	data.IsMulticast = parsedIP.IsMulticast()
	data.IsLinkLocalUnicast = parsedIP.IsLinkLocalUnicast()
	data.IsGlobalUnicast = parsedIP.IsGlobalUnicast()

	names, _ := net.LookupAddr(ipStr)
	if len(names) > 0 {
		cleanedNames := make([]string, len(names))
		for i, name := range names {
			if len(name) > 0 && name[len(name)-1] == '.' {
				cleanedNames[i] = name[:len(name)-1]
			} else {
				cleanedNames[i] = name
			}
		}
		data.ReverseDNSNames = cleanedNames
	}

	var geoErrs []string

	// City/Country/Location Lookup
	if cityDB != nil {
		cityRecord, err := cityDB.City(parsedIP) // .City() method can also be used on Country DBs
		if err == nil && cityRecord != nil {
			if cityRecord.Country.IsoCode != "" {
				data.CountryCode = cityRecord.Country.IsoCode
			}
			if name, ok := cityRecord.Country.Names["en"]; ok {
				data.CountryName = name
			}
			if name, ok := cityRecord.City.Names["en"]; ok {
				data.CityName = name
			}
			if cityRecord.Postal.Code != "" {
				data.PostalCode = cityRecord.Postal.Code
			}
			if cityRecord.Location.Latitude != 0 {
				data.Latitude = cityRecord.Location.Latitude
			}
			if cityRecord.Location.Longitude != 0 {
				data.Longitude = cityRecord.Location.Longitude
			}
			if cityRecord.Location.TimeZone != "" {
				data.TimeZone = cityRecord.Location.TimeZone
			}
			// Note: cityRecord.Traits for GeoLite2-City does NOT have ASN info directly.
		} else if err != nil {
			geoErrs = append(geoErrs, fmt.Sprintf("City/Country lookup error: %v", err))
		}
	} else if cityLoadErr != nil {
		geoErrs = append(geoErrs, fmt.Sprintf("City/Country DB not loaded: %v", cityLoadErr))
	}

	// ASN Lookup
	if asnDB != nil {
		asnRecord, err := asnDB.ASN(parsedIP) // Use the .ASN() method with the ASN database reader
		if err == nil && asnRecord != nil {
			if asnRecord.AutonomousSystemNumber != 0 {
				data.ASN = asnRecord.AutonomousSystemNumber
			}
			if asnRecord.AutonomousSystemOrganization != "" {
				data.ASOrganization = asnRecord.AutonomousSystemOrganization
			}
		} else if err != nil {
			geoErrs = append(geoErrs, fmt.Sprintf("ASN lookup error: %v", err))
		}
	} else if asnLoadErr != nil {
		geoErrs = append(geoErrs, fmt.Sprintf("ASN DB not loaded: %v", asnLoadErr))
	}

	if len(geoErrs) > 0 {
		data.GeoError = strings.Join(geoErrs, "; ")
	}

	return data
}
