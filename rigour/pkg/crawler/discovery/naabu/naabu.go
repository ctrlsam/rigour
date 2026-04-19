package naabu

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/discovery"
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
		Host:     goflags.StringSlice{ipRange},
		ScanType: opts.ScanType,
		Rate:     opts.Rate,
		Retries:  opts.Retries,
		Threads:  25,
	}
	// Ports and TopPorts are mutually exclusive in Naabu; setting both combines them.
	if opts.Ports != "" {
		naabuOpts.Ports = opts.Ports
	} else {
		naabuOpts.TopPorts = opts.TopPorts
	}

	naabuOpts.OnReceive = func(hr *naabuResult.HostResult) {
		for _, p := range hr.Ports {
			//fmt.Println("[DISCOVERY] Open port found:", hr.IP, p.Port)
			onResult(discovery.Result{
				Host:     hr.IP,
				Port:     p.Port,
				Protocol: p.Protocol.String(),
			})
		}
	}

	r, err := naabuRunner.NewRunner(naabuOpts)
	if err != nil {
		return fmt.Errorf("naabu.NewRunner failed: %w", err)
	}

	// Naabu has a race condition in its Scanner.Close() method: it nils
	// ListenHandler while background workers (TCPResultWorker, etc.) may
	// still be reading from ListenHandler's channels in a select. When Go
	// evaluates the select's channel expressions, it dereferences the nil
	// pointer and panics in naabu's own goroutine (unrecoverable from here).
	//
	// RunEnumeration creates an internal context and cancels it via defer when
	// it returns, which signals the workers to exit. We must not call Close()
	// until those workers have fully exited. A generous delay (30 s) ensures
	// all workers have drained any buffered packets and exited via ctx.Done().
	done := make(chan error, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				done <- fmt.Errorf("naabu panic: %v", p)
			}
		}()
		done <- r.RunEnumeration(ctx)
	}()

	select {
	case <-ctx.Done():
		deferClose(r)
		<-done
		return ctx.Err()
	case err := <-done:
		deferClose(r)
		if err != nil {
			return fmt.Errorf("naabu enumeration failed: %w", err)
		}
		return nil
	}
}

// deferClose schedules r.Close() after a 30-second delay to avoid a race with
// naabu's internal worker goroutines (see comment in Run). The closure also
// recovers from any panic in case the race is still hit on an edge case.
func deferClose(r *naabuRunner.Runner) {
	time.AfterFunc(30*time.Second, func() {
		defer func() { recover() }()
		r.Close()
	})
}
