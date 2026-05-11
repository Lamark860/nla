package model

import "time"

// This file holds the scoring-related model types that have no dependency
// on the scoring engine package itself. BondScore lives next to its
// persistence in repository.ScoringRepo because it composes scoring.
// ScoreResult — pulling the engine into the model package would create
// an import cycle (scoring → model → scoring).

// ScoringProfile is a named weight set used by the scoring engine. Loaded
// from the `scoring_profiles` table. Preset rows (is_preset=true) own no
// user_id and ship with the migration; user-defined profiles will plug into
// this same table later (Фаза 7 plan).
type ScoringProfile struct {
	Code      string             `json:"code"`
	Name      string             `json:"name"`
	IsPreset  bool               `json:"is_preset"`
	UserID    *int64             `json:"user_id,omitempty"`
	Weights   map[string]float64 `json:"weights"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// BondScoreExplanation is a cached LLM-generated text explanation of one
// BondScore. The full text is stored verbatim; regeneration replaces the
// row via FK CASCADE on bond_scores.
type BondScoreExplanation struct {
	ID          int64     `json:"id"`
	BondScoreID int64     `json:"bond_score_id"`
	LLMModel    string    `json:"llm_model"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"created_at"`
}

// ScoreExplainJobData is the payload stored in queue_jobs.data for a
// JobTypeScoreExplain task.
type ScoreExplainJobData struct {
	Profile string `json:"profile"`
	ScoreID int64  `json:"score_id"`
}
