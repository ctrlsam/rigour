package campaign

import (
	"fmt"
	"regexp"

	"github.com/ctrlsam/rigour/pkg/types"
)

// validPortsRegex accepts Naabu-style port strings: "80", "80,443", "80-1000", "80,443,8080-8090".
var validPortsRegex = regexp.MustCompile(`^(\d+(-\d+)?)(,\d+(-\d+)?)*$`)

// ValidateCampaign validates campaign fields before creation.
func ValidateCampaign(c *types.Campaign) error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(c.CIDRs) == 0 && c.CountryCode == "" {
		return fmt.Errorf("at least one cidr or a country_code is required")
	}
	if c.Ports != "" && c.TopPorts != "" {
		return fmt.Errorf("ports and top_ports are mutually exclusive")
	}
	if c.Ports != "" && !validPortsRegex.MatchString(c.Ports) {
		return fmt.Errorf("ports %q is invalid: use format like \"80\", \"80,443\", \"80-1000\", or \"80,443,8080-8090\"", c.Ports)
	}
	if c.TopPorts != "" && c.TopPorts != "100" && c.TopPorts != types.DefaultTopPorts && c.TopPorts != "full" {
		return fmt.Errorf("top_ports must be one of: 100, 1000, full")
	}
	if c.ScanType != "" && c.ScanType != types.ScanTypeConnect && c.ScanType != types.ScanTypeSyn {
		return fmt.Errorf("scan_type must be %q (connect) or %q (syn)", types.ScanTypeConnect, types.ScanTypeSyn)
	}
	if c.Rate > 15000 {
		return fmt.Errorf("rate %d exceeds the maximum of 15000 pps; use a lower value to avoid system instability", c.Rate)
	}
	return nil
}
