package api

import (
	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Router provides the Chi router configuration for the API.
type Router struct {
	router  *chi.Mux
	handler *Handler
}

// NewRouter creates a new API router.
func NewRouter(repository storage.ServiceRepository) *Router {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	handler := NewHandler(repository)

	// Register routes
	r.Get("/health", handler.HealthHandler)

	r.Route("/api", func(r chi.Router) {
		r.Get("/hosts/search", handler.SearchHandler)
		r.Get("/facets", handler.FacetsHandler)
	})

	return &Router{
		router:  r,
		handler: handler,
	}
}

// Handler returns the underlying Chi router for use with http.ListenAndServe.
func (r *Router) Handler() *chi.Mux {
	return r.router
}
