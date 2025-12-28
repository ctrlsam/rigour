package persistence

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/ctrlsam/rigour/internal/geoip"
	"github.com/ctrlsam/rigour/pkg/types"
)

type Enricher struct {
	geoipClient *geoip.Client
}

func NewEnricher(geoipClient *geoip.Client) *Enricher {
	return &Enricher{
		geoipClient: geoipClient,
	}
}

func (enricher *Enricher) EnrichHost(ctx context.Context, host *types.Host) (*types.Host, error) {
	// Lookup GeoIP and ASN information
	lookupCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	geo, err := enricher.geoipClient.Lookup(lookupCtx, host.IP)
	cancel()

	if err != nil {
		return nil, err
	}

	// Populate IP integer representation
	ipInt, err := enricher.IpToIPInt(host.IP)
	if err != nil {
		return nil, err
	}
	host.IPInt = ipInt

	// Populate location info
	host.Location = &types.Location{
		Coordinates: [2]float64{geo.Longitude, geo.Latitude},
		City:        geo.City,
		Timezone:    geo.Timezone,
	}

	// Populate ASN info
	host.ASN = &types.ASNInfo{
		Number:       uint32(geo.ASN),
		Organization: geo.Organization,
		Country:      geo.Country,
	}

	// Add labels based on GeoIP flags
	if geo.IsAnonymousProxy {
		host.Labels = append(host.Labels, "anonymous_proxy")
	}

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
