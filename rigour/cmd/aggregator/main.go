package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ctrlsam/rigour/internal"
	"github.com/ctrlsam/rigour/internal/geoip/gogeoip"
	"github.com/ctrlsam/rigour/internal/messaging"
	"github.com/ctrlsam/rigour/internal/messaging/kafka"
	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/ctrlsam/rigour/internal/storage/mongodb"
	"github.com/ctrlsam/rigour/pkg/enricher"
	"github.com/ctrlsam/rigour/pkg/scanner"
	"github.com/spf13/cobra"
)

type cliConfig struct {
	kafkaBrokers string
	kafkaTopic   string
	kafkaGroupID string

	mongoURI        string
	mongoDatabase   string
	mongoCollection string

	geoipBaseURL string
	geoipAPIKey  string

	logVerbose bool
}

type services struct {
	mongoClient *mongodb.Client
	repository  storage.ServiceRepository
	enricher    *enricher.Enricher
	consumer    messaging.Consumer
	codec       messaging.Codec[scanner.ScannedServiceEvent]
	geoipClient *gogeoip.Client
}

var config cliConfig

var rootCmd = &cobra.Command{
	Use:   "rigour-aggregator",
	Short: "Service aggregator for Rigour",
	Long:  "Consumes scanned service events from Kafka, enriches them with GeoIP data, and stores them in MongoDB",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAggregator(cmd.Context())
	},
}

func init() {
	rootCmd.Flags().StringVar(&config.kafkaBrokers, "kafka-brokers", "localhost:29092", "Kafka brokers (comma-separated host:port)")
	rootCmd.Flags().StringVar(&config.kafkaTopic, "kafka-topic", internal.KafkaTopicScannedServices, "Kafka topic")
	rootCmd.Flags().StringVar(&config.kafkaGroupID, "kafka-group", "rigour-aggregator", "Kafka consumer group id")

	rootCmd.Flags().StringVar(&config.mongoURI, "mongo-uri", "mongodb://localhost:27017", "MongoDB connection URI")
	rootCmd.Flags().StringVar(&config.mongoDatabase, "mongo-db", internal.DatabaseName, "MongoDB database")
	rootCmd.Flags().StringVar(&config.mongoCollection, "mongo-collection", internal.CollectionServices, "MongoDB collection")

	rootCmd.Flags().StringVar(&config.geoipBaseURL, "geoip-base-url", "http://localhost:5000", "GeoIP service base URL")
	rootCmd.Flags().StringVar(&config.geoipAPIKey, "geoip-api-key", "", "GeoIP service Authorization header value")
	rootCmd.Flags().BoolVar(&config.logVerbose, "verbose", false, "Verbose logging")
}

func main() {
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runAggregator(ctx context.Context) error {
	if strings.TrimSpace(config.geoipAPIKey) == "" {
		log.Println("warning: geoip-api-key is empty; GeoIP enrichment will be skipped")
	}

	svc, err := createServices(ctx, config)
	if err != nil {
		return fmt.Errorf("create services: %w", err)
	}
	defer func() {
		if err := svc.close(context.Background()); err != nil {
			log.Printf("error during shutdown: %v", err)
		}
	}()

	// Log startup info
	log.Printf("aggregator started: topic=%s group=%s mongo=%s/%s geoip=%s",
		config.kafkaTopic, config.kafkaGroupID, config.mongoDatabase, config.mongoCollection, config.geoipBaseURL)

	// Start consuming messages
	if err := run(ctx, svc); err != nil {
		return fmt.Errorf("worker error: %w", err)
	}

	log.Println("aggregator stopped")
	return nil
}

func run(ctx context.Context, svc *services) error {
	for {
		msg, err := svc.consumer.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			log.Printf("consumer fetch error: %v", err)
			continue
		}

		var serviceEvent scanner.ScannedServiceEvent
		if err := svc.codec.Unmarshal(msg.Value, &serviceEvent); err != nil {
			log.Printf("decode message failed: %v", err)
			continue
		}

		enrichedRecord, err := svc.enricher.EnrichEvent(ctx, &serviceEvent)
		if err != nil {
			log.Printf("enrich event failed: %v", err)
			continue
		}

		writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = svc.repository.InsertServiceRecord(writeCtx, *enrichedRecord)
		cancel()

		if err != nil {
			log.Printf("insert record failed: %v", err)
			continue
		}
	}
}

func createServices(ctx context.Context, cfg cliConfig) (*services, error) {
	// Initialize services directly from their packages
	svc := &services{}

	// Initialize MongoDB client
	mongoClient, err := mongodb.NewClient(ctx, cfg.mongoURI, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("initialize mongodb client: %w", err)
	}
	svc.mongoClient = mongoClient

	// Initialize repository
	repository, err := mongoClient.NewRepository(ctx, storage.RepositoryConfig{
		Database:   cfg.mongoDatabase,
		Collection: cfg.mongoCollection,
		Timeout:    10,
	})
	if err != nil {
		mongoClient.Close(context.Background())
		return nil, fmt.Errorf("initialize repository: %w", err)
	}
	svc.repository = repository

	// Initialize GeoIP client (optional)
	var geoipClient *gogeoip.Client
	if strings.TrimSpace(cfg.geoipAPIKey) != "" {
		geoipClient, err = gogeoip.NewClient(cfg.geoipBaseURL, cfg.geoipAPIKey, 3*time.Second)
		if err != nil {
			log.Printf("warning: geoip initialization failed: %v", err)
		}
	}
	svc.geoipClient = geoipClient

	// Initialize enricher
	svc.enricher = enricher.NewEnricher(geoipClient)

	// Initialize message consumer
	consumer, err := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: cfg.kafkaBrokers,
		Topic:   cfg.kafkaTopic,
		GroupID: cfg.kafkaGroupID,
	})
	if err != nil {
		svc.close(context.Background())
		return nil, fmt.Errorf("initialize message consumer: %w", err)
	}
	svc.consumer = consumer

	// Initialize message codec
	svc.codec = messaging.JSONCodec[scanner.ScannedServiceEvent]{}

	return svc, nil
}

func (s *services) close(ctx context.Context) error {
	var errs []string

	if s.consumer != nil {
		if err := s.consumer.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if s.mongoClient != nil {
		if err := s.mongoClient.Close(ctx); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
