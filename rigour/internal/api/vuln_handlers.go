package api

import (
	"net/http"

	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// VulnHandler handles vulnerability query endpoints.
type VulnHandler struct {
	repo storage.VulnerabilityRepository
}

// NewVulnHandler creates a new VulnHandler.
func NewVulnHandler(repo storage.VulnerabilityRepository) *VulnHandler {
	return &VulnHandler{repo: repo}
}

// ListByIP handles GET /api/hosts/{ip}/vulnerabilities.
func (h *VulnHandler) ListByIP(w http.ResponseWriter, r *http.Request) {
	ip := chi.URLParam(r, "ip")
	if ip == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "ip is required"})
		return
	}

	results, err := h.repo.ListByIP(r.Context(), ip)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	render.JSON(w, r, results)
}

// ListByCampaign handles GET /api/vulnerabilities?campaign_id=&severity=.
// When campaign_id is omitted, returns vulnerabilities across all campaigns.
func (h *VulnHandler) ListByCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := r.URL.Query().Get("campaign_id")
	severity := r.URL.Query().Get("severity")

	if campaignID == "" {
		results, err := h.repo.ListAll(r.Context(), severity)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": err.Error()})
			return
		}
		render.JSON(w, r, results)
		return
	}

	results, err := h.repo.ListByCampaign(r.Context(), campaignID, severity)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	render.JSON(w, r, results)
}

// SummariseByTemplate handles GET /api/vulnerabilities/summary.
// Returns vulnerabilities grouped by template, sorted by severity.
func (h *VulnHandler) SummariseByTemplate(w http.ResponseWriter, r *http.Request) {
	severity := r.URL.Query().Get("severity")

	results, err := h.repo.SummariseByTemplate(r.Context(), severity)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	render.JSON(w, r, results)
}
