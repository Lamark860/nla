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

// Worker polls MongoDB for pending jobs and processes them
type Worker struct {
	queueSvc     *service.QueueService
	analysisSvc  *service.AnalysisService
	bondSvc      *service.BondService
	detailsSvc   *service.DetailsService
	pollInterval time.Duration
}

func NewWorker(
	queueSvc *service.QueueService,
	analysisSvc *service.AnalysisService,
	bondSvc *service.BondService,
	detailsSvc *service.DetailsService,
) *Worker {
	return &Worker{
		queueSvc:     queueSvc,
		analysisSvc:  analysisSvc,
		bondSvc:      bondSvc,
		detailsSvc:   detailsSvc,
		pollInterval: 2 * time.Second,
	}
}

// Run starts the worker loop. Blocks until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
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
	default:
		log.Printf("WARN: unknown job type %q, marking error", job.Type)
		w.queueSvc.MarkError(ctx, job.JobID, "unknown job type: "+job.Type)
	}
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
