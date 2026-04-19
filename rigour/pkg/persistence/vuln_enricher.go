package persistence

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"sync"
	"time"

	nuclei "github.com/projectdiscovery/nuclei/v3/lib"
	"github.com/projectdiscovery/nuclei/v3/pkg/output"

	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/ctrlsam/rigour/pkg/types"
)

// VulnEnricherConfig holds configuration for the vulnerability enricher.
type VulnEnricherConfig struct {
	NucleiTemplateDir string
	TemplateTags      []string // e.g. ["cve", "rce", "default-login", "exposure"]
	Concurrency       int      // max concurrent Nuclei scans
}

// VulnEnricher dispatches Nuclei scans for discovered services and persists findings.
// Scans are dispatched asynchronously; call Wait() before shutdown.
type VulnEnricher struct {
	engine   *nuclei.ThreadSafeNucleiEngine
	vulnRepo storage.VulnerabilityRepository
	tags     []string

	// sem limits concurrent scans
	sem chan struct{}
	wg  sync.WaitGroup

	// campaignByTarget maps target URL string to campaign ID so the global
	// result callback can stamp findings with the originating campaign.
	campaignByTarget sync.Map
}

// NewVulnEnricher initialises the shared Nuclei engine and returns a VulnEnricher.
func NewVulnEnricher(ctx context.Context, cfg VulnEnricherConfig, vulnRepo storage.VulnerabilityRepository) (*VulnEnricher, error) {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 5
	}
	if len(cfg.TemplateTags) == 0 {
		cfg.TemplateTags = []string{"cve", "rce", "default-login", "exposure"}
	}

	ve := &VulnEnricher{
		vulnRepo: vulnRepo,
		tags:     cfg.TemplateTags,
		sem:      make(chan struct{}, cfg.Concurrency),
	}

	opts := []nuclei.NucleiSDKOptions{
		nuclei.WithTemplateFilters(nuclei.TemplateFilters{
			Tags: cfg.TemplateTags,
		}),
	}
	if cfg.NucleiTemplateDir != "" {
		slog.Info("vuln enricher using custom templates", "dir", cfg.NucleiTemplateDir)
		opts = append(opts, nuclei.WithTemplatesOrWorkflows(nuclei.TemplateSources{
			Templates: []string{cfg.NucleiTemplateDir},
		}))
	} else {
		slog.Info("vuln enricher using default nuclei templates", "path", "~/nuclei-templates")
	}

	slog.Info("initialising nuclei engine", "tags", cfg.TemplateTags, "concurrency", cfg.Concurrency)
	engine, err := nuclei.NewThreadSafeNucleiEngineCtx(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("vuln enricher: init nuclei engine: %w", err)
	}

	engine.GlobalResultCallback(func(event *output.ResultEvent) {
		ve.handleResult(event)
	})

	slog.Info("loading nuclei templates")
	if err := engine.GlobalLoadAllTemplates(); err != nil {
		engine.Close()
		return nil, fmt.Errorf("vuln enricher: load templates: %w", err)
	}
	slog.Info("nuclei templates loaded")

	ve.engine = engine
	return ve, nil
}

// Dispatch schedules a Nuclei scan for svc in a background goroutine.
// It respects the concurrency limit and returns immediately.
func (ve *VulnEnricher) Dispatch(ctx context.Context, svc types.Service) error {
	target := buildTarget(svc)
	if target == "" {
		return nil
	}

	bare := fmt.Sprintf("%s:%d", svc.IP, svc.Port)
	ve.campaignByTarget.Store(target, svc.CampaignID)
	ve.campaignByTarget.Store(bare, svc.CampaignID)

	ve.wg.Add(1)
	go func() {
		ve.sem <- struct{}{} // acquire slot inside goroutine so Dispatch never blocks caller
		defer ve.wg.Done()
		defer func() { <-ve.sem }()

		slog.Info("vuln scan dispatched", "target", target)
		scanCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		_ = ve.engine.ExecuteNucleiWithOptsCtx(scanCtx, []string{target})
		ve.campaignByTarget.Delete(target)
		ve.campaignByTarget.Delete(bare)
		slog.Info("vuln scan complete", "target", target)
	}()
	return nil
}

// Wait blocks until all in-flight scans complete.
func (ve *VulnEnricher) Wait() {
	ve.wg.Wait()
}

// Close shuts down the Nuclei engine. Call Wait() first.
func (ve *VulnEnricher) Close() {
	if ve.engine != nil {
		ve.engine.Close()
	}
}

func (ve *VulnEnricher) handleResult(event *output.ResultEvent) {
	// GlobalResultCallback fires for every template execution, not just matches.
	// MatcherStatus is false when the template ran but conditions were NOT met.
	if !event.MatcherStatus {
		return
	}
	// Nuclei fires cluster-* results for batched request groups; these have no
	// template name or severity and are not real findings — drop them.
	if event.Info.Name == "" {
		return
	}

	campaignID := ""
	for _, key := range []string{event.Host, event.IP + ":" + event.Port, event.IP} {
		if v, ok := ve.campaignByTarget.Load(key); ok {
			campaignID, _ = v.(string)
			break
		}
	}

	ip, port, protocol := parseTargetFields(event.Host)
	// Use Nuclei's dedicated IP and port fields when the host-parse misses them.
	if event.IP != "" {
		ip = event.IP
	}
	if port == 0 && event.Port != "" {
		fmt.Sscanf(event.Port, "%d", &port)
	}

	// event.Matched is not populated by the SDK global callback; fall back to
	// event.Host which contains the full target URL.
	matchedAt := event.Matched
	if matchedAt == "" {
		matchedAt = event.Host
	}

	cvss := 0.0
	var cveIDs []string
	if event.Info.Classification != nil {
		cvss = event.Info.Classification.CVSSScore
		cveIDs = event.Info.Classification.CVEID.ToSlice()
	}

	vuln := types.Vulnerability{
		IP:           ip,
		Port:         port,
		Protocol:     protocol,
		TemplateID:   event.TemplateID,
		Name:         event.Info.Name,
		Severity:     event.Info.SeverityHolder.Severity.String(),
		Description:  event.Info.Description,
		CVSS:         cvss,
		CVEIDs:       cveIDs,
		MatchedAt:    matchedAt,
		DiscoveredAt: time.Now().UTC(),
		CampaignID:   campaignID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	slog.Warn("vulnerability found",
		"template_id", event.TemplateID,
		"severity", vuln.Severity,
		"ip", ip,
		"port", port,
		"name", vuln.Name,
	)
	if err := ve.vulnRepo.UpsertVulnerability(ctx, vuln); err != nil {
		slog.Error("failed to store vulnerability", "template_id", event.TemplateID, "err", err)
	}
}

// buildTarget constructs a Nuclei target from a discovered service.
// HTTP/HTTPS services use URL form so HTTP templates apply correctly.
// All other services use ip:port so Nuclei network/TCP templates can match them.
func buildTarget(svc types.Service) string {
	switch svc.Protocol {
	case "http":
		scheme := "http"
		if svc.TLS {
			scheme = "https"
		}
		return fmt.Sprintf("%s://%s:%d", scheme, svc.IP, svc.Port)
	case "https":
		return fmt.Sprintf("https://%s:%d", svc.IP, svc.Port)
	default:
		return fmt.Sprintf("%s:%d", svc.IP, svc.Port)
	}
}

// parseTargetFields extracts ip, port, and protocol from a target string.
func parseTargetFields(target string) (ip string, port int, protocol string) {
	u, err := url.Parse(target)
	if err == nil && u.Host != "" {
		ip = u.Hostname()
		protocol = u.Scheme
		fmt.Sscanf(u.Port(), "%d", &port)
		return
	}
	// fallback: bare "ip:port"
	if host, portStr, splitErr := net.SplitHostPort(target); splitErr == nil {
		ip = host
		fmt.Sscanf(portStr, "%d", &port)
		return
	}
	ip = target
	return
}
