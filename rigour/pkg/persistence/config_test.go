package persistence

import (
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid",
			cfg: Config{
				DbURI:        "mongodb://localhost:27017",
				GeoIPDataDir: "/data/geoip",
			},
			wantErr: false,
		},
		{
			name:    "missing db uri",
			cfg:     Config{GeoIPDataDir: "/data/geoip"},
			wantErr: true,
		},
		{
			name:    "missing geoip dir is allowed",
			cfg:     Config{DbURI: "mongodb://localhost:27017"},
			wantErr: false, // GeoIP is optional; enrichment is skipped when absent
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestConfigWithDefaults(t *testing.T) {
	cfg := Config{DbURI: "mongodb://localhost:27017", GeoIPDataDir: "/data"}.withDefaults()

	if cfg.DbName == "" {
		t.Error("DbName should have a default")
	}
	if cfg.DbCollection == "" {
		t.Error("DbCollection should have a default")
	}
	if cfg.DbTimeout <= 0 {
		t.Error("DbTimeout should have a positive default")
	}
	if cfg.DbTimeout != 10*time.Second {
		t.Errorf("DbTimeout default: got %v, want 10s", cfg.DbTimeout)
	}
}
