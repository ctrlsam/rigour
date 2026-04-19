package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	internalconst "github.com/ctrlsam/rigour/internal"
	"github.com/ctrlsam/rigour/internal/api"
	"github.com/ctrlsam/rigour/internal/logging"
	"github.com/ctrlsam/rigour/internal/messaging/channel"
	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/ctrlsam/rigour/internal/storage/mongodb"
	"github.com/ctrlsam/rigour/pkg/orchestrator"
	"github.com/ctrlsam/rigour/pkg/persistence"
	"github.com/ctrlsam/rigour/pkg/types"
	"github.com/ctrlsam/rigour/pkg/worker"
	"github.com/spf13/cobra"
)

type config struct {
	addr      string
	mongoURI  string
	geoipPath string

	vulnScan          bool
	nucleiTemplates   string
	nucleiTags        []string
	nucleiConcurrency int

	noResumeCampaigns bool

	logFormat string // "text" | "json"
	logLevel  string // "debug" | "info" | "warn" | "error"
}

func main() {
	var cfg config

	root := &cobra.Command{
		Use:   "rigour-server",
		Short: "Rigour server — API + campaign runner + persistence in one process",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), cfg)
		},
	}

	root.Flags().StringVar(&cfg.addr, "addr", ":8080", "HTTP server address")
	root.Flags().StringVar(&cfg.mongoURI, "mongo-uri", "mongodb://localhost:27017", "MongoDB connection URI")
	root.Flags().StringVar(&cfg.geoipPath, "geoip-path", "", "Path to GeoIP data directory (optional)")
	root.Flags().BoolVar(&cfg.vulnScan, "vuln-scan", false, "Enable automated vulnerability scanning via Nuclei")
	root.Flags().StringVar(&cfg.nucleiTemplates, "nuclei-templates", "", "Path to Nuclei template directory (required when --vuln-scan is set)")
	root.Flags().StringSliceVar(&cfg.nucleiTags, "nuclei-tags", nil, "Comma-separated Nuclei template tags to run (default: cve,rce,default-login,exposure,misconfig)")
	root.Flags().IntVar(&cfg.nucleiConcurrency, "nuclei-concurrency", 5, "Max concurrent Nuclei scans")
	root.Flags().BoolVar(&cfg.noResumeCampaigns, "no-resume", false, "Do not resume campaigns that were in-progress when the server last stopped")
	root.Flags().StringVar(&cfg.logFormat, "log-format", "text", "Log output format: text or json")
	root.Flags().StringVar(&cfg.logLevel, "log-level", "info", "Log level: debug, info, warn, error")

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config) error {
	lvl := parseLevel(cfg.logLevel)
	opts := &slog.HandlerOptions{Level: lvl}
	var h slog.Handler
	if cfg.logFormat == "json" {
		h = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		h = logging.NewColorHandler(os.Stderr, opts)
	}
	slog.SetDefault(slog.New(h))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	mongoClient, err := mongodb.NewClient(ctx, cfg.mongoURI, 10*time.Second)
	if err != nil {
		return fmt.Errorf("mongodb: %w", err)
	}
	defer mongoClient.Close(context.Background())

	repo, err := mongoClient.NewHostsRepository(ctx, storage.RepositoryConfig{
		Database:   internalconst.DatabaseName,
		Collection: internalconst.HostsRepositoryName,
	})
	if err != nil {
		return fmt.Errorf("repository: %w", err)
	}

	campaignRepo, err := mongoClient.NewCampaignRepository(ctx, storage.RepositoryConfig{
		Database:   internalconst.DatabaseName,
		Collection: internalconst.CampaignsRepositoryName,
	})
	if err != nil {
		return fmt.Errorf("campaign repository: %w", err)
	}

	svcChannel := channel.New[types.Service](1000)
	defer svcChannel.Close()

	persistApp, err := persistence.NewApp(ctx, persistence.Config{
		DbURI:              cfg.mongoURI,
		GeoIPDataDir:       cfg.geoipPath,
		VulnScanEnabled:    cfg.vulnScan,
		NucleiTemplateDir:  cfg.nucleiTemplates,
		NucleiTemplateTags: cfg.nucleiTags,
		NucleiConcurrency:  cfg.nucleiConcurrency,
	}, svcChannel)
	if err != nil {
		return fmt.Errorf("persistence: %w", err)
	}
	defer persistApp.Close(context.Background())

	go func() {
		if err := persistApp.Run(ctx); err != nil && err != context.Canceled {
			slog.Error("persistence worker error", "err", err)
		}
	}()

	// Wire vuln repo to API if vuln scanning is enabled.
	var vulnRepo storage.VulnerabilityRepository
	if cfg.vulnScan {
		vulnRepo, err = mongoClient.NewVulnerabilityRepository(ctx, storage.RepositoryConfig{
			Database:   internalconst.DatabaseName,
			Collection: internalconst.VulnerabilitiesRepositoryName,
		})
		if err != nil {
			return fmt.Errorf("vuln repository: %w", err)
		}
	}

	taskChannel := channel.New[types.CrawlTask](10_000)
	defer taskChannel.Close()
	resultChannel := channel.New[types.TaskResult](10_000)
	defer resultChannel.Close()

	orch := orchestrator.New(campaignRepo, taskChannel, resultChannel, cfg.geoipPath)
	crawlWorker := worker.New(taskChannel, svcChannel, resultChannel, 3)

	go func() {
		if err := orch.Run(ctx); err != nil && err != context.Canceled {
			slog.Error("orchestrator error", "err", err)
		}
	}()
	go func() {
		if err := crawlWorker.Run(ctx); err != nil && err != context.Canceled {
			slog.Error("crawl worker error", "err", err)
		}
	}()

	if interrupted, err := campaignRepo.ListCampaigns(ctx, storage.CampaignFilter{Status: types.CampaignRunning}); err != nil {
		slog.Error("failed to list interrupted campaigns", "err", err)
	} else if cfg.noResumeCampaigns {
		slog.Info("campaign auto-resume disabled, pausing interrupted campaigns", "flag", "--no-resume", "count", len(interrupted))
		for _, c := range interrupted {
			if err := campaignRepo.UpdateCampaignStatus(ctx, c.ID, types.CampaignPaused, ""); err != nil {
				slog.Error("failed to pause interrupted campaign on startup", "campaign_id", c.ID, "err", err)
			}
		}
	} else {
		for _, c := range interrupted {
			c := c
			go orch.Start(ctx, &c)
			slog.Info("resuming interrupted campaign", "campaign_id", c.ID)
		}
	}

	router := api.NewRouter(ctx, repo, campaignRepo, orch, vulnRepo)
	server := &http.Server{
		Addr:         cfg.addr,
		Handler:      router.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("API server listening", "addr", cfg.addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "err", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)

	slog.Info("server stopped")
	return nil
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
