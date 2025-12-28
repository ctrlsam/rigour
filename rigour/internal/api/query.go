package api

import (
	"context"

	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/ctrlsam/rigour/pkg/types"
)

// QueryHandler provides methods for executing queries using the storage abstraction.
type QueryHandler struct {
	repository storage.HostRepository
}

// NewQueryHandler creates a new QueryHandler.
func NewQueryHandler(repository storage.HostRepository) *QueryHandler {
	return &QueryHandler{
		repository: repository,
	}
}

// Search queries hosts with filter and pagination support.
func (qh *QueryHandler) Search(ctx context.Context, filter map[string]interface{}, lastID string, limit int) ([]types.Host, string, error) {
	return qh.repository.Search(ctx, filter, lastID, limit)
}

// GetByIP retrieves a single host by IP address.
func (qh *QueryHandler) GetByIP(ctx context.Context, ip string) (*types.Host, error) {
	return qh.repository.GetByIP(ctx, ip)
}

// Facets performs aggregation for facet counts.
func (qh *QueryHandler) Facets(ctx context.Context, filter map[string]interface{}) (*storage.FacetCounts, error) {
	return qh.repository.Facets(ctx, filter)
}
