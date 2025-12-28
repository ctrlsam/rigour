package discovery

type DiscoveryConfig struct {
	ScanType string
	Ports    string
	TopPorts string
	Retries  int
	Rate     int
}

type Result struct {
	Host     string
	Port     int
	Protocol string
}
