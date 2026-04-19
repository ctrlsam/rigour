package fingerprint

import (
	"strings"
	"time"
)

const (
	DefaultWorkers   = 50
	DefaultQueueSize = 500
)

type FingerprintConfig struct {
	UDP            bool
	FastMode       bool
	DefaultTimeout time.Duration
	Verbose        bool
	// PluginFilter limits fingerprinting to the named protocols (e.g. ["http","ssh"]).
	// Empty means all plugins run.
	PluginFilter []string
	// Workers is the number of concurrent fingerprinting goroutines. 0 = DefaultWorkers.
	Workers int
	// QueueSize is the capacity of the fingerprinting work queue. 0 = DefaultQueueSize.
	QueueSize int
}

func (c *FingerprintConfig) pluginAllowed(name string) bool {
	return c.PluginAllowed(name)
}

// PluginAllowed is the exported equivalent, used by callers outside this package.
func (c *FingerprintConfig) PluginAllowed(name string) bool {
	if len(c.PluginFilter) == 0 {
		return true
	}
	for _, f := range c.PluginFilter {
		if strings.EqualFold(f, name) {
			return true
		}
	}
	return false
}
