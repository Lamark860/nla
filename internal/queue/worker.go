package queue

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"nla/internal/model"
	"nla/internal/service"
)

// Worker polls MongoDB for pending jobs and processes them
type Worker struct {
	queueSvc    *service.QueueService
	analysisSvc *service.AnalysisService
	bondSvc     *service.BondService
	pollInterval time.Duration
}

func NewWorker(
	queueSvc *service.QueueService,
	analysisSvc *service.AnalysisService,
	bondSvc *service.BondService,
) *Worker {
	return &Worker{
		queueSvc:     queueSvc,
		analysisSvc:  analysisSvc,
		bondSvc:      bondSvc,
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

	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
