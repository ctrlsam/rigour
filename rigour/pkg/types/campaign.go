package types

import "time"

type CampaignStatus string

const (
	CampaignPending   CampaignStatus = "pending"
	CampaignRunning   CampaignStatus = "running"
	CampaignPaused    CampaignStatus = "paused"
	CampaignCompleted CampaignStatus = "completed"
	CampaignFailed    CampaignStatus = "failed"
	CampaignCancelled CampaignStatus = "cancelled"
)

// ValidCampaignStatuses is the set of accepted status filter values.
var ValidCampaignStatuses = map[CampaignStatus]struct{}{
	CampaignPending:   {},
	CampaignRunning:   {},
	CampaignPaused:    {},
	CampaignCompleted: {},
	CampaignFailed:    {},
	CampaignCancelled: {},
}

const (
	ScanTypeConnect = "c"
	ScanTypeSyn     = "s"
	DefaultTopPorts = "1000"
)

type Campaign struct {
	ID          string         `bson:"_id,omitempty" json:"id"`
	Name        string         `bson:"name" json:"name"`
	Description string         `bson:"description,omitempty" json:"description,omitempty"`
	Status      CampaignStatus `bson:"status" json:"status"`

	// Targeting: either CIDRs or CountryCode (resolved to CIDRs at start time)
	CIDRs       []string `bson:"cidrs" json:"cidrs"`
	CountryCode string   `bson:"country_code,omitempty" json:"country_code,omitempty"`

	// Port config — mutually exclusive: Ports overrides TopPorts
	Ports    string `bson:"ports,omitempty" json:"ports,omitempty"`         // e.g. "80,443,8080-8090"
	TopPorts string `bson:"top_ports,omitempty" json:"top_ports,omitempty"` // "100"|"1000"|"full"

	// Service filter: empty means all plugins run
	ServiceFilter []string `bson:"service_filter,omitempty" json:"service_filter,omitempty"`

	// Scan tuning
	ScanType        string `bson:"scan_type" json:"scan_type"`                                   // "c"|"s"
	Rate            int    `bson:"rate" json:"rate"`                                             // packets/sec; 0 = default (500)
	Retries         int    `bson:"retries" json:"retries"`
	Timeout         int    `bson:"timeout" json:"timeout"`                                       // ms
	FastMode        bool   `bson:"fast_mode" json:"fast_mode"`
	UseUDP          bool   `bson:"use_udp" json:"use_udp"`
	Workers         int    `bson:"workers,omitempty" json:"workers,omitempty"`                   // fingerprint goroutines; 0 = default (50)
	QueueSize       int    `bson:"queue_size,omitempty" json:"queue_size,omitempty"`             // fingerprint queue depth; 0 = default (500)
	ConcurrentCIDRs int    `bson:"concurrent_cidrs,omitempty" json:"concurrent_cidrs,omitempty"` // parallel CIDRs; 0 = default (3)

	// Vulnerability scanning: opt-in per campaign
	RunVulnScan       bool     `bson:"run_vuln_scan,omitempty" json:"run_vuln_scan,omitempty"`
	VulnScanTemplates []string `bson:"vuln_scan_templates,omitempty" json:"vuln_scan_templates,omitempty"`

	// Progress: updated atomically after each CIDR completes
	Progress CampaignProgress `bson:"progress" json:"progress"`

	CreatedAt   time.Time  `bson:"created_at" json:"created_at"`
	StartedAt   *time.Time `bson:"started_at,omitempty" json:"started_at,omitempty"`
	CompletedAt *time.Time `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	UpdatedAt   time.Time  `bson:"updated_at" json:"updated_at"`
	Error       string     `bson:"error,omitempty" json:"error,omitempty"`
}

type CampaignProgress struct {
	TotalCIDRs      int    `bson:"total_cidrs" json:"total_cidrs"`
	CompletedCIDRs  int    `bson:"completed_cidrs" json:"completed_cidrs"`
	CurrentCIDR     string `bson:"current_cidr,omitempty" json:"current_cidr,omitempty"`
	HostsDiscovered int64  `bson:"hosts_discovered" json:"hosts_discovered"`
	ServicesFound   int64  `bson:"services_found" json:"services_found"`
}
