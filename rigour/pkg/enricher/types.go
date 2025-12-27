package enricher

import (
	"encoding/json"
	"time"

	"github.com/ctrlsam/rigour/internal/geoip"
)

type EnrichedServiceEvent struct {
	Timestamp time.Time          `json:"timestamp"`
	IP        string             `json:"ip"`
	Port      int                `json:"port"`
	Protocol  string             `json:"protocol"`
	TLS       bool               `json:"tls"`
	Transport string             `json:"transport"`
	Metadata  json.RawMessage    `json:"metadata,omitempty"`
	GeoIP     *geoip.GeoIPRecord `json:"geoip,omitempty"`
}
