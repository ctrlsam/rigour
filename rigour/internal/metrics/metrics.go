package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Port discovery
	PortsDiscovered = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rigour_ports_discovered_total",
		Help: "Total open ports found by port discovery.",
	})

	// Fingerprinting outcomes
	ServicesIdentified = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rigour_services_identified_total",
		Help: "Total services with a matched protocol, labelled by protocol name.",
	}, []string{"protocol"})

	ServicesUnknown = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rigour_services_unknown_total",
		Help: "Total open ports where protocol could not be identified.",
	})

	// Worker pool internals
	FingerprintQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rigour_fingerprint_queue_depth",
		Help: "Number of ports currently queued for fingerprinting.",
	})

	FingerprintActiveWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rigour_fingerprint_active_workers",
		Help: "Number of fingerprint workers currently busy.",
	})

	FingerprintInlineTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rigour_fingerprint_inline_total",
		Help: "Ports fingerprinted inline because the queue was full.",
	})

	// Campaign-level
	CampaignsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rigour_campaigns_active",
		Help: "Number of campaigns currently executing.",
	})

	CampaignServicesFound = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rigour_campaign_services_found_total",
		Help: "Total services published across all campaigns.",
	})
)
