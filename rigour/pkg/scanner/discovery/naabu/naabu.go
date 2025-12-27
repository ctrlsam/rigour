package naabu

import (
	"context"
	"fmt"
	"strings"

	"github.com/ctrlsam/rigour/pkg/scanner/discovery"
	"github.com/projectdiscovery/goflags"
	naabuResult "github.com/projectdiscovery/naabu/v2/pkg/result"
	naabuRunner "github.com/projectdiscovery/naabu/v2/pkg/runner"
)

// Run executes Naabu discovery for a single input target and invokes onResult
// for each open port found.
func Run(ctx context.Context, ipRange string, opts discovery.DiscoveryConfig, onResult func(discovery.Result)) error {
	if strings.TrimSpace(ipRange) == "" {
		return fmt.Errorf("naabu discovery input is empty")
	}

	naabuOpts := &naabuRunner.Options{
		Host: goflags.StringSlice{ipRange},
		// caller-configurable
		ScanType: opts.ScanType,
		Ports:    opts.Ports,
		TopPorts: opts.TopPorts,
		Rate:     opts.Rate,
		Retries:  opts.Retries,
		//Silent:            true,
	}

	naabuOpts.OnReceive = func(hr *naabuResult.HostResult) {
		for _, p := range hr.Ports {
			//fmt.Println("[DISCOVERY] Open port found:", hr.IP, p.Port)
			onResult(discovery.Result{
				Host:     hr.IP,
				Port:     p.Port,
				Protocol: "tcp",
			})
		}
	}

	r, err := naabuRunner.NewRunner(naabuOpts)
	if err != nil {
		return fmt.Errorf("naabu.NewRunner failed: %w", err)
	}
	defer r.Close()

	// Naabu runner is not fully context-aware; honour ctx by stopping early if canceled.
	done := make(chan error, 1)
	go func() {
		done <- r.RunEnumeration(ctx)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return fmt.Errorf("naabu enumeration failed: %w", err)
		}
		return nil
	}
}
