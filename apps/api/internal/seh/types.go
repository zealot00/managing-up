// Package seh provides HTTP handlers for the Skill Evaluation Harness API.
package seh

import "time"

// ----------------------------------------------------------------------------
// Auth DTOs
// ----------------------------------------------------------------------------

// AuthTokenRequest is the request body for POST /auth/token.
type AuthTokenRequest struct {
	APIKey string `json:"api_key"`
}

// AuthTokenResponse is the response for POST /auth/token.
type AuthTokenResponse struct {
	Token    string `json:"token"`
	Role     string `json:"role"`
	IssuedAt string `json:"issued_at"`
}

// AuthTokenRecord represents a stored auth token.
type AuthTokenRecord struct {
	Token      string    `json:"token"`
	Role       string    `json:"role"`
	APIKeyHash string    `json:"api_key_hash"`
	IssuedAt   time.Time `json:"issued_at"`
}

// ----------------------------------------------------------------------------
// Dataset DTOs
// ----------------------------------------------------------------------------

// DatasetSummaryDTO is the summary view of a dataset.
type DatasetSummaryDTO struct {
	DatasetID string `json:"dataset_id"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	Owner     string `json:"owner"`
	CaseCount int    `json:"case_count"`
	CreatedAt string `json:"created_at"`
}

// DatasetDetailDTO is the detailed view of a dataset.
type DatasetDetailDTO struct {
	DatasetID   string          `json:"dataset_id"`
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Owner       string          `json:"owner"`
	Description string          `json:"description"`
	Manifest    DatasetManifest `json:"manifest"`
	CaseCount   int             `json:"case_count"`
	Checksum    string          `json:"checksum"`
	CreatedAt   string          `json:"created_at"`
}

// DatasetManifest contains dataset metadata.
type DatasetManifest struct {
	DatasetName string `json:"dataset_name"`
	Version     string `json:"version"`
	Owner       string `json:"owner"`
	Description string `json:"description"`
}

// DatasetVerifyDTO is the response for dataset verification.
type DatasetVerifyDTO struct {
	Valid      bool   `json:"valid"`
	Checksum   string `json:"checksum"`
	Expected   string `json:"expected,omitempty"`
	Actual     string `json:"actual,omitempty"`
	CaseCount  int    `json:"case_count"`
	VerifiedAt string `json:"verified_at"`
}

// DatasetCasesResponse is the response for GET /datasets/:dataset_id/cases.
type DatasetCasesResponse struct {
	Manifest   DatasetManifest     `json:"manifest"`
	Cases      []EvaluationCaseDTO `json:"cases"`
	Pagination Pagination          `json:"pagination"`
}

// DatasetListResponse is the response for GET /datasets.
type DatasetListResponse struct {
	Datasets   []DatasetSummaryDTO `json:"datasets"`
	Pagination Pagination          `json:"pagination"`
}

// ----------------------------------------------------------------------------
// Case DTOs
// ----------------------------------------------------------------------------

// EvaluationCaseDTO represents an evaluation case.
type EvaluationCaseDTO struct {
	CaseID     string         `json:"case_id"`
	Skill      string         `json:"skill"`
	Source     string         `json:"source"`
	Status     string         `json:"status"`
	Provenance CaseProvenance `json:"provenance"`
	Input      map[string]any `json:"input"`
	Expected   map[string]any `json:"expected"`
	Tags       []string       `json:"tags"`
}

// CaseProvenance contains provenance information for a case.
type CaseProvenance struct {
	ApprovedBy     string `json:"approved_by"`
	ContributorID  string `json:"contributor_id"`
	AttackCategory string `json:"attack_category"`
	GeneratorID    string `json:"generator_id"`
	Method         string `json:"method"`
	Seed           string `json:"seed"`
}

// ----------------------------------------------------------------------------
// Run DTOs
// ----------------------------------------------------------------------------

// CaseRunResultDTO represents the result of a single case run.
type CaseRunResultDTO struct {
	CaseID           string               `json:"case_id"`
	Success          bool                 `json:"success"`
	LatencyMs        int64                `json:"latency_ms"`
	TokenUsage       int64                `json:"token_usage"`
	Output           map[string]any       `json:"output"`
	Error            string               `json:"error"`
	Classification   string               `json:"classification"`
	FailureClusterID string               `json:"failure_cluster_id"`
	Provenance       CaseResultProvenance `json:"provenance"`
	Trajectory       CaseTrajectory       `json:"trajectory"`
}

// CaseResultProvenance contains provenance for a run result.
type CaseResultProvenance struct {
	Source string `json:"source"`
	Status string `json:"status"`
}

// CaseTrajectory contains the execution trajectory of a case.
type CaseTrajectory struct {
	Steps                   []string `json:"steps"`
	ToolCalls               []string `json:"tool_calls"`
	ReasoningTokensEstimate int64    `json:"reasoning_tokens_estimate"`
}

// RunMetricsDTO contains metrics for a run.
type RunMetricsDTO struct {
	SuccessRate          float64 `json:"success_rate"`
	AvgTokens            float64 `json:"avg_tokens"`
	P95Latency           int64   `json:"p95_latency"`
	CostFactor           float64 `json:"cost_factor"`
	ClassificationFactor float64 `json:"classification_factor"`
	CostUSD              float64 `json:"cost_usd"`
	StabilityVariance    float64 `json:"stability_variance"`
	Score                float64 `json:"score"`
}

// RunResultDTO represents a complete run result.
type RunResultDTO struct {
	RunID     string             `json:"run_id"`
	DatasetID string             `json:"dataset_id"`
	Skill     string             `json:"skill"`
	Runtime   string             `json:"runtime"`
	Metrics   RunMetricsDTO      `json:"metrics"`
	Results   []CaseRunResultDTO `json:"results"`
	CreatedAt string             `json:"created_at"`
}

// RunCreateRequest is the request body for POST /runs.
type RunCreateRequest struct {
	DatasetID string             `json:"dataset_id"`
	Skill     string             `json:"skill"`
	Runtime   string             `json:"runtime"`
	Metrics   RunMetricsDTO      `json:"metrics"`
	Results   []CaseRunResultDTO `json:"results"`
}

// RunCreateResponse is the response for POST /runs.
type RunCreateResponse struct {
	RunID       string  `json:"run_id"`
	Score       float64 `json:"score"`
	SuccessRate float64 `json:"success_rate"`
	CreatedAt   string  `json:"created_at"`
}

// RunListResponse is the response for GET /runs.
type RunListResponse struct {
	Runs       []RunResultDTO `json:"runs"`
	Pagination Pagination     `json:"pagination"`
}

// ----------------------------------------------------------------------------
// Gate DTOs
// ----------------------------------------------------------------------------

// GateRequest is the request body for POST /runs/:run_id/gate.
type GateRequest struct {
	PolicyID string `json:"policy_id"`
}

// GateResponse is the response for gate evaluation.
type GateResponse struct {
	Passed      bool         `json:"passed"`
	RunID       string       `json:"run_id"`
	PolicyID    string       `json:"policy_id"`
	Details     []GateDetail `json:"details"`
	EvaluatedAt string       `json:"evaluated_at"`
}

// GateDetail contains details about a single gate rule evaluation.
type GateDetail struct {
	Rule     string  `json:"rule"`
	Passed   bool    `json:"passed"`
	Required float64 `json:"required"`
	Actual   float64 `json:"actual"`
}

// ----------------------------------------------------------------------------
// Policy DTOs
// ----------------------------------------------------------------------------

// GovernancePolicyDTO represents a governance policy.
type GovernancePolicyDTO struct {
	PolicyID                string         `json:"policy_id"`
	Name                    string         `json:"name"`
	RequireProvenance       bool           `json:"require_provenance"`
	RequireApprovedForScore bool           `json:"require_approved_for_score"`
	MinSourceDiversity      int            `json:"min_source_diversity"`
	MinGoldenWeight         float64        `json:"min_golden_weight"`
	SourcePolicies          []SourcePolicy `json:"source_policies"`
	CreatedAt               string         `json:"created_at"`
}

// SourcePolicy represents a policy for a specific source.
type SourcePolicy struct {
	Source               string  `json:"source"`
	Weight               float64 `json:"weight"`
	CountInScore         bool    `json:"count_in_score"`
	MinSuccessRate       float64 `json:"min_success_rate"`
	AdversarialThreshold float64 `json:"adversarial_threshold"`
}

// ----------------------------------------------------------------------------
// Common DTOs
// ----------------------------------------------------------------------------

// ErrorResponse represents an API error.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details.
type ErrorDetail struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

// FieldError represents a field-level error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Pagination contains pagination information.
type Pagination struct {
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	Total   int  `json:"total"`
	HasMore bool `json:"has_more"`
}
