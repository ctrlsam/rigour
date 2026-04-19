package persistence

import (
	"context"
	"fmt"
	"net"

	"github.com/ctrlsam/rigour/pkg/types"
)

// HostEnricher enriches a host with additional metadata.
type HostEnricher interface {
	EnrichHost(ctx context.Context, host *types.Host) (*types.Host, error)
	Close()
}

type Enricher struct {
	geoIPReaders *GeoIPReaders
}

func NewEnricher(geoIPReaders *GeoIPReaders) *Enricher {
	return &Enricher{
		geoIPReaders: geoIPReaders,
	}
}

func (enricher *Enricher) EnrichHost(ctx context.Context, host *types.Host) (*types.Host, error) {
	ip := net.ParseIP(host.IP)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", host.IP)
	}

	// Lookup ASN
	if enricher.geoIPReaders.ASN != nil {
		if record, err := enricher.geoIPReaders.ASN.ASN(ip); err == nil {
			if host.ASN == nil {
				host.ASN = &types.ASNInfo{}
			}
			host.ASN.Number = uint32(record.AutonomousSystemNumber)
			host.ASN.Organization = record.AutonomousSystemOrganization
		}
	}

	// Lookup GeoIP City
	if enricher.geoIPReaders.City != nil {
		if record, err := enricher.geoIPReaders.City.City(ip); err == nil {
			host.Location = &types.Location{
				Coordinates: [2]float64{record.Location.Longitude, record.Location.Latitude},
				City:        record.City.Names["en"],
				Timezone:    record.Location.TimeZone,
				CountryCode: record.Country.IsoCode,
				CountryName: record.Country.Names["en"],
			}

			// Add Satellite Provider label
			host.ASN.IsSatelliteProvider = record.Traits.IsSatelliteProvider

			// Add labels
			if record.Traits.IsAnonymousProxy {
				host.Labels = append(host.Labels, "anonymous-proxy")
			}

		}
	}

	// Populate IP integer representation
	ipInt, err := enricher.IpToIPInt(host.IP)
	if err != nil {
		return nil, err
	}
	host.IPInt = ipInt

	return host, nil
}

func (enricher *Enricher) IpToIPInt(ip string) (uint64, error) {
	// Parse the IP address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return 0, fmt.Errorf("invalid IP address: %s", ip)
	}

	// Convert to IPv4
	ipv4 := parsedIP.To4()
	if ipv4 == nil {
		return 0, fmt.Errorf("only IPv4 addresses are supported: %s", ip)
	}

	// Convert 4 bytes to uint64
	// Each byte is shifted to its appropriate position
	return uint64(ipv4[0])<<24 | uint64(ipv4[1])<<16 | uint64(ipv4[2])<<8 | uint64(ipv4[3]), nil
}

// Close closes the GeoIP database readers.
func (enricher *Enricher) Close() {
	if enricher.geoIPReaders != nil {
		enricher.geoIPReaders.Close()
	}
}
