package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ctrlsam/rigour/internal/storage"
	apimodels "github.com/ctrlsam/rigour/pkg/api"
	"github.com/go-chi/render"
)

// Handler provides HTTP handler methods for the API.
type Handler struct {
	queryHandler *QueryHandler
}

// NewHandler creates a new API handler.
func NewHandler(repository storage.HostRepository) *Handler {
	return &Handler{
		queryHandler: NewQueryHandler(repository),
	}
}

// SearchHandler handles GET /api/hosts/search requests with query parameters.
func (handler *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	// Parse filter from query parameter if provided
	filter := make(map[string]interface{})
	filterParam := r.URL.Query().Get("filter")
	if filterParam != "" {
		if err := json.Unmarshal([]byte(filterParam), &filter); err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "Invalid filter parameter"})
			return
		}
	}

	// Parse limit from query parameter
	limit := apimodels.DefaultPageSize
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		var parsed int
		if _, err := fmt.Sscanf(limitParam, "%d", &parsed); err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "Invalid limit parameter"})
			return
		}
		limit = apimodels.ValidatePageSize(parsed)
	}

	// Parse pagination token if provided
	var lastID string
	if pageToken := r.URL.Query().Get("page_token"); pageToken != "" {
		token, err := apimodels.DecodePaginationToken(pageToken)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "Invalid page token"})
			return
		}
		if token != nil {
			lastID = token.LastID
		}
	}

	// Execute search
	hosts, nextID, err := handler.queryHandler.Search(r.Context(), filter, lastID, limit)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Search failed: " + err.Error()})
		return
	}

	// Build response
	resp := apimodels.SearchResponse{
		Hosts: hosts,
	}

	// Generate next page token if there are more results
	if nextID != "" {
		token, err := apimodels.EncodePaginationToken(nextID, "_id")
		if err == nil {
			resp.NextPageToken = token
		}
	}

	// Return response
	render.JSON(w, r, resp)
}

// GetHostHandler handles GET /api/hosts/{ip} requests.
func (handler *Handler) GetHostHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.PathValue("ip")
	if ip == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "IP address is required"})
		return
	}

	host, err := handler.queryHandler.GetByIP(r.Context(), ip)
	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	render.JSON(w, r, host)
}

// FacetsHandler handles GET /api/facets requests.
func (handler *Handler) FacetsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse filter from query parameter if provided
	filter := make(map[string]interface{})
	filterParam := r.URL.Query().Get("filter")
	if filterParam != "" {
		if err := json.Unmarshal([]byte(filterParam), &filter); err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "Invalid filter parameter"})
			return
		}
	}

	// Execute facet aggregation
	agg, err := handler.queryHandler.Facets(r.Context(), filter)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Facet aggregation failed: " + err.Error()})
		return
	}

	// Build response
	resp := apimodels.FacetResponse{
		Facets: apimodels.FacetCounts{
			Services:  agg.Services,
			Countries: agg.Countries,
			ASNs:      agg.ASNs,
		},
	}

	// Return response
	render.JSON(w, r, resp)
}

// HealthHandler handles GET /health requests.
func (handler *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"status": "ok"})
}
