package api

import (
	"encoding/json"
	"time"

	"github.com/ctrlsam/rigour/internal/geoip"
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

// SearchRequest represents a search query request.
type SearchRequest struct {
	Filter    map[string]interface{} `json:"filter,omitempty"`
	PageToken string                 `json:"page_token,omitempty"`
	Limit     int                    `json:"limit,omitempty"`
}

// SearchResponse represents the response for a search query.
type SearchResponse struct {
	Hosts         []Host       `json:"hosts"`
	Facets        *FacetCounts `json:"facets,omitempty"`
	NextPageToken string       `json:"next_page_token,omitempty"`
}

// FacetCounts represents aggregated counts for various facets.
type FacetCounts struct {
	Services  map[string]int `json:"services,omitempty"`
	Countries map[string]int `json:"countries,omitempty"`
	ASNs      map[string]int `json:"asns,omitempty"`
}

// FacetRequest represents a request for facet aggregation.
type FacetRequest struct {
	Filter map[string]interface{} `json:"filter,omitempty"`
}

// FacetResponse represents the response for a facet aggregation query.
type FacetResponse struct {
	Facets FacetCounts `json:"facets"`
}

// PaginationToken represents the pagination cursor token structure.
type PaginationToken struct {
	LastID    string `json:"last_id"`
	SortField string `json:"sort_field"`
}
