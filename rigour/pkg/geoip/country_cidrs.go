package geoip

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oschwald/maxminddb-golang"
)

// CountryCIDRs returns all IPv4 CIDR blocks associated with the given ISO 3166-1 alpha-2
// country code by scanning the MaxMind City or Country MMDB at geoipDataDir.
//
// The result is suitable for use as campaign CIDR targets. Only IPv4 networks are returned.
// Results are derived at call time and should be persisted to avoid repeated enumeration.
func CountryCIDRs(geoipDataDir string, countryCode string) ([]string, error) {
	if countryCode == "" {
		return nil, fmt.Errorf("country code is required")
	}
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))

	dbPath, err := findMMDB(geoipDataDir)
	if err != nil {
		return nil, err
	}

	db, err := maxminddb.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open mmdb %s: %w", dbPath, err)
	}
	defer db.Close()

	// Minimal record shape — works against GeoLite2-City, GeoLite2-Country, and GeoIP2 databases.
	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}

	var cidrs []string
	networks := db.Networks(maxminddb.SkipAliasedNetworks)
	for networks.Next() {
		subnet, err := networks.Network(&record)
		if err != nil {
			continue
		}
		if record.Country.ISOCode != countryCode {
			continue
		}
		// Skip IPv6 — naabu operates on IPv4 only
		if subnet.IP.To4() == nil {
			continue
		}
		cidrs = append(cidrs, subnet.String())
	}
	if err := networks.Err(); err != nil {
		return nil, fmt.Errorf("enumerate mmdb: %w", err)
	}
	if len(cidrs) == 0 {
		return nil, fmt.Errorf("no IPv4 CIDRs found for country %q (check country code and GeoIP database)", countryCode)
	}
	return cidrs, nil
}

// findMMDB locates the best MMDB file in dir: prefers a country-level DB, falls back to city.
func findMMDB(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read geoip dir %s: %w", dir, err)
	}

	var cityPath string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if !strings.HasSuffix(name, ".mmdb") {
			continue
		}
		if strings.Contains(name, "country") {
			return filepath.Join(dir, e.Name()), nil
		}
		if strings.Contains(name, "city") && cityPath == "" {
			cityPath = filepath.Join(dir, e.Name())
		}
	}
	if cityPath != "" {
		return cityPath, nil
	}
	return "", fmt.Errorf("no .mmdb file found in %s", dir)
}

