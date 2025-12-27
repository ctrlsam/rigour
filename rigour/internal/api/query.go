package api

import (
	"context"

	"github.com/ctrlsam/rigour/internal/storage"
)

// QueryHandler provides methods for executing queries using the storage abstraction.
type QueryHandler struct {
	repository storage.ServiceRepository
}

// NewQueryHandler creates a new QueryHandler.
func NewQueryHandler(repository storage.ServiceRepository) *QueryHandler {
	return &QueryHandler{
		repository: repository,
	}
}

// Search queries hosts with filter and pagination support.
func (qh *QueryHandler) Search(ctx context.Context, filter map[string]interface{}, lastID string, limit int) ([]storage.Host, string, error) {
	return qh.repository.Search(ctx, filter, lastID, limit)
}

// Facets performs aggregation for facet counts.
func (qh *QueryHandler) Facets(ctx context.Context, filter map[string]interface{}) (*storage.FacetCounts, error) {
	return qh.repository.Facets(ctx, filter)
}
