package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ctrlsam/rigour/internal/geoip"
	"github.com/ctrlsam/rigour/pkg/enricher"
)

// Service represents a single service discovered on a host.
type Service struct {
	Port      int             `json:"port"`
	Protocol  string          `json:"protocol"`
	TLS       bool            `json:"tls"`
	Transport string          `json:"transport"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// Host represents a scanned host with its associated services and metadata.
type Host struct {
	ID        string             `json:"id"`
	IP        string             `json:"ip"`
	Country   string             `json:"country,omitempty"`
	ASN       string             `json:"asn,omitempty"`
	Services  []Service          `json:"services,omitempty"`
	GeoIP     *geoip.GeoIPRecord `json:"geoip,omitempty"`
	Timestamp time.Time          `json:"timestamp"`
}

// FacetCounts represents aggregated counts for various facets.
type FacetCounts struct {
	Services  map[string]int `json:"services,omitempty"`
	Countries map[string]int `json:"countries,omitempty"`
	ASNs      map[string]int `json:"asns,omitempty"`
}

// ServiceRepository is the interface for storing and querying service records.
type ServiceRepository interface {
	// InsertServiceRecord inserts or updates a service record for a host.
	InsertServiceRecord(ctx context.Context, record enricher.EnrichedServiceEvent) error

	// Search queries hosts with filter and pagination support.
	// Returns hosts, next cursor ID, and error.
	Search(ctx context.Context, filter map[string]interface{}, lastID string, limit int) ([]Host, string, error)

	// Facets performs aggregation for facet counts.
	Facets(ctx context.Context, filter map[string]interface{}) (*FacetCounts, error)
}

// RepositoryConfig holds configuration for repository initialization.
type RepositoryConfig struct {
	URI        string
	Database   string
	Collection string
	Timeout    int // in seconds
}
