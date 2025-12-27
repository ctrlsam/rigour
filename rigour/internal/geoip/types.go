package geoip

// GeoIPRecord represents a GeoIP lookup result.

type GeoIPRecord struct {
	City                string  `json:"city" bson:"city,omitempty"`
	Country             string  `json:"country" bson:"country,omitempty"`
	Continent           string  `json:"continent" bson:"continent,omitempty"`
	ISOCountryCode      string  `json:"iso_country_code" bson:"iso_country_code,omitempty"`
	ISOContinentCode    string  `json:"iso_continent_code" bson:"iso_continent_code,omitempty"`
	IsAnonymousProxy    bool    `json:"is_anonymous_proxy" bson:"is_anonymous_proxy,omitempty"`
	IsSatelliteProvider bool    `json:"is_satellite_provider" bson:"is_satellite_provider,omitempty"`
	Timezone            string  `json:"timezone" bson:"timezone,omitempty"`
	Latitude            float64 `json:"latitude" bson:"latitude,omitempty"`
	Longitude           float64 `json:"longitude" bson:"longitude,omitempty"`
	ASN                 int64   `json:"asn" bson:"asn,omitempty"`
	Organization        string  `json:"organization" bson:"organization,omitempty"`
	IP                  string  `json:"ip" bson:"ip,omitempty"`
}

type GeoIPDatabase interface {
	Lookup(ip string) (*GeoIPRecord, error)
}
