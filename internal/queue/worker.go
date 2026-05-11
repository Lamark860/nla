package queue

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"nla/internal/model"
	"nla/internal/service"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Worker polls Postgres for pending jobs and processes them.
type Worker struct {
	queueSvc     *service.QueueService
	analysisSvc  *service.AnalysisService
	bondSvc      *service.BondService
	detailsSvc   *service.DetailsService
	scoringSvc   *service.ScoringService
	pollInterval time.Duration
}

func NewWorker(
	queueSvc *service.QueueService,
	analysisSvc *service.AnalysisService,
	bondSvc *service.BondService,
	detailsSvc *service.DetailsService,
	scoringSvc *service.ScoringService,
) *Worker {
	return &Worker{
		queueSvc:     queueSvc,
		analysisSvc:  analysisSvc,
		bondSvc:      bondSvc,
		detailsSvc:   detailsSvc,
		scoringSvc:   scoringSvc,
		pollInterval: 2 * time.Second,
	}
}

// Run starts the worker loop. Blocks until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	// Reset jobs stuck in "running" from a previous crash/restart
	if n, err := w.queueSvc.ResetStaleJobs(ctx, 3*time.Minute); err != nil {
		log.Printf("WARN: reset stale jobs: %v", err)
	} else if n > 0 {
		log.Printf("Reset %d stale running jobs back to pending", n)
	}

	log.Println("Queue worker started")
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Queue worker stopped")
			return
		case <-ticker.C:
			w.processNext(ctx)
		}
	}
}

func (w *Worker) processNext(ctx context.Context) {
	job, err := w.queueSvc.FetchPending(ctx)
	if err != nil {
		log.Printf("ERROR: fetch pending job: %v", err)
		return
	}
	if job == nil {
		return // no pending jobs
	}

	log.Printf("Processing job %s (type=%s, secid=%s)", job.JobID, job.Type, job.SECID)

	switch job.Type {
	case model.JobTypeAIAnalysis:
		w.processAIAnalysis(ctx, job)
	case model.JobTypeParseDohod:
		w.processDohodParse(ctx, job)
	case model.JobTypeScoreExplain:
		w.processScoreExplain(ctx, job)
	default:
		log.Printf("WARN: unknown job type %q, marking error", job.Type)
		w.queueSvc.MarkError(ctx, job.JobID, "unknown job type: "+job.Type)
	}
}

// processScoreExplain narrates a previously-computed BondScore via the LLM
// and stashes the text in bond_score_explanations. The job payload always
// carries the frozen score_id (handler resolved it before enqueue), so a
// concurrent 24h cache refresh between enqueue and pick-up cannot swap
// out the breakdown mid-explanation.
func (w *Worker) processScoreExplain(ctx context.Context, job *model.QueueJob) {
	scoreID, ok := extractScoreID(job.Data)
	if !ok {
		log.Printf("ERROR: missing score_id in score_explain job %s (data type %T)", job.JobID, job.Data)
		w.queueSvc.MarkError(ctx, job.JobID, "missing score_id in job data")
		return
	}

	exp, err := w.scoringSvc.Explain(ctx, scoreID)
	if err != nil {
		log.Printf("ERROR: score explain for job %s (score_id=%d): %v", job.JobID, scoreID, err)
		w.queueSvc.MarkError(ctx, job.JobID, "score explain: "+err.Error())
		return
	}

	result := map[string]any{
		"score_id":       scoreID,
		"explanation_id": exp.ID,
	}
	if err := w.queueSvc.MarkDone(ctx, job.JobID, result); err != nil {
		log.Printf("ERROR: mark score_explain done %s: %v", job.JobID, err)
	}
	log.Printf("Job %s done: explanation %d for score %d", job.JobID, exp.ID, scoreID)
}

// extractScoreID pulls score_id from the job's data blob, regardless of
// whether pgx delivered it as map[string]any (the normal pg-driven path)
// or as a number that came in via JSON encoder somewhere else.
// Returns ok=false when nothing parseable is present.
func extractScoreID(data any) (int64, bool) {
	m, ok := data.(map[string]any)
	if !ok {
		return 0, false
	}
	v, ok := m["score_id"]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64: // JSON numbers always come back as float64
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	}
	return 0, false
}

func (w *Worker) processAIAnalysis(ctx context.Context, job *model.QueueJob) {
	// Get bond data for analysis
	bondJSON, err := w.getBondDataJSON(ctx, job)
	if err != nil {
		log.Printf("ERROR: get bond data for %s: %v", job.SECID, err)
		w.queueSvc.MarkError(ctx, job.JobID, "get bond data: "+err.Error())
		return
	}

	// Run AI analysis
	analysis, err := w.analysisSvc.Analyze(ctx, job.SECID, bondJSON)
	if err != nil {
		log.Printf("ERROR: AI analysis for %s: %v", job.SECID, err)
		w.queueSvc.MarkError(ctx, job.JobID, "ai analysis: "+err.Error())
		return
	}

	// Mark done with analysis ID as result
	result := map[string]any{
		"analysis_id": analysis.ID,
		"rating":      analysis.Rating,
	}
	if err := w.queueSvc.MarkDone(ctx, job.JobID, result); err != nil {
		log.Printf("ERROR: mark job done %s: %v", job.JobID, err)
	}

	log.Printf("Job %s done: analysis=%s, rating=%v", job.JobID, analysis.ID, analysis.Rating)
}

// getBondDataJSON fetches bond data from MOEX and serializes it for AI
func (w *Worker) getBondDataJSON(ctx context.Context, job *model.QueueJob) (string, error) {
	// If job.Data already contains json_data, use it
	if job.Data != nil {
		if dataMap, ok := job.Data.(map[string]any); ok {
			if jsonData, ok := dataMap["json_data"]; ok && jsonData != nil {
				b, err := json.Marshal(jsonData)
				if err == nil && len(b) > 2 { // not "{}" or "null"
					return string(b), nil
				}
			}
		}
	}

	// Otherwise fetch from MOEX
	detail, err := w.bondSvc.GetBondDetail(ctx, job.SECID)
	if err != nil {
		return "", err
	}

	coupons, err := w.bondSvc.GetBondCoupons(ctx, job.SECID)
	if err != nil {
		return "", err
	}

	history, err := w.bondSvc.GetBondHistory(ctx, job.SECID)
	if err != nil {
		return "", err
	}

	data := map[string]any{
		"bond":    detail,
		"coupons": coupons,
		"history": history,
	}

	// Fetch dohod.ru data (emitter quality, credit ratings, financials)
	if isin := detail.ISIN; isin != "" {
		dohodData, err := w.detailsSvc.GetDetails(ctx, job.SECID, isin)
		if err != nil {
			log.Printf("WARN: dohod cache lookup for %s: %v", isin, err)
		}
		if dohodData == nil {
			// Not cached — fetch synchronously (with retries)
			log.Printf("Fetching dohod.ru data for %s (%s) before AI analysis", job.SECID, isin)
			dohodData, err = w.detailsSvc.FetchAndSave(ctx, job.SECID, isin)
			if err != nil {
				log.Printf("WARN: dohod fetch for %s failed: %v (AI will proceed without it)", isin, err)
			}
		}
		if dohodData != nil {
			data["dohodDetails"] = dohodData
		}
	}

	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (w *Worker) processDohodParse(ctx context.Context, job *model.QueueJob) {
	// Extract ISIN from job data (handles both map[string]any and primitive.D)
	isin := ""
	if job.Data != nil {
		switch data := job.Data.(type) {
		case map[string]any:
			if v, ok := data["isin"].(string); ok {
				isin = v
			}
		case primitive.D:
			for _, elem := range data {
				if elem.Key == "isin" {
					if v, ok := elem.Value.(string); ok {
						isin = v
					}
				}
			}
		}
	}
	if isin == "" {
		log.Printf("ERROR: missing isin in job data for %s (data type: %T)", job.JobID, job.Data)
		w.queueSvc.MarkError(ctx, job.JobID, "missing isin in job data")
		return
	}

	data, err := w.detailsSvc.FetchAndSave(ctx, job.SECID, isin)
	if err != nil {
		log.Printf("ERROR: dohod parse for %s (%s): %v", job.SECID, isin, err)
		w.queueSvc.MarkError(ctx, job.JobID, "dohod parse: "+err.Error())
		return
	}

	result := map[string]any{
		"isin":          data.ISIN,
		"credit_rating": data.CreditRatingText,
		"quality":       data.Quality,
	}
	if err := w.queueSvc.MarkDone(ctx, job.JobID, result); err != nil {
		log.Printf("ERROR: mark job done %s: %v", job.JobID, err)
	}

	log.Printf("Job %s done: dohod data for %s (%s)", job.JobID, job.SECID, isin)
}
