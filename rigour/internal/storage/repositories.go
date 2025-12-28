package storage

import (
	"context"
	"time"

	"github.com/ctrlsam/rigour/pkg/types"
)

// RepositoryConfig holds configuration for repository initialization.
type RepositoryConfig struct {
	URI        string
	Database   string
	Collection string
	Timeout    int // in seconds
}

// FacetCounts represents aggregated counts for various facets.
type FacetCounts struct {
	Services  map[string]int `json:"services,omitempty"`
	Countries map[string]int `json:"countries,omitempty"`
	ASNs      map[string]int `json:"asns,omitempty"`
}

// HostRepository is the interface for storing and querying host records.
type HostRepository interface {
	// EnsureHost ensures a host record exists for ip.
	// Implementations should be idempotent.
	EnsureHost(ctx context.Context, ip string, now time.Time) error

	// UpsertService stores/updates a single service under its host.
	// Implementations should create the host if it doesn't exist.
	UpsertService(ctx context.Context, svc types.Service) error

	// UpdateHost updates top-level host fields (ASN/Location/Labels/etc).
	// Implementations may upsert.
	UpdateHost(ctx context.Context, host types.Host) error

	// GetByIP retrieves a single host by IP address.
	// Returns the host or an error if not found.
	GetByIP(ctx context.Context, ip string) (*types.Host, error)

	// Search queries hosts with filter and pagination support.
	// Returns hosts, next cursor ID, and error.
	Search(ctx context.Context, filter map[string]interface{}, lastID string, limit int) ([]types.Host, string, error)

	// Facets performs aggregation for facet counts.
	Facets(ctx context.Context, filter map[string]interface{}) (*FacetCounts, error)
}
