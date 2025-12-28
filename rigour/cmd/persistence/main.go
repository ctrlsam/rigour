package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	internalconst "github.com/ctrlsam/rigour/internal"
	"github.com/ctrlsam/rigour/pkg/persistence"
	"github.com/spf13/cobra"
)

type cliConfig struct {
	brokers    string
	groupID    string
	topic      string
	mongoURI   string
	database   string
	collection string
	geoipURL   string
	geoipKey   string
}

func main() {
	var cfg cliConfig

	root := &cobra.Command{
		Use:   "rigour-persistence",
		Short: "Consume crawler service events and persist/enrich hosts in MongoDB",
		RunE: func(cmd *cobra.Command, args []string) error {
			appCfg := persistence.Config{
				KafkaBrokers:    cfg.brokers,
				KafkaGroupID:    cfg.groupID,
				Topic:           cfg.topic,
				MongoURI:        cfg.mongoURI,
				MongoDatabase:   cfg.database,
				MongoCollection: cfg.collection,
				MongoTimeout:    10 * time.Second,
				GeoIPBaseURL:    cfg.geoipURL,
				GeoIPAPIKey:     cfg.geoipKey,
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Shutdown on SIGINT/SIGTERM
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigCh
				cancel()
			}()

			app, err := persistence.NewApp(ctx, appCfg)
			if err != nil {
				return err
			}
			defer func() { _ = app.Close(context.Background()) }()

			err = app.Run(ctx)
			if err == context.Canceled {
				return nil
			}
			return err
		},
	}

	root.Flags().StringVar(&cfg.brokers, "brokers", "localhost:29092", "Kafka brokers (comma-separated)")
	root.Flags().StringVar(&cfg.groupID, "group", "rigour-persistence", "Kafka consumer group id")
	root.Flags().StringVar(&cfg.topic, "topic", internalconst.KafkaTopicScannedServices, "Kafka topic to consume")

	root.Flags().StringVar(&cfg.mongoURI, "mongo-uri", "mongodb://localhost:27017", "MongoDB connection URI")
	root.Flags().StringVar(&cfg.database, "mongo-db", internalconst.DatabaseName, "MongoDB database name")
	root.Flags().StringVar(&cfg.collection, "mongo-coll", internalconst.HostsRepositoryName, "MongoDB hosts collection name")

	root.Flags().StringVar(&cfg.geoipURL, "geoip-url", "http://localhost:5000", "GoGeoIP base URL")
	root.Flags().StringVar(&cfg.geoipKey, "geoip-key", "mykey", "GoGeoIP api key")

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
