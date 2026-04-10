package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"nla/internal/model"
	"nla/internal/service"
)

type DetailsHandler struct {
	detailsSvc *service.DetailsService
	queueSvc   *service.QueueService
	bondSvc    *service.BondService
}

func NewDetailsHandler(detailsSvc *service.DetailsService, queueSvc *service.QueueService, bondSvc *service.BondService) *DetailsHandler {
	return &DetailsHandler{
		detailsSvc: detailsSvc,
		queueSvc:   queueSvc,
		bondSvc:    bondSvc,
	}
}

// GetDohodDetails GET /api/v1/bonds/{secid}/dohod — returns dohod.ru data for bond
func (h *DetailsHandler) GetDohodDetails(w http.ResponseWriter, r *http.Request) {
	secid := chi.URLParam(r, "secid")
	if secid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "secid required"})
		return
	}

	// Get bond to find ISIN
	bond, err := h.bondSvc.GetBondDetail(r.Context(), secid)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "bond not found"})
		return
	}

	isin := bond.ISIN
	if isin == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bond has no ISIN"})
		return
	}

	// Try cached data
	data, err := h.detailsSvc.GetDetails(r.Context(), secid, isin)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if data != nil {
		writeJSON(w, http.StatusOK, data)
		return
	}

	// No cache — enqueue async fetch
	jobData := map[string]any{
		"isin":  isin,
		"secid": secid,
	}
	job, err := h.queueSvc.Enqueue(r.Context(), model.JobTypeParseDohod, secid, jobData)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"job_id": job.JobID,
		"status": "pending",
	})
}
