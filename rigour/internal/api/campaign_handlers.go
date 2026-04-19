package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/ctrlsam/rigour/pkg/campaign"
	"github.com/ctrlsam/rigour/pkg/types"
	"github.com/go-chi/render"
)

// CampaignHandler provides HTTP handlers for campaign management.
type CampaignHandler struct {
	repo      storage.CampaignRepository
	runner    CampaignController
	serverCtx context.Context
}

func NewCampaignHandler(serverCtx context.Context, repo storage.CampaignRepository, runner CampaignController) *CampaignHandler {
	return &CampaignHandler{repo: repo, runner: runner, serverCtx: serverCtx}
}

// CreateCampaign handles POST /api/admin/campaigns.
func (h *CampaignHandler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	var c types.Campaign
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		jsonError(w, r, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if err := campaign.ValidateCampaign(&c); err != nil {
		jsonError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.repo.CreateCampaign(r.Context(), &c)
	if err != nil {
		jsonError(w, r, http.StatusInternalServerError, "create campaign: "+err.Error())
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, created)
}

// ListCampaigns handles GET /api/admin/campaigns and GET /api/campaigns.
func (h *CampaignHandler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	filter := storage.CampaignFilter{}
	if s := r.URL.Query().Get("status"); s != "" {
		st := types.CampaignStatus(s)
		if _, ok := types.ValidCampaignStatuses[st]; !ok {
			jsonError(w, r, http.StatusBadRequest, "invalid status filter")
			return
		}
		filter.Status = st
	}

	campaigns, err := h.repo.ListCampaigns(r.Context(), filter)
	if err != nil {
		jsonError(w, r, http.StatusInternalServerError, "list campaigns: "+err.Error())
		return
	}
	if campaigns == nil {
		campaigns = []types.Campaign{}
	}
	render.JSON(w, r, campaigns)
}

// GetCampaign handles GET /api/admin/campaigns/{id} and GET /api/campaigns/{id}.
func (h *CampaignHandler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, r, http.StatusBadRequest, "id is required")
		return
	}
	c, err := h.repo.GetCampaign(r.Context(), id)
	if err != nil {
		jsonError(w, r, http.StatusNotFound, err.Error())
		return
	}
	render.JSON(w, r, c)
}

// DeleteCampaign handles DELETE /api/admin/campaigns/{id}.
func (h *CampaignHandler) DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, r, http.StatusBadRequest, "id is required")
		return
	}
	c, err := h.repo.GetCampaign(r.Context(), id)
	if err != nil {
		jsonError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if c.Status == types.CampaignRunning {
		jsonError(w, r, http.StatusConflict, "cancel the campaign before deleting it")
		return
	}
	if err := h.repo.DeleteCampaign(r.Context(), id); err != nil {
		jsonError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// StartCampaign handles POST /api/admin/campaigns/{id}/start.
func (h *CampaignHandler) StartCampaign(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, err := h.repo.GetCampaign(r.Context(), id)
	if err != nil {
		jsonError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if c.Status == types.CampaignRunning {
		jsonError(w, r, http.StatusConflict, "campaign is already running")
		return
	}
	if c.Status == types.CampaignCompleted {
		jsonError(w, r, http.StatusConflict, "campaign has already completed")
		return
	}
	if h.runner.IsActive(id) {
		jsonError(w, r, http.StatusConflict, "campaign is already active")
		return
	}
	// Update status synchronously so the next read sees "running" immediately,
	// before the goroutine has a chance to do it inside execute().
	if err := h.repo.UpdateCampaignStatus(r.Context(), id, types.CampaignRunning, ""); err != nil {
		jsonError(w, r, http.StatusInternalServerError, "update status: "+err.Error())
		return
	}
	go h.runner.Start(h.serverCtx, c)
	render.JSON(w, r, map[string]string{"status": "started"})
}

// CancelCampaign handles POST /api/admin/campaigns/{id}/cancel.
func (h *CampaignHandler) CancelCampaign(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.runner.Cancel(id); err != nil {
		jsonError(w, r, http.StatusConflict, err.Error())
		return
	}
	render.JSON(w, r, map[string]string{"status": "cancelled"})
}

// PauseCampaign handles POST /api/admin/campaigns/{id}/pause.
func (h *CampaignHandler) PauseCampaign(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.runner.Pause(id); err != nil {
		jsonError(w, r, http.StatusConflict, err.Error())
		return
	}
	render.JSON(w, r, map[string]string{"status": "paused"})
}

// ResumeCampaign handles POST /api/admin/campaigns/{id}/resume.
func (h *CampaignHandler) ResumeCampaign(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, err := h.repo.GetCampaign(r.Context(), id)
	if err != nil {
		jsonError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if c.Status != types.CampaignPaused {
		jsonError(w, r, http.StatusConflict, "campaign is not paused")
		return
	}
	if h.runner.IsActive(id) {
		jsonError(w, r, http.StatusConflict, "campaign is already active")
		return
	}
	if err := h.repo.UpdateCampaignStatus(r.Context(), id, types.CampaignRunning, ""); err != nil {
		jsonError(w, r, http.StatusInternalServerError, "update status: "+err.Error())
		return
	}
	go h.runner.Start(h.serverCtx, c)
	render.JSON(w, r, map[string]string{"status": "resumed"})
}

func jsonError(w http.ResponseWriter, r *http.Request, code int, msg string) {
	render.Status(r, code)
	render.JSON(w, r, map[string]string{"error": msg})
}
