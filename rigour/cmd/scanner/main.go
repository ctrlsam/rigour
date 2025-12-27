package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/user"
	"runtime"
	"time"

	"github.com/ctrlsam/rigour/internal/messaging/kafka"
	"github.com/ctrlsam/rigour/pkg/scanner"
	"github.com/ctrlsam/rigour/pkg/scanner/discovery"
	"github.com/ctrlsam/rigour/pkg/scanner/fingerprint"
	"github.com/spf13/cobra"
)

type cliConfig struct {
	fastMode bool
	timeout  int
	useUDP   bool
	verbose  bool

	// Kafka output
	kafkaBrokers string
	kafkaTopic   string

	// Discovery settings
	scanType string
	ports    string
	topPorts string
	retries  int
	rate     int
}

var (
	config  cliConfig
	rootCmd = &cobra.Command{
		Use: "rigour [flags]\nTARGET SPECIFICATION:\n\tRequires an ip address or CIDR range\n" +
			"EXAMPLES:\n\trigour 192.168.1.0/24\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			configErr := checkConfig(config)
			if configErr != nil {
				return configErr
			}

			cidrRange := args[0]
			ipCount := getCIDRRangeSize(cidrRange)
			fmt.Printf("Starting scan of %d IPs in range %s\n", ipCount, cidrRange)

			// Initialize Kafka producer
			producer, err := kafka.NewProducer(kafka.ProducerConfig{
				Brokers: config.kafkaBrokers,
				Topic:   config.kafkaTopic,
			})
			if err != nil {
				return err
			}
			defer func() { _ = producer.Close() }()

			onEvent := func(ev scanner.ScannedServiceEvent) {
				// Encode once and reuse for both outputs.
				serializedEvent, err := json.Marshal(ev)
				if err != nil {
					// Streaming should never abort the whole scan due to a single marshal failure.
					fmt.Fprintf(os.Stderr, "failed to marshal event: %v\n", err)
					return
				}

				// Key by IP:port for stable partitioning.
				key := []byte(fmt.Sprintf("%s:%d", ev.IP, ev.Port))
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				err = producer.PublishBytes(ctx, key, serializedEvent)
				cancel()
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to publish kafka event: %v\n", err)
				}

				// NDJSON output.
				_, _ = os.Stdout.Write(append(serializedEvent, '\n'))
			}

			err = scanner.ScanTargetWithDiscoveryStream(cidrRange, createDiscoveryConfig(config), createScanConfig(config), onEvent)
			if err != nil {
				return fmt.Errorf("Failed running discovery+scan stream (%w)", err)
			}
			return nil
		},
	}
)

func checkConfig(config cliConfig) error {
	if config.useUDP && config.verbose {
		user, err := user.Current()
		if err != nil {
			return fmt.Errorf("Failed to retrieve current user (error: %w)", err)
		}
		if !((runtime.GOOS == "linux" || runtime.GOOS == "darwin") && user.Uid == "0") {
			fmt.Fprintln(os.Stderr, "Note: UDP Scan may require root privileges")
		}
	}

	return nil
}

func getCIDRRangeSize(cidr string) int {
	_, ipnet, _ := net.ParseCIDR(cidr)
	ones, bits := ipnet.Mask.Size()
	numIPs := 1 << (bits - ones)
	return numIPs
}

func createScanConfig(config cliConfig) fingerprint.FingerprintConfig {
	return fingerprint.FingerprintConfig{
		DefaultTimeout: time.Duration(config.timeout) * time.Millisecond,
		FastMode:       config.fastMode,
		UDP:            config.useUDP,
		Verbose:        config.verbose,
	}
}

func createDiscoveryConfig(config cliConfig) discovery.DiscoveryConfig {
	return discovery.DiscoveryConfig{
		ScanType: config.scanType,
		Ports:    config.ports,
		TopPorts: config.topPorts,
		Retries:  config.retries,
		Rate:     config.rate,
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	rootCmd.PersistentFlags().BoolVarP(&config.fastMode, "fast", "f", false, "fast mode")
	rootCmd.PersistentFlags().
		BoolVarP(&config.useUDP, "udp", "U", false, "run UDP plugins")

	rootCmd.PersistentFlags().BoolVarP(&config.verbose, "verbose", "v", false, "verbose mode")
	rootCmd.PersistentFlags().
		IntVarP(&config.timeout, "timeout", "w", 2000, "timeout (milliseconds)")

	// Kafka output flags
	rootCmd.PersistentFlags().StringVar(&config.kafkaBrokers, "kafka-brokers", "localhost:29092", "Kafka brokers (comma-separated host:port)")
	rootCmd.PersistentFlags().StringVar(&config.kafkaTopic, "kafka-topic", "rigour.scanner.service", "Kafka topic")

	// Discovery flags - These control how rigour discovers open ports.
	rootCmd.PersistentFlags().StringVar(&config.scanType, "scan-type", "c", "discovery scan type (naabu; e.g. c=connect)")
	rootCmd.PersistentFlags().StringVar(&config.ports, "ports", "", "ports list (e.g. 80,443). If set, overrides top ports")
	rootCmd.PersistentFlags().StringVar(&config.topPorts, "top-ports", "100", "top ports (e.g. 100, 1000, full)") // full
	rootCmd.PersistentFlags().IntVar(&config.retries, "retries", 3, "discovery retries")
	rootCmd.PersistentFlags().IntVar(&config.rate, "rate", 50_000, "discovery rate (packets per second)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
