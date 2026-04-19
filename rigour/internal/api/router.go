package api

import (
	"context"

	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Router provides the Chi router configuration for the API.
type Router struct {
	router  *chi.Mux
	handler *Handler
}

// NewRouter creates a new API router.
func NewRouter(serverCtx context.Context, repository storage.HostRepository, campaignRepo storage.CampaignRepository, runner CampaignController, vulnRepo ...storage.VulnerabilityRepository) *Router {
	r := chi.NewRouter()

	// Add CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Add other middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	handler := NewHandler(repository)
	ch := NewCampaignHandler(serverCtx, campaignRepo, runner)

	var vh *VulnHandler
	if len(vulnRepo) > 0 && vulnRepo[0] != nil {
		vh = NewVulnHandler(vulnRepo[0])
	}

	// Register routes
	r.Get("/health", handler.HealthHandler)
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	r.Route("/api", func(r chi.Router) {
		r.Get("/hosts/search", handler.SearchHandler)
		r.Get("/hosts/{ip}", handler.GetHostHandler)
		r.Get("/facets", handler.FacetsHandler)

		if vh != nil {
			r.Get("/hosts/{ip}/vulnerabilities", vh.ListByIP)
			r.Get("/vulnerabilities/summary", vh.SummariseByTemplate)
			r.Get("/vulnerabilities", vh.ListByCampaign)
		}

		// Public campaign status endpoints
		r.Get("/campaigns", ch.ListCampaigns)
		r.Get("/campaigns/{id}", ch.GetCampaign)

		// Admin campaign management endpoints
		r.Route("/admin/campaigns", func(r chi.Router) {
			r.Post("/", ch.CreateCampaign)
			r.Get("/", ch.ListCampaigns)
			r.Get("/{id}", ch.GetCampaign)
			r.Delete("/{id}", ch.DeleteCampaign)
			r.Post("/{id}/start", ch.StartCampaign)
			r.Post("/{id}/cancel", ch.CancelCampaign)
			r.Post("/{id}/pause", ch.PauseCampaign)
			r.Post("/{id}/resume", ch.ResumeCampaign)
		})
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
