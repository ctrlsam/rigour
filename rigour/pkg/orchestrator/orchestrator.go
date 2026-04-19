package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ctrlsam/rigour/internal/messaging"
	"github.com/ctrlsam/rigour/internal/metrics"
	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/ctrlsam/rigour/pkg/geoip"
	"github.com/ctrlsam/rigour/pkg/types"
)

// campaignState tracks in-flight progress for a single campaign.
type campaignState struct {
	mu           sync.Mutex
	progress     types.CampaignProgress
	pendingCount int
	doneCh       chan struct{}
}

// Orchestrator manages campaign lifecycle by decomposing campaigns into
// CrawlTasks, dispatching them to the task queue, and tracking results.
// It is transport-agnostic: swapping the queue implementations is enough
// to move from standalone (in-process channels) to distributed mode.
type Orchestrator struct {
	campaignRepo storage.CampaignRepository
	taskQueue    messaging.Producer[types.CrawlTask]
	resultQueue  messaging.Consumer[types.TaskResult]
	geoipDir     string

	mu     sync.Mutex
	active map[string]context.CancelFunc
	states sync.Map // campaignID → *campaignState
}

func New(
	campaignRepo storage.CampaignRepository,
	taskQueue messaging.Producer[types.CrawlTask],
	resultQueue messaging.Consumer[types.TaskResult],
	geoipDir string,
) *Orchestrator {
	return &Orchestrator{
		campaignRepo: campaignRepo,
		taskQueue:    taskQueue,
		resultQueue:  resultQueue,
		geoipDir:     geoipDir,
		active:       make(map[string]context.CancelFunc),
	}
}

// Start enqueues CIDR tasks for a campaign and begins tracking results.
// It is safe to call from multiple goroutines; duplicate starts are no-ops.
// Satisfies the api.CampaignController interface.
func (o *Orchestrator) Start(ctx context.Context, c *types.Campaign) {
	orchCtx, cancel := context.WithCancel(ctx)

	o.mu.Lock()
	if _, already := o.active[c.ID]; already {
		o.mu.Unlock()
		cancel()
		return
	}
	o.active[c.ID] = cancel
	o.mu.Unlock()

	metrics.CampaignsActive.Inc()
	go o.run(orchCtx, c, cancel)
}

func (o *Orchestrator) run(ctx context.Context, c *types.Campaign, cancel context.CancelFunc) {
	defer func() {
		metrics.CampaignsActive.Dec()
		cancel()
		o.mu.Lock()
		delete(o.active, c.ID)
		o.mu.Unlock()
	}()

	slog.Info("campaign starting", "campaign_id", c.ID, "name", c.Name)

	// Resolve country → CIDRs if needed (persisted so recovery doesn't re-resolve).
	if c.CountryCode != "" && len(c.CIDRs) == 0 {
		if o.geoipDir == "" {
			o.failCampaign(c.ID, "geoip_dir not configured; required for country-based targeting")
			return
		}
		cidrs, err := geoip.CountryCIDRs(o.geoipDir, c.CountryCode)
		if err != nil {
			o.failCampaign(c.ID, fmt.Sprintf("resolve country CIDRs for %q: %s", c.CountryCode, err))
			return
		}
		if err := o.campaignRepo.SetCampaignCIDRs(ctx, c.ID, cidrs); err != nil {
			o.failCampaign(c.ID, err.Error())
			return
		}
		c.CIDRs = cidrs
		slog.Info("country CIDRs resolved",
			"campaign_id", c.ID,
			"country", c.CountryCode,
			"cidr_count", len(cidrs),
		)
	}

	totalTasks := len(c.CIDRs) - c.Progress.CompletedCIDRs
	if totalTasks == 0 {
		o.markCompleted(c.ID, 0)
		return
	}

	state := &campaignState{
		progress:     c.Progress,
		pendingCount: totalTasks,
		doneCh:       make(chan struct{}),
	}
	state.progress.TotalCIDRs = len(c.CIDRs)
	o.states.Store(c.ID, state)

	// Enqueue one task per CIDR from the progress checkpoint.
	cfg := taskConfigFrom(c)
	for i := c.Progress.CompletedCIDRs; i < len(c.CIDRs); i++ {
		if ctx.Err() != nil {
			o.states.Delete(c.ID)
			return
		}
		task := types.CrawlTask{
			TaskID:     fmt.Sprintf("%s-%d", c.ID, i),
			CampaignID: c.ID,
			CIDR:       c.CIDRs[i],
			CIDRIndex:  i,
			Config:     cfg,
		}
		if err := o.taskQueue.Publish(ctx, []byte(c.ID), task); err != nil {
			if ctx.Err() != nil {
				o.states.Delete(c.ID)
				return
			}
			slog.Error("failed to enqueue task", "campaign_id", c.ID, "cidr", task.CIDR, "err", err)
		}
	}

	// Wait for all task results or context cancellation (pause/cancel/shutdown).
	select {
	case <-state.doneCh:
		state.mu.Lock()
		finalServices := state.progress.ServicesFound
		state.mu.Unlock()
		o.states.Delete(c.ID)
		if ctx.Err() == nil {
			o.markCompleted(c.ID, finalServices)
		}
	case <-ctx.Done():
		o.states.Delete(c.ID)
	}
}

// Cancel stops a running campaign and marks it cancelled.
func (o *Orchestrator) Cancel(id string) error {
	o.mu.Lock()
	cancelFn, ok := o.active[id]
	o.mu.Unlock()
	if !ok {
		return fmt.Errorf("campaign %q is not running", id)
	}
	cancelFn()
	ctx, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()
	return o.campaignRepo.UpdateCampaignStatus(ctx, id, types.CampaignCancelled, "")
}

// Pause stops task enqueuing and marks the campaign paused.
// Progress is preserved; Start may be called again to resume.
func (o *Orchestrator) Pause(id string) error {
	o.mu.Lock()
	cancelFn, ok := o.active[id]
	o.mu.Unlock()
	if !ok {
		return fmt.Errorf("campaign %q is not running", id)
	}
	cancelFn()
	ctx, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()
	return o.campaignRepo.UpdateCampaignStatus(ctx, id, types.CampaignPaused, "")
}

// IsActive reports whether this orchestrator instance is managing the campaign.
func (o *Orchestrator) IsActive(id string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	_, ok := o.active[id]
	return ok
}

// Run starts the result consumer loop. It blocks until ctx is cancelled.
// Must be called once in a background goroutine.
func (o *Orchestrator) Run(ctx context.Context) error {
	for {
		msg, err := o.resultQueue.Fetch(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Error("result queue fetch error", "err", err)
			continue
		}
		o.handleResult(ctx, msg.Value)
	}
}

func (o *Orchestrator) handleResult(ctx context.Context, result types.TaskResult) {
	val, ok := o.states.Load(result.CampaignID)
	if !ok {
		return // campaign was paused, cancelled, or already completed
	}
	state := val.(*campaignState)

	state.mu.Lock()
	state.progress.CompletedCIDRs++
	state.progress.ServicesFound += result.ServicesFound
	progress := state.progress
	state.pendingCount--
	remaining := state.pendingCount
	state.mu.Unlock()

	updateCtx, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()
	_ = o.campaignRepo.UpdateCampaignProgress(updateCtx, result.CampaignID, progress)

	if result.Error != "" {
		slog.Error("CIDR scan error", "campaign_id", result.CampaignID, "cidr_index", result.CIDRIndex, "err", result.Error)
	}

	if remaining == 0 {
		close(state.doneCh)
	}
}

func (o *Orchestrator) markCompleted(id string, servicesFound int64) {
	ctx, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()
	_ = o.campaignRepo.UpdateCampaignStatus(ctx, id, types.CampaignCompleted, "")
	slog.Info("campaign complete", "campaign_id", id, "services_found", servicesFound)
}

func (o *Orchestrator) failCampaign(id, msg string) {
	ctx, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()
	_ = o.campaignRepo.UpdateCampaignStatus(ctx, id, types.CampaignFailed, msg)
	slog.Error("campaign failed", "campaign_id", id, "reason", msg)
}

// taskConfigFrom snapshots the scan parameters from a Campaign into a TaskConfig.
// Workers use TaskConfig exclusively — they never read the Campaign document.
func taskConfigFrom(c *types.Campaign) types.TaskConfig {
	topPorts := ""
	if c.Ports == "" {
		topPorts = defaultStr(c.TopPorts, types.DefaultTopPorts)
	}
	return types.TaskConfig{
		ScanType:      defaultStr(c.ScanType, types.ScanTypeSyn),
		Ports:         c.Ports,
		TopPorts:      topPorts,
		Retries:       defaultInt(c.Retries, 1),
		Rate:          defaultInt(c.Rate, 500),
		Timeout:       defaultInt(c.Timeout, 1000),
		FastMode:      c.FastMode,
		UseUDP:        c.UseUDP,
		Workers:       c.Workers,
		QueueSize:     c.QueueSize,
		ServiceFilter: c.ServiceFilter,
		RunVulnScan:   c.RunVulnScan,
	}
}

func defaultStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func defaultInt(v, fallback int) int {
	if v == 0 {
		return fallback
	}
	return v
}
