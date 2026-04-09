package model

import "time"

// QueueJob represents a background task tracked in MongoDB
type QueueJob struct {
	JobID       string     `json:"job_id" bson:"_id,omitempty"`
	Type        string     `json:"type" bson:"type"` // ai_analysis, parse_bond, parse_emitter, sync_issuer
	SECID       string     `json:"secid,omitempty" bson:"secid,omitempty"`
	ReferenceID string     `json:"reference_id,omitempty" bson:"reference_id,omitempty"`
	Status      string     `json:"status" bson:"status"` // pending, running, done, error
	Data        any        `json:"data,omitempty" bson:"data,omitempty"`
	Result      any        `json:"result,omitempty" bson:"result,omitempty"`
	Error       string     `json:"error,omitempty" bson:"error,omitempty"`
	Attempts    int        `json:"attempts" bson:"attempts"`
	MaxAttempts int        `json:"max_attempts" bson:"max_attempts"`
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" bson:"updated_at"`
	StartedAt   *time.Time `json:"started_at,omitempty" bson:"started_at,omitempty"`
	FinishedAt  *time.Time `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}

const (
	JobStatusPending = "pending"
	JobStatusRunning = "running"
	JobStatusDone    = "done"
	JobStatusError   = "error"
)

const (
	JobTypeAIAnalysis   = "ai_analysis"
	JobTypeParseBond    = "parse_bond"
	JobTypeParseEmitter = "parse_emitter"
	JobTypeSyncIssuer   = "sync_issuer"
)

// CreateJobRequest for POST /bonds/:secid/analyze
type AnalyzeRequest struct {
	JSONData   any `json:"json_data,omitempty"`
	CustomJSON any `json:"custom_json,omitempty"`
}

// JobStatusResponse for GET /jobs/:id
type JobStatusResponse struct {
	JobID      string     `json:"job_id"`
	Type       string     `json:"type"`
	Status     string     `json:"status"`
	Result     any        `json:"result,omitempty"`
	Error      string     `json:"error,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}
