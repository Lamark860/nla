package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"nla/internal/model"
	"nla/internal/scoring"
	"nla/internal/service"
)

// ScoringHandler exposes Phase 2 endpoints — the analytical-index (балл
// 0..100) per weight profile plus the LLM-generated explanation pipeline.
type ScoringHandler struct {
	scoringSvc *service.ScoringService
	queueSvc   *service.QueueService
}

func NewScoringHandler(scoringSvc *service.ScoringService, queueSvc *service.QueueService) *ScoringHandler {
	return &ScoringHandler{scoringSvc: scoringSvc, queueSvc: queueSvc}
}

// ListProfiles GET /api/v1/scoring/profiles
// Returns the available scoring profiles (currently only presets).
func (h *ScoringHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.scoringSvc.ListProfiles(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, profiles)
}

// GetScore GET /api/v1/bonds/{secid}/score[?profile=low|mid|high]
//
// No profile param  → returns all three profiles as an array.
// With profile param → returns the single profile as an object.
//
// Both shapes share ScoreResponse so the frontend's existing «AiScore»
// component can render either by mapping `profile_code` → display.
func (h *ScoringHandler) GetScore(w http.ResponseWriter, r *http.Request) {
	secid := chi.URLParam(r, "secid")
	if secid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "secid is required"})
		return
	}

	profileCode := strings.ToLower(r.URL.Query().Get("profile"))
	if profileCode == "" {
		results, err := h.scoringSvc.ComputeAll(r.Context(), secid)
		if err != nil {
			writeJSON(w, mapScoreErr(err), map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, results)
		return
	}

	res, err := h.scoringSvc.ComputeOne(r.Context(), secid, profileCode)
	if err != nil {
		writeJSON(w, mapScoreErr(err), map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// ExplainScore POST /api/v1/bonds/{secid}/score/explain?profile=X
//
// Async: enqueues a JobTypeScoreExplain that the queue worker will pick
// up and call ScoringService.Explain on. Returns the job_id immediately;
// the frontend polls /jobs/{id} as it already does for analyze.
//
// We resolve the latest score row up front and put its ID into the job
// payload so the worker narrates a frozen breakdown — a 24h cache refresh
// that lands between enqueue and execution can't switch out the data
// under the LLM mid-flight.
func (h *ScoringHandler) ExplainScore(w http.ResponseWriter, r *http.Request) {
	secid := chi.URLParam(r, "secid")
	if secid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "secid is required"})
		return
	}
	profileCode := strings.ToLower(r.URL.Query().Get("profile"))
	if profileCode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "profile query param is required"})
		return
	}

	// Triggering compute guarantees a fresh BondScore row exists, and
	// gives us its ID for the job payload.
	score, err := h.scoringSvc.ComputeOne(r.Context(), secid, profileCode)
	if err != nil {
		writeJSON(w, mapScoreErr(err), map[string]string{"error": err.Error()})
		return
	}

	jobData, err := json.Marshal(model.ScoreExplainJobData{
		Profile: profileCode,
		ScoreID: score.ScoreID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "encode job data: " + err.Error()})
		return
	}
	// queueSvc.Enqueue accepts the data as any — keep it as a map for
	// the existing pgx persistence path, which serialises map[string]any
	// straight to JSONB.
	var asMap map[string]any
	_ = json.Unmarshal(jobData, &asMap)

	job, err := h.queueSvc.Enqueue(r.Context(), model.JobTypeScoreExplain, secid, asMap)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"job_id":   job.JobID,
		"status":   job.Status,
		"score_id": score.ScoreID,
	})
}

// mapScoreErr turns service-layer sentinels into the HTTP status we want
// the frontend to surface. Anything we don't recognise becomes a 500 —
// it's almost certainly a DB or upstream MOEX failure worth a louder
// signal than 4xx.
func mapScoreErr(err error) int {
	switch {
	case errors.Is(err, service.ErrUnknownProfile):
		return http.StatusNotFound
	case err != nil && strings.Contains(err.Error(), "not found"):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// Compile-time check that breakdown ordering matches the engine's canonical
// list. If a new factor lands in scoring/engine.go but isn't yet wired into
// the model layer, this will fail to compile rather than silently shipping
// half-baked breakdown JSON.
var _ = scoring.AllFactors
