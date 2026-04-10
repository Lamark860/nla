package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"nla/internal/model"
	"nla/internal/service"
)

type AnalysisHandler struct {
	analysisSvc *service.AnalysisService
	queueSvc    *service.QueueService
}

func NewAnalysisHandler(analysisSvc *service.AnalysisService, queueSvc *service.QueueService) *AnalysisHandler {
	return &AnalysisHandler{
		analysisSvc: analysisSvc,
		queueSvc:    queueSvc,
	}
}

// StartAnalysis POST /api/v1/bonds/{secid}/analyze — enqueue AI analysis job
func (h *AnalysisHandler) StartAnalysis(w http.ResponseWriter, r *http.Request) {
	secid := chi.URLParam(r, "secid")
	if secid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "secid is required"})
		return
	}

	var req model.AnalyzeRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
	}

	jobData := map[string]any{}
	if req.JSONData != nil {
		jobData["json_data"] = req.JSONData
	}
	if req.CustomJSON != nil {
		jobData["custom_json"] = req.CustomJSON
	}

	job, err := h.queueSvc.Enqueue(r.Context(), model.JobTypeAIAnalysis, secid, jobData)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"job_id": job.JobID,
		"status": job.Status,
	})
}

// GetAnalyses GET /api/v1/bonds/{secid}/analyses — list analyses for a bond
func (h *AnalysisHandler) GetAnalyses(w http.ResponseWriter, r *http.Request) {
	secid := chi.URLParam(r, "secid")
	if secid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "secid is required"})
		return
	}

	analyses, err := h.analysisSvc.GetBySecid(r.Context(), secid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, analyses)
}

// GetAnalysis GET /api/v1/analyses/{id} — get single analysis
func (h *AnalysisHandler) GetAnalysis(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	analysis, err := h.analysisSvc.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, analysis)
}

// DeleteAnalysis DELETE /api/v1/analyses/{id} — delete analysis
func (h *AnalysisHandler) DeleteAnalysis(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	if err := h.analysisSvc.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// GetAnalysisStats GET /api/v1/bonds/{secid}/analysis-stats — aggregate stats
func (h *AnalysisHandler) GetAnalysisStats(w http.ResponseWriter, r *http.Request) {
	secid := chi.URLParam(r, "secid")
	if secid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "secid is required"})
		return
	}

	stats, err := h.analysisSvc.GetStats(r.Context(), secid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// GetBulkAnalysisStats GET /api/v1/analyses/bulk-stats — stats for all analyzed bonds
func (h *AnalysisHandler) GetBulkAnalysisStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.analysisSvc.GetBulkStats(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// GetJobStatus GET /api/v1/jobs/{id} — poll job status
func (h *AnalysisHandler) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	status, err := h.queueSvc.GetStatus(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, status)
}

// GetQueueStats GET /api/v1/queue/stats — queue statistics
func (h *AnalysisHandler) GetQueueStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.queueSvc.GetStats(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
