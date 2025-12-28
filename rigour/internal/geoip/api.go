package geoip

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// https://github.com/hibare/GoGeoIP

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

type GoGeoIPRecord struct {
	City                string
	Country             string
	Continent           string
	ISOCountryCode      string
	ISOContinentCode    string
	IsAnonymousProxy    bool
	IsSatelliteProvider bool
	Timezone            string
	Latitude            float64
	Longitude           float64
	ASN                 int64
	Organization        string
	IP                  string
}

func NewClient(baseURL, apiKey string, timeout time.Duration) (*Client, error) {
	if baseURL == "" || apiKey == "" {
		return nil, errors.New("geoip: baseURL and apiKey are required")
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http:    &http.Client{Timeout: timeout},
	}, nil
}

func (client *Client) Lookup(ctx context.Context, ip string) (*GoGeoIPRecord, error) {
	if ip == "" {
		return nil, errors.New("geoip: ip is required")
	}

	url := fmt.Sprintf("%s/api/v1/ip/%s", client.baseURL, ip)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", client.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, fmt.Errorf("geoip: status %d: %s", resp.StatusCode, body)
	}

	var rec GoGeoIPRecord
	if err := json.NewDecoder(resp.Body).Decode(&rec); err != nil {
		return nil, err
	}

	client.SetNullableFieldsToDefault(&rec)

	return &rec, nil
}

func (client *Client) SetNullableFieldsToDefault(rec *GoGeoIPRecord) {
	if rec.City == "" {
		rec.City = "Unknown"
	}
	if rec.Country == "" {
		rec.Country = "Unknown"
	}
	if rec.Continent == "" {
		rec.Continent = "Unknown"
	}
	if rec.ISOCountryCode == "" {
		rec.ISOCountryCode = "Unknown"
	}
	if rec.ISOContinentCode == "" {
		rec.ISOContinentCode = "Unknown"
	}
	if rec.Timezone == "" {
		rec.Timezone = "Unknown"
	}
	if rec.Organization == "" {
		rec.Organization = "Unknown"
	}
}
