package enricher

import (
	"context"
	"time"

	"github.com/ctrlsam/rigour/internal/geoip/gogeoip"
	"github.com/ctrlsam/rigour/pkg/scanner"
)

type Enricher struct {
	geoipClient *gogeoip.Client
}

func NewEnricher(geoipClient *gogeoip.Client) *Enricher {
	return &Enricher{
		geoipClient: geoipClient,
	}
}

func (enricher *Enricher) EnrichEvent(ctx context.Context, serviceEvent *scanner.ScannedServiceEvent) (*EnrichedServiceEvent, error) {
	enriched := &EnrichedServiceEvent{
		Timestamp: serviceEvent.Timestamp,
		IP:        serviceEvent.IP,
		Port:      serviceEvent.Port,
		Protocol:  serviceEvent.Protocol,
		TLS:       serviceEvent.TLS,
		Transport: serviceEvent.Transport,
		Metadata:  serviceEvent.Metadata,
	}

	// Perform GeoIP lookup if client is available
	if enricher.geoipClient != nil {
		lookupCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		geo, err := enricher.geoipClient.Lookup(lookupCtx, serviceEvent.IP)
		cancel()
		if err == nil {
			enriched.GeoIP = geo
		}
	}

	return enriched, nil
}
