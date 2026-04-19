package types

import "time"

// TaskStatus represents the lifecycle state of a single CIDR crawl task.
type TaskStatus string

const (
	TaskStatusPending TaskStatus = "pending"
	TaskStatusRunning TaskStatus = "running"
	TaskStatusDone    TaskStatus = "done"
	TaskStatusFailed  TaskStatus = "failed"
)

// CrawlTask is the unit of work dispatched to a crawler worker.
// It carries everything a worker needs to execute one CIDR scan.
type CrawlTask struct {
	TaskID     string
	CampaignID string
	CIDR       string
	CIDRIndex  int // 0-based position in campaign.CIDRs
	Config     TaskConfig
}

// TaskConfig is a snapshot of campaign scan parameters taken at enqueue time.
// The worker needs nothing from the Campaign document itself.
type TaskConfig struct {
	ScanType      string
	Ports         string
	TopPorts      string
	Retries       int
	Rate          int
	Timeout       int // ms
	FastMode      bool
	UseUDP        bool
	Workers       int
	QueueSize     int
	ServiceFilter []string
	RunVulnScan   bool
}

// TaskResult is published by a worker after a CIDR scan completes or fails.
// The orchestrator consumes this to advance campaign progress.
type TaskResult struct {
	TaskID        string
	CampaignID    string
	CIDRIndex     int
	ServicesFound int64
	Error         string
	CompletedAt   time.Time
}
