package fingerprint

import (
	"time"
)

type FingerprintConfig struct {
	// UDP scan
	UDP bool

	FastMode bool

	// The timeout specifies how long certain tasks should wait during the scanning process.
	// This may include the timeouts set on the handshake process and the time to wait for a response to return.
	DefaultTimeout time.Duration

	// Prints logging messages to stderr
	Verbose bool
}
