package persistence

import (
	"errors"
	"time"

	"github.com/ctrlsam/rigour/internal"
)

type Config struct {
	DbURI            string
	DbName           string
	DbCollection     string
	DbVulnCollection string
	DbTimeout        time.Duration

	GeoIPDataDir string // optional; enrichment skipped if empty

	VulnScanEnabled    bool
	NucleiTemplateDir  string
	NucleiConcurrency  int
	NucleiTemplateTags []string
}

func (c Config) withDefaults() Config {
	out := c
	if out.DbName == "" {
		out.DbName = internal.DatabaseName
	}
	if out.DbCollection == "" {
		out.DbCollection = internal.HostsRepositoryName
	}
	if out.DbVulnCollection == "" {
		out.DbVulnCollection = internal.VulnerabilitiesRepositoryName
	}
	if out.DbTimeout <= 0 {
		out.DbTimeout = 10 * time.Second
	}
	if out.NucleiConcurrency <= 0 {
		out.NucleiConcurrency = 5
	}
	if len(out.NucleiTemplateTags) == 0 {
		out.NucleiTemplateTags = []string{"cve", "rce", "default-login", "exposure", "misconfig", "unauth"}
	}
	return out
}

func (c Config) Validate() error {
	c = c.withDefaults()
	if c.DbURI == "" {
		return errors.New("persistence: db uri is required")
	}
	return nil
}
