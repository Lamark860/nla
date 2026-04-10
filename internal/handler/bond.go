package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"nla/internal/service"
)

type BondHandler struct {
	bondService *service.BondService
}

func NewBondHandler(bondService *service.BondService) *BondHandler {
	return &BondHandler{bondService: bondService}
}

// ListBonds GET /api/v1/bonds
func (h *BondHandler) ListBonds(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "best"
	}

	result, err := h.bondService.GetBondsPaginated(r.Context(), page, perPage, sortBy)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetBond GET /api/v1/bonds/{secid}
func (h *BondHandler) GetBond(w http.ResponseWriter, r *http.Request) {
	secid := chi.URLParam(r, "secid")
	if secid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "secid is required"})
		return
	}

	bond, err := h.bondService.GetBondDetail(r.Context(), secid)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, bond)
}

// GetBondCoupons GET /api/v1/bonds/{secid}/coupons
func (h *BondHandler) GetBondCoupons(w http.ResponseWriter, r *http.Request) {
	secid := chi.URLParam(r, "secid")
	if secid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "secid is required"})
		return
	}

	coupons, err := h.bondService.GetBondCoupons(r.Context(), secid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, coupons)
}

// GetBondHistory GET /api/v1/bonds/{secid}/history
func (h *BondHandler) GetBondHistory(w http.ResponseWriter, r *http.Request) {
	secid := chi.URLParam(r, "secid")
	if secid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "secid is required"})
		return
	}

	history, err := h.bondService.GetBondHistory(r.Context(), secid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, history)
}

// GetMonthlyBonds GET /api/v1/bonds/monthly
func (h *BondHandler) GetMonthlyBonds(w http.ResponseWriter, r *http.Request) {
	bonds, err := h.bondService.GetMonthlyBonds(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, bonds)
}

// GetBondsGrouped GET /api/v1/bonds/grouped
func (h *BondHandler) GetBondsGrouped(w http.ResponseWriter, r *http.Request) {
	monthlyOnly := r.URL.Query().Get("monthly") == "true"

	result, err := h.bondService.GetBondsGroupedByIssuer(r.Context(), monthlyOnly)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ClearCache POST /api/v1/bonds/clear-cache — invalidate bond cache
func (h *BondHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	count, err := h.bondService.ClearCache(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "cleared": count})
}

// ToggleIssuer POST /api/v1/issuers/{id}/toggle — hide/show issuer
func (h *BondHandler) ToggleIssuer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	emitterID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || emitterID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid emitter id"})
		return
	}

	var req struct {
		Hidden bool `json:"hidden"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	count, err := h.bondService.ToggleIssuer(r.Context(), emitterID, req.Hidden)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"emitter_id": emitterID, "hidden": req.Hidden, "affected": count})
}
