package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	internalconst "github.com/ctrlsam/rigour/internal"
	"github.com/ctrlsam/rigour/internal/geoip"
	"github.com/ctrlsam/rigour/internal/messaging/kafka"
	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/ctrlsam/rigour/internal/storage/mongodb"
	"github.com/ctrlsam/rigour/pkg/types"
)

// Config is the persistence application's runtime configuration.
//
// Contract:
// - Consumes types.Service events from Kafka topic internalconst.KafkaTopicScannedServices
// - Ensures host exists, upserts the service under the host
// - Enriches the host (GeoIP/ASN) and updates the host document
//
// Error modes:
// - Any error returned from Mongo/Kafka init is fatal
// - Per-message processing errors are returned (caller can decide to stop or keep going)
//
// Note: this package intentionally keeps most of the app logic out of cmd/.
//
// Deprecated fields are not present.
type Config struct {
	KafkaBrokers string
	KafkaGroupID string
	Topic        string

	MongoURI        string
	MongoDatabase   string
	MongoCollection string
	MongoTimeout    time.Duration

	GeoIPBaseURL string
	GeoIPAPIKey  string
}

func (c Config) withDefaults() Config {
	out := c
	if out.Topic == "" {
		out.Topic = internalconst.KafkaTopicScannedServices
	}
	if out.MongoDatabase == "" {
		out.MongoDatabase = internalconst.DatabaseName
	}
	if out.MongoCollection == "" {
		out.MongoCollection = internalconst.HostsRepositoryName
	}
	if out.MongoTimeout <= 0 {
		out.MongoTimeout = 10 * time.Second
	}
	return out
}

func (c Config) Validate() error {
	c = c.withDefaults()
	if c.KafkaBrokers == "" {
		return errors.New("persistence: kafka brokers is required")
	}
	if c.KafkaGroupID == "" {
		return errors.New("persistence: kafka group id is required")
	}
	if c.MongoURI == "" {
		return errors.New("persistence: mongo uri is required")
	}
	if c.GeoIPBaseURL == "" || c.GeoIPAPIKey == "" {
		return errors.New("persistence: geoip base url and api key are required")
	}
	return nil
}

// App wires Kafka consumer + Mongo repository + enricher.
type App struct {
	cfg Config

	consumer *kafka.TypedConsumer[types.Service]
	repo     storage.HostRepository
	enricher *Enricher

	mongoClient *mongodb.Client
}

func NewApp(ctx context.Context, cfg Config) (*App, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	consumer, err := kafka.NewTypedConsumer[types.Service](kafka.ConsumerConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.Topic,
		GroupID: cfg.KafkaGroupID,
	})
	if err != nil {
		return nil, fmt.Errorf("persistence: kafka consumer: %w", err)
	}

	mongoClient, err := mongodb.NewClient(ctx, cfg.MongoURI, cfg.MongoTimeout)
	if err != nil {
		_ = consumer.Close()
		return nil, fmt.Errorf("persistence: mongodb client: %w", err)
	}

	repo, err := mongoClient.NewHostsRepository(ctx, storage.RepositoryConfig{
		URI:        cfg.MongoURI,
		Database:   cfg.MongoDatabase,
		Collection: cfg.MongoCollection,
		Timeout:    int(cfg.MongoTimeout.Seconds()),
	})
	if err != nil {
		_ = consumer.Close()
		_ = mongoClient.Close(ctx)
		return nil, fmt.Errorf("persistence: hosts repository: %w", err)
	}

	geoEnricher, err := newGeoEnricher(cfg)
	if err != nil {
		_ = consumer.Close()
		_ = mongoClient.Close(ctx)
		return nil, err
	}

	return &App{
		cfg:         cfg,
		consumer:    consumer,
		repo:        repo,
		enricher:    geoEnricher,
		mongoClient: mongoClient,
	}, nil
}

func newGeoEnricher(cfg Config) (*Enricher, error) {
	client, err := newGeoIPClient(cfg)
	if err != nil {
		return nil, err
	}
	return NewEnricher(client), nil
}

func newGeoIPClient(cfg Config) (*geoip.Client, error) {
	return geoip.NewClient(cfg.GeoIPBaseURL, cfg.GeoIPAPIKey, 3*time.Second)
}

// Close closes underlying resources.
func (app *App) Close(ctx context.Context) error {
	var firstErr error
	if app == nil {
		return nil
	}
	if app.consumer != nil {
		if err := app.consumer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if app.mongoClient != nil {
		if err := app.mongoClient.Close(ctx); err != nil && firstErr == nil {
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
	fmt.Println("persistence: started consuming messages...")
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
			return err
		}
	}
}

func (app *App) handleService(ctx context.Context, svc types.Service) error {
	fmt.Println("persistence: processing service:", svc.IP, svc.Port, svc.Protocol)

	now := time.Now()
	if svc.LastScan.IsZero() {
		svc.LastScan = now
	}

	// 1. Ensure host exists (first time only).
	if err := app.repo.EnsureHost(ctx, svc.IP, now); err != nil {
		return err
	}

	// 2. Enrich host with GeoIP/ASN data.
	host := &types.Host{IP: svc.IP, LastSeen: now}
	host, err := app.enricher.EnrichHost(ctx, host)
	if err != nil {
		return err
	}

	// 3. Update host with enrichment data (ASN, Location, Labels).
	if err := app.repo.UpdateHost(ctx, *host); err != nil {
		return err
	}

	// 4. Upsert service under the enriched host.
	return app.repo.UpsertService(ctx, svc)
}
