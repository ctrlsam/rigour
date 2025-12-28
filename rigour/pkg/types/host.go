package types

import (
	"time"
)

type Host struct {
	ID        string    `json:"id" bson:"_id"`
	IP        string    `json:"ip"`
	IPInt     uint64    `json:"ip_int" bson:"ip_int"`
	ASN       *ASNInfo  `json:"asn,omitempty"`
	Location  *Location `json:"location,omitempty"`
	FirstSeen time.Time `json:"first_seen" bson:"first_seen"`
	LastSeen  time.Time `json:"last_seen" bson:"last_seen"`
	Services  []Service `json:"services,omitempty"`
	Labels    []string  `json:"labels,omitempty"`
}

type Location struct {
	Coordinates [2]float64 `json:"coordinates"`        // [longitude, latitude]
	City        string     `json:"city,omitempty"`     // City name
	Timezone    string     `json:"timezone,omitempty"` // IANA timezone identifier
}

type ASNInfo struct {
	Number              uint32 `json:"number"`                                                       // ASN number
	Organization        string `json:"organization,omitempty"`                                       // ISP/Organization name
	Country             string `json:"country,omitempty"`                                            // Country code (ISO 3166-1 alpha-2)
	IsSatelliteProvider bool   `json:"is_satellite_provider,omitempty" bson:"is_satellite_provider"` // Whether the ASN is a satellite internet provider
}
