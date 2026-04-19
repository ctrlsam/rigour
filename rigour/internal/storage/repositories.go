package storage

import (
	"context"
	"time"

	"github.com/ctrlsam/rigour/pkg/types"
)

// RepositoryConfig holds configuration for repository initialization.
type RepositoryConfig struct {
	Database   string
	Collection string
}

// CountryFacet represents a country facet with code and name.
type CountryFacet struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ASNFacet represents an ASN facet with code, name, and count.
type ASNFacet struct {
	Code  uint32 `json:"code"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// FacetCounts represents aggregated counts for various facets.
type FacetCounts struct {
	Services  map[string]int `json:"services"`
	Countries []CountryFacet `json:"countries"`
	ASNs      []ASNFacet     `json:"asns"`
}

// CampaignFilter holds optional filters for listing campaigns.
type CampaignFilter struct {
	Status types.CampaignStatus // empty = all statuses
}

// CampaignRepository is the interface for storing and querying campaign records.
type CampaignRepository interface {
	CreateCampaign(ctx context.Context, c *types.Campaign) (*types.Campaign, error)
	GetCampaign(ctx context.Context, id string) (*types.Campaign, error)
	ListCampaigns(ctx context.Context, filter CampaignFilter) ([]types.Campaign, error)
	UpdateCampaignStatus(ctx context.Context, id string, status types.CampaignStatus, errMsg string) error
	UpdateCampaignProgress(ctx context.Context, id string, progress types.CampaignProgress) error
	SetCampaignCIDRs(ctx context.Context, id string, cidrs []string) error
	DeleteCampaign(ctx context.Context, id string) error
}

// VulnerabilityRepository is the interface for storing and querying vulnerability findings.
type VulnerabilityRepository interface {
	UpsertVulnerability(ctx context.Context, v types.Vulnerability) error
	ListByIP(ctx context.Context, ip string) ([]types.Vulnerability, error)
	ListByCampaign(ctx context.Context, campaignID string, severity string) ([]types.Vulnerability, error)
	ListAll(ctx context.Context, severity string) ([]types.Vulnerability, error)
	SummariseByTemplate(ctx context.Context, severity string) ([]types.VulnerabilityTemplateSummary, error)
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
