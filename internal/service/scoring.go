package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"nla/internal/client/openai"
	"nla/internal/model"
	"nla/internal/repository"
	"nla/internal/scoring"
)

// ScoreCacheTTL controls how long a cached bond_scores row counts as fresh
// before the next API hit recomputes from live MOEX data. 24h matches the
// roadmap: scoring is deterministic over the day's snapshot, calibration
// happens off the cached values.
const ScoreCacheTTL = 24 * time.Hour

// ScoreResponse is the API-shaped response the handler returns. It
// combines the cached row's metadata (id, computed_at) with the decoded
// ScoreResult so the frontend gets one flat JSON object instead of
// {score, breakdown:{score, breakdown}}.
type ScoreResponse struct {
	ScoreID        int64                   `json:"score_id"`
	SECID          string                  `json:"secid"`
	ProfileCode    string                  `json:"profile_code"`
	ProfileName    string                  `json:"profile_name,omitempty"`
	Score          float64                 `json:"score"`
	Breakdown      []scoring.BreakdownItem `json:"breakdown"`
	MissingFactors []string                `json:"missing_factors,omitempty"`
	ComputedAt     time.Time               `json:"computed_at"`
	Explanation    *model.BondScoreExplanation `json:"explanation,omitempty"`
}

// ErrUnknownProfile is returned when a code can't be resolved against the
// in-memory presets or the DB. Handler maps it to HTTP 404.
var ErrUnknownProfile = errors.New("unknown scoring profile")

// ScoringService stitches the pure scoring engine (internal/scoring)
// together with the data sources (bonds / ratings / dohod) and the cache.
//
// The OpenAI client and prompt path are optional — if both are zero,
// Explain() returns a sentinel error rather than panicking. Useful for
// tests and for early dev environments without an OPENAI_API_KEY.
type ScoringService struct {
	repo       *repository.ScoringRepo
	bondSvc    *BondService
	ratingSvc  *RatingService
	detailsSvc *DetailsService

	openai     *openai.Client
	promptPath string
}

func NewScoringService(
	repo *repository.ScoringRepo,
	bondSvc *BondService,
	ratingSvc *RatingService,
	detailsSvc *DetailsService,
	openaiClient *openai.Client,
	promptPath string,
) *ScoringService {
	return &ScoringService{
		repo:       repo,
		bondSvc:    bondSvc,
		ratingSvc:  ratingSvc,
		detailsSvc: detailsSvc,
		openai:     openaiClient,
		promptPath: promptPath,
	}
}

// ListProfiles returns the available scoring profiles. For now only presets
// (low/mid/high) — user profiles will land with Фаза 7 / billing.
func (s *ScoringService) ListProfiles(ctx context.Context) ([]model.ScoringProfile, error) {
	return s.repo.ListProfiles(ctx, nil)
}

// ComputeAll runs every preset profile for one bond. Cheap because each
// individual ComputeOne hits the cache when fresh; only the first call of
// the day on a given bond pays the full cost.
func (s *ScoringService) ComputeAll(ctx context.Context, secid string) ([]ScoreResponse, error) {
	codes := []string{scoring.ProfileLow, scoring.ProfileMid, scoring.ProfileHigh}
	out := make([]ScoreResponse, 0, len(codes))
	for _, code := range codes {
		res, err := s.ComputeOne(ctx, secid, code)
		if err != nil {
			return nil, fmt.Errorf("compute %s: %w", code, err)
		}
		out = append(out, *res)
	}
	return out, nil
}

// ComputeOne returns the score for one profile. Cache hit when the latest
// row is younger than ScoreCacheTTL; otherwise pulls live data, runs the
// engine, persists.
func (s *ScoringService) ComputeOne(ctx context.Context, secid, profileCode string) (*ScoreResponse, error) {
	profile, err := s.resolveProfile(ctx, profileCode)
	if err != nil {
		return nil, err
	}

	// Cache lookup. A miss is not an error.
	cached, err := s.repo.GetLatestScore(ctx, secid, profileCode)
	if err != nil {
		return nil, fmt.Errorf("cache lookup: %w", err)
	}
	if cached != nil && time.Since(cached.ComputedAt) < ScoreCacheTTL {
		return s.buildResponse(ctx, cached, profile.Name), nil
	}

	// Compute fresh.
	input, err := s.loadInput(ctx, secid)
	if err != nil {
		return nil, fmt.Errorf("load input for %s: %w", secid, err)
	}

	engineProfile := scoring.Profile{
		Code:    profile.Code,
		Name:    profile.Name,
		Weights: profile.Weights,
	}
	result := scoring.Compute(input, engineProfile)

	stored := &repository.BondScore{
		SECID:       secid,
		ProfileCode: profile.Code,
		Result:      result,
	}
	if err := s.repo.InsertScore(ctx, stored); err != nil {
		return nil, fmt.Errorf("persist score: %w", err)
	}

	return s.buildResponse(ctx, stored, profile.Name), nil
}

// GetExplanation returns the latest cached explanation for one scored row.
// Returns (nil, nil) when nothing has been generated yet — the handler maps
// that to "no explanation, click to generate".
func (s *ScoringService) GetExplanation(ctx context.Context, scoreID int64) (*model.BondScoreExplanation, error) {
	return s.repo.GetExplanationByScoreID(ctx, scoreID)
}

// Explain is called by the worker for a score_explain job. It fetches the
// stored BondScore, asks the LLM to narrate its breakdown, persists the
// text, and returns the new explanation row.
func (s *ScoringService) Explain(ctx context.Context, scoreID int64) (*model.BondScoreExplanation, error) {
	if s.openai == nil {
		return nil, errors.New("openai client not configured")
	}

	scored, err := s.repo.GetScoreByID(ctx, scoreID)
	if err != nil {
		return nil, fmt.Errorf("get score: %w", err)
	}
	if scored == nil {
		return nil, fmt.Errorf("score %d not found", scoreID)
	}

	prompt, err := s.loadPrompt()
	if err != nil {
		return nil, fmt.Errorf("load prompt: %w", err)
	}

	userMsg, err := s.buildExplainMessage(scored)
	if err != nil {
		return nil, fmt.Errorf("build explain message: %w", err)
	}

	text, err := s.openai.ChatCompletion(ctx, []openai.ChatMessage{
		{Role: "system", Content: prompt},
		{Role: "user", Content: userMsg},
	})
	if err != nil {
		return nil, fmt.Errorf("openai completion: %w", err)
	}

	exp := &model.BondScoreExplanation{
		BondScoreID: scoreID,
		LLMModel:    s.openai.Model(),
		Text:        text,
	}
	if err := s.repo.InsertExplanation(ctx, exp); err != nil {
		return nil, fmt.Errorf("save explanation: %w", err)
	}
	return exp, nil
}

// ---------- helpers ----------

// resolveProfile checks the in-process preset map first (covers the 99%
// case and survives transient DB hiccups), then falls back to the repo for
// custom profiles. The DB row is preferred over the preset constant when
// both exist — operators can adjust seed weights live by UPDATE-ing the
// row, no redeploy.
func (s *ScoringService) resolveProfile(ctx context.Context, code string) (*model.ScoringProfile, error) {
	row, err := s.repo.GetProfile(ctx, code)
	if err == nil {
		return row, nil
	}
	// DB miss or noRows — fall through to preset.
	if preset, ok := scoring.Presets[code]; ok {
		return &model.ScoringProfile{
			Code:     preset.Code,
			Name:     preset.Name,
			IsPreset: true,
			Weights:  preset.Weights,
		}, nil
	}
	return nil, ErrUnknownProfile
}

// loadInput gathers everything the engine needs for one bond. Missing
// dohod / benchmark data is treated as "factor unavailable" downstream,
// not as an error — partial scoring is better than no scoring.
func (s *ScoringService) loadInput(ctx context.Context, secid string) (scoring.Input, error) {
	bond, err := s.bondSvc.GetBondDetail(ctx, secid)
	if err != nil {
		return scoring.Input{}, fmt.Errorf("get bond %s: %w", secid, err)
	}
	if bond == nil {
		return scoring.Input{}, fmt.Errorf("bond %s not found", secid)
	}

	in := scoring.Input{Bond: *bond}

	// Ratings: keyed by emitter_id, so we need a resolved issuer first.
	if bond.EmitterID != nil && *bond.EmitterID > 0 {
		rr, err := s.ratingSvc.GetByEmitterID(ctx, *bond.EmitterID)
		if err == nil && rr != nil {
			in.Ratings = rr.Ratings
		}
	}

	// Dohod is best-effort: cached row if present, no fetch.
	if d, err := s.detailsSvc.GetDetails(ctx, secid, bond.ISIN); err == nil && d != nil {
		in.Dohod = d
	}

	// ОФЗ benchmark is Phase 2 follow-up — leaving Bench unset means
	// factor #3 (ytm_premium) ends up in MissingFactors. Acceptable
	// during initial calibration on the other 11 factors.
	in.BenchmarkYieldPct = nil

	return in, nil
}

func (s *ScoringService) buildResponse(ctx context.Context, stored *repository.BondScore, profileName string) *ScoreResponse {
	resp := &ScoreResponse{
		ScoreID:        stored.ID,
		SECID:          stored.SECID,
		ProfileCode:    stored.ProfileCode,
		ProfileName:    profileName,
		Score:          stored.Result.Score,
		Breakdown:      stored.Result.Breakdown,
		MissingFactors: stored.Result.MissingFactors,
		ComputedAt:     stored.ComputedAt,
	}
	// Attach the latest cached explanation when one exists. Cheap query
	// hit, but skip if the score was never persisted (id=0 for an
	// unsaved-in-tests path).
	if stored.ID > 0 {
		if exp, err := s.repo.GetExplanationByScoreID(ctx, stored.ID); err == nil && exp != nil {
			resp.Explanation = exp
		} else if err != nil {
			log.Printf("WARN: explanation lookup for score %d: %v", stored.ID, err)
		}
	}
	return resp
}

// loadPrompt reads the LLM system prompt from disk. Tiny file (a few hundred
// bytes) — read on each call to pick up edits in dev without a restart.
func (s *ScoringService) loadPrompt() (string, error) {
	if s.promptPath == "" {
		return "", errors.New("empty prompt path")
	}
	b, err := os.ReadFile(s.promptPath)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", s.promptPath, err)
	}
	return string(b), nil
}

// buildExplainMessage turns the breakdown into a compact user message for
// the LLM. We deliberately do NOT send the full bond payload — the engine
// already distilled what matters into numbered factors, and feeding
// 50KB of MOEX JSON costs tokens without improving the answer.
func (s *ScoringService) buildExplainMessage(stored *repository.BondScore) (string, error) {
	type factorRow struct {
		Factor     string  `json:"factor"`
		Name       string  `json:"name"`
		Normalized float64 `json:"normalized"`
		Weight     float64 `json:"weight"`
		Contribution float64 `json:"contribution"`
		HasData    bool    `json:"has_data"`
	}
	rows := make([]factorRow, 0, len(stored.Result.Breakdown))
	for _, b := range stored.Result.Breakdown {
		rows = append(rows, factorRow{
			Factor: b.Factor, Name: b.Name,
			Normalized: b.Normalized, Weight: b.Weight,
			Contribution: b.Contribution, HasData: b.HasData,
		})
	}
	payload := map[string]any{
		"secid":           stored.SECID,
		"profile":         stored.ProfileCode,
		"score":           stored.Result.Score,
		"breakdown":       rows,
		"missing_factors": stored.Result.MissingFactors,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Объясни на простом языке (3-4 абзаца), почему получился такой балл по профилю «%s» для облигации %s.\n\nРазбор по факторам:\n%s",
		stored.ProfileCode, stored.SECID, string(b)), nil
}
