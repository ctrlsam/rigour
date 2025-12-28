package api

import "github.com/ctrlsam/rigour/pkg/types"

// SearchRequest represents a search query request.
type SearchRequest struct {
	Filter    map[string]interface{} `json:"filter,omitempty"`
	PageToken string                 `json:"page_token,omitempty"`
	Limit     int                    `json:"limit,omitempty"`
}

// SearchResponse represents the response for a search query.
type SearchResponse struct {
	Hosts         []types.Host `json:"hosts"`
	Facets        *FacetCounts `json:"facets,omitempty"`
	NextPageToken string       `json:"next_page_token,omitempty"`
}

// FacetCounts represents aggregated counts for various facets.
type FacetCounts struct {
	Services  map[string]int `json:"services,omitempty"`
	Countries map[string]int `json:"countries,omitempty"`
	ASNs      map[string]int `json:"asns,omitempty"`
}

// FacetRequest represents a request for facet aggregation.
type FacetRequest struct {
	Filter map[string]interface{} `json:"filter,omitempty"`
}

// FacetResponse represents the response for a facet aggregation query.
type FacetResponse struct {
	Facets FacetCounts `json:"facets"`
}

// PaginationToken represents the pagination cursor token structure.
type PaginationToken struct {
	LastID    string `json:"last_id"`
	SortField string `json:"sort_field"`
}
