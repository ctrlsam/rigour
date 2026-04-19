package worker

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ctrlsam/rigour/internal/messaging"
	"github.com/ctrlsam/rigour/internal/metrics"
	"github.com/ctrlsam/rigour/pkg/crawler"
	"github.com/ctrlsam/rigour/pkg/crawler/discovery"
	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint"
	"github.com/ctrlsam/rigour/pkg/types"
)

// Worker consumes CrawlTasks, executes a port-discovery + fingerprinting scan
// for each CIDR, and publishes discovered services and a completion result.
// It is stateless with respect to campaigns — all scan parameters come from
// the task itself.
type Worker struct {
	taskQueue   messaging.Consumer[types.CrawlTask]
	svcQueue    messaging.Producer[types.Service]
	resultQueue messaging.Producer[types.TaskResult]
	concurrency int
}

func New(
	taskQueue messaging.Consumer[types.CrawlTask],
	svcQueue messaging.Producer[types.Service],
	resultQueue messaging.Producer[types.TaskResult],
	concurrency int,
) *Worker {
	if concurrency <= 0 {
		concurrency = 3
	}
	return &Worker{
		taskQueue:   taskQueue,
		svcQueue:    svcQueue,
		resultQueue: resultQueue,
		concurrency: concurrency,
	}
}

// Run blocks, consuming tasks until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) error {
	sem := make(chan struct{}, w.concurrency)
	var wg sync.WaitGroup

	defer func() {
		wg.Wait()
	}()

	for {
		msg, err := w.taskQueue.Fetch(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Error("task queue fetch error", "err", err)
			continue
		}
		task := msg.Value

		sem <- struct{}{}
		wg.Add(1)
		go func(t types.CrawlTask) {
			defer func() {
				<-sem
				wg.Done()
			}()
			w.execute(ctx, t)
		}(task)
	}
}

func (w *Worker) execute(ctx context.Context, task types.CrawlTask) {
	slog.Info("worker scanning CIDR",
		"campaign_id", task.CampaignID,
		"cidr_index", task.CIDRIndex,
		"cidr", task.CIDR,
	)

	cfg := task.Config
	discoveryCfg := discovery.DiscoveryConfig{
		ScanType: cfg.ScanType,
		Ports:    cfg.Ports,
		TopPorts: cfg.TopPorts,
		Retries:  cfg.Retries,
		Rate:     cfg.Rate,
	}
	fpCfg := fingerprint.FingerprintConfig{
		DefaultTimeout: time.Duration(defaultInt(cfg.Timeout, 1000)) * time.Millisecond,
		FastMode:       cfg.FastMode,
		UDP:            cfg.UseUDP,
		PluginFilter:   cfg.ServiceFilter,
		Workers:        cfg.Workers,
		QueueSize:      cfg.QueueSize,
	}

	var found atomic.Int64
	onEvent := func(svc types.Service) {
		svc.CampaignID = task.CampaignID
		svc.RunVulnScan = task.Config.RunVulnScan
		if err := w.svcQueue.Publish(ctx, []byte(svc.IP), svc); err == nil {
			found.Add(1)
			metrics.CampaignServicesFound.Inc()
		}
	}

	var scanErr string
	if err := crawler.ScanTargetWithDiscoveryStream(ctx, task.CIDR, discoveryCfg, fpCfg, onEvent); err != nil {
		if ctx.Err() == nil {
			slog.Error("worker CIDR scan error", "campaign_id", task.CampaignID, "cidr", task.CIDR, "err", err)
			scanErr = err.Error()
		}
	}

	servicesFound := found.Load()
	slog.Info("worker CIDR complete",
		"campaign_id", task.CampaignID,
		"cidr_index", task.CIDRIndex,
		"cidr", task.CIDR,
		"services_found", servicesFound,
	)

	result := types.TaskResult{
		TaskID:        task.TaskID,
		CampaignID:    task.CampaignID,
		CIDRIndex:     task.CIDRIndex,
		ServicesFound: servicesFound,
		Error:         scanErr,
		CompletedAt:   time.Now(),
	}
	if err := w.resultQueue.Publish(ctx, []byte(task.CampaignID), result); err != nil {
		if ctx.Err() == nil {
			slog.Error("failed to publish task result", "task_id", task.TaskID, "err", err)
		}
	}
}

func defaultInt(v, fallback int) int {
	if v == 0 {
		return fallback
	}
	return v
}
