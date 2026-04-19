package persistence

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ctrlsam/rigour/internal/messaging"
	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/ctrlsam/rigour/internal/storage/mongodb"
	"github.com/ctrlsam/rigour/pkg/types"
)

// App wires a service consumer + Mongo repository + enricher pipeline.
type App struct {
	cfg      Config
	consumer messaging.Consumer[types.Service]
	repo     storage.HostRepository
	enrichers []HostEnricher

	vulnEnricher *VulnEnricher
	mongoClient  *mongodb.Client
}

// NewApp creates an App. consumer is the source of service events; for single-node
// mode pass a *channel.TypedChannel, for cluster mode pass a Kafka consumer.
func NewApp(ctx context.Context, cfg Config, consumer messaging.Consumer[types.Service]) (*App, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	mongoClient, err := mongodb.NewClient(ctx, cfg.DbURI, cfg.DbTimeout)
	if err != nil {
		return nil, fmt.Errorf("persistence: mongodb client: %w", err)
	}

	repo, err := mongoClient.NewHostsRepository(ctx, storage.RepositoryConfig{
		Database:   cfg.DbName,
		Collection: cfg.DbCollection,
	})
	if err != nil {
		_ = mongoClient.Close(ctx)
		return nil, fmt.Errorf("persistence: hosts repository: %w", err)
	}

	var enrichers []HostEnricher
	if cfg.GeoIPDataDir != "" {
		geoipReaders, err := OpenGeoIPReaders(cfg.GeoIPDataDir)
		if err != nil {
			_ = mongoClient.Close(ctx)
			return nil, fmt.Errorf("persistence: geoip: %w", err)
		}
		enrichers = append(enrichers, NewEnricher(geoipReaders))
	}

	app := &App{
		cfg:         cfg,
		consumer:    consumer,
		repo:        repo,
		enrichers:   enrichers,
		mongoClient: mongoClient,
	}

	if cfg.VulnScanEnabled {
		vulnRepo, err := mongoClient.NewVulnerabilityRepository(ctx, storage.RepositoryConfig{
			Database:   cfg.DbName,
			Collection: cfg.DbVulnCollection,
		})
		if err != nil {
			_ = mongoClient.Close(ctx)
			return nil, fmt.Errorf("persistence: vuln repository: %w", err)
		}
		ve, err := NewVulnEnricher(ctx, VulnEnricherConfig{
			NucleiTemplateDir: cfg.NucleiTemplateDir,
			TemplateTags:      cfg.NucleiTemplateTags,
			Concurrency:       cfg.NucleiConcurrency,
		}, vulnRepo)
		if err != nil {
			_ = mongoClient.Close(ctx)
			return nil, fmt.Errorf("persistence: vuln enricher: %w", err)
		}
		app.vulnEnricher = ve
	}

	return app, nil
}

// Close closes underlying resources.
func (app *App) Close(ctx context.Context) error {
	var firstErr error
	if app == nil {
		return nil
	}
	if app.vulnEnricher != nil {
		app.vulnEnricher.Wait()
	}
	for _, e := range app.enrichers {
		e.Close()
	}
	if app.mongoClient != nil {
		if err := app.mongoClient.Close(ctx); err != nil {
			firstErr = err
		}
	}
	return firstErr
}

// Run blocks consuming messages until ctx is canceled.
func (app *App) Run(ctx context.Context) error {
	if app == nil {
		return errors.New("persistence: app is nil")
	}
	slog.Info("persistence consumer started")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := app.consumer.Fetch(ctx)
		if err != nil {
			return err
		}
		if msg == nil {
			continue
		}

		if err := app.handleService(ctx, msg.Value); err != nil {
			slog.Error("persistence: failed to handle service, skipping",
				"ip", msg.Value.IP,
				"port", msg.Value.Port,
				"err", err,
			)
		}
	}
}

func (app *App) handleService(ctx context.Context, svc types.Service) error {
	slog.Debug("processing service", "ip", svc.IP, "port", svc.Port, "protocol", svc.Protocol)

	now := time.Now()
	if svc.LastScan.IsZero() {
		svc.LastScan = now
	}

	// 1. Ensure host exists (first time only).
	if err := app.repo.EnsureHost(ctx, svc.IP, now); err != nil {
		return err
	}

	// 2. Run enrichment pipeline.
	host := &types.Host{IP: svc.IP, LastSeen: now}
	var err error
	for _, e := range app.enrichers {
		host, err = e.EnrichHost(ctx, host)
		if err != nil {
			return err
		}
	}

	// 3. Update host with enrichment data (ASN, Location, Labels).
	if err := app.repo.UpdateHost(ctx, *host); err != nil {
		return err
	}

	// 4. Upsert service under the enriched host.
	if err := app.repo.UpsertService(ctx, svc); err != nil {
		return err
	}

	// 5. Dispatch vuln scan if enabled for this campaign (best-effort, non-blocking).
	if svc.RunVulnScan && app.vulnEnricher != nil {
		if err := app.vulnEnricher.Dispatch(ctx, svc); err != nil {
			slog.Error("vuln scan dispatch error", "err", err)
		}
	}

	return nil
}
