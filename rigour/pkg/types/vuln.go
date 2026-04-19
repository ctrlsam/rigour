package types

import "time"

// VulnSummary is a compact summary embedded in host results.
type VulnSummary struct {
	Count       int      `json:"count" bson:"count"`
	MaxSeverity string   `json:"max_severity,omitempty" bson:"max_severity,omitempty"`
	CVEIDs      []string `json:"cve_ids,omitempty" bson:"cve_ids,omitempty"`
}

// VulnerabilityTemplateSummary aggregates all findings for a single Nuclei template across all hosts.
type VulnerabilityTemplateSummary struct {
	TemplateID        string   `bson:"_id"                json:"template_id"`
	Name              string   `bson:"name"               json:"name"`
	Severity          string   `bson:"severity"           json:"severity"`
	CVSS              float64  `bson:"cvss"               json:"cvss,omitempty"`
	CVEIDs            []string `bson:"cve_ids"            json:"cve_ids,omitempty"`
	Description       string   `bson:"description"        json:"description,omitempty"`
	AffectedHostCount int      `bson:"affected_host_count" json:"affected_host_count"`
}

// Vulnerability represents a security finding discovered by automated scanning.
type Vulnerability struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	IP           string    `json:"ip" bson:"ip"`
	Port         int       `json:"port" bson:"port"`
	Protocol     string    `json:"protocol" bson:"protocol"`
	TemplateID   string    `json:"template_id" bson:"template_id"`
	Name         string    `json:"name" bson:"name"`
	Severity     string    `json:"severity" bson:"severity"` // critical|high|medium|low|info
	Description  string    `json:"description,omitempty" bson:"description,omitempty"`
	CVSS         float64   `json:"cvss,omitempty" bson:"cvss,omitempty"`
	CVEIDs       []string  `json:"cve_ids,omitempty" bson:"cve_ids,omitempty"`
	MatchedAt    string    `json:"matched_at" bson:"matched_at"`
	DiscoveredAt time.Time `json:"discovered_at" bson:"discovered_at"`
	CampaignID   string    `json:"campaign_id,omitempty" bson:"campaign_id,omitempty"`
}
