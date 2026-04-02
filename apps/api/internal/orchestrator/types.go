// Package orchestrator provides HTTP handlers for sop-to-skill CLI orchestration API.
package orchestrator

import "time"

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
	Time    string `json:"time"`
}

// ErrorResponse represents an API error.
type ErrorResponse struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
	RequestID string         `json:"requestId,omitempty"`
}

// SOPSource represents a source of SOP content (oneOf: inline_text or file_uri).
type SOPSource struct {
	Type    string `json:"type"` // "inline_text" or "file_uri"
	Content string `json:"content,omitempty"`
	URI     string `json:"uri,omitempty"`
}

// ExtractionOptions controls extraction behavior.
type ExtractionOptions struct {
	Language                string  `json:"language,omitempty"`            // auto, zh, en
	ConfidenceThreshold     float64 `json:"confidenceThreshold,omitempty"` // 0-1
	RoleConfigPath          string  `json:"roleConfigPath,omitempty"`
	EnableBoundaryDetection bool    `json:"enableBoundaryDetection,omitempty"` // default true
}

// RunOptions controls orchestration run behavior.
type RunOptions struct {
	Framework   string              `json:"framework,omitempty"` // openclaw, opencode, codex, gpts, mcp, claude, langchain, all
	ConfigRef   string              `json:"configRef,omitempty"`
	Extraction  *ExtractionOptions  `json:"extraction,omitempty"`
	Progressive *ProgressiveOptions `json:"progressive,omitempty"`
}

// ProgressiveOptions controls progressive execution.
type ProgressiveOptions struct {
	Enabled bool `json:"enabled"`
}

// CreateRunRequest is the request body for POST /v1/runs.
type CreateRunRequest struct {
	SkillName string      `json:"skillName"`
	Source    SOPSource   `json:"source"`
	Options   *RunOptions `json:"options,omitempty"`
}

// RunAcceptedResponse is the 202 response for POST /v1/runs.
type RunAcceptedResponse struct {
	RunID     string   `json:"runId"`
	Status    string   `json:"status"` // "queued"
	CreatedAt string   `json:"createdAt"`
	Links     RunLinks `json:"links"`
}

// RunLinks contains hypermedia links for a run.
type RunLinks struct {
	Self string `json:"self"`
}

// ArtifactRef references an artifact by kind and URI.
type ArtifactRef struct {
	Kind string `json:"kind"` // skill_md, full_skill_md, schema_json, manifest_yaml, constraints_dir, framework_bundle
	URI  string `json:"uri"`
}

// RunResult contains the result of a completed run.
type RunResult struct {
	SkillID   string        `json:"skillId,omitempty"`
	Version   string        `json:"version,omitempty"`
	Artifacts []ArtifactRef `json:"artifacts,omitempty"`
}

// RunDetail is the detailed view of a run.
type RunDetail struct {
	RunID     string          `json:"runId"`
	Status    string          `json:"status"` // queued, running, succeeded, failed, canceled
	Stage     string          `json:"stage"`  // extraction, generation, validation, versioning, testing, completed
	SkillName string          `json:"skillName"`
	CreatedAt string          `json:"createdAt"`
	UpdatedAt string          `json:"updatedAt"`
	Result    *RunResult      `json:"result,omitempty"`
	Errors    []ErrorResponse `json:"errors,omitempty"`
}

// ArtifactListResponse is the response for GET /v1/runs/{runId}/artifacts.
type ArtifactListResponse struct {
	RunID     string        `json:"runId"`
	Artifacts []ArtifactRef `json:"artifacts"`
}

// Constraint represents an extracted constraint from SOP.
type Constraint struct {
	ID          string   `json:"id"`
	Level       string   `json:"level"` // MUST, SHOULD, MAY
	Description string   `json:"description"`
	Condition   string   `json:"condition,omitempty"`
	Action      string   `json:"action,omitempty"`
	Roles       []string `json:"roles"`
	Confidence  float64  `json:"confidence"` // 0-1
}

// DecisionRule represents a rule within a decision.
type DecisionRule struct {
	Condition string         `json:"condition"`
	Output    map[string]any `json:"output"`
	Priority  int            `json:"priority,omitempty"`
}

// Decision represents an extracted decision from SOP.
type Decision struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	InputVars  []string       `json:"inputVars"`
	OutputVars []string       `json:"outputVars"`
	Rules      []DecisionRule `json:"rules"`
}

// RoleStat represents statistics about a role mentioned in SOP.
type RoleStat struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Mentions    int    `json:"mentions"`
	Source      string `json:"source,omitempty"`
}

// BoundaryParameter represents a boundary parameter from SOP.
type BoundaryParameter struct {
	Name         string  `json:"name"`
	MinValue     float64 `json:"minValue,omitempty"`
	MaxValue     float64 `json:"maxValue,omitempty"`
	DefaultValue float64 `json:"defaultValue,omitempty"`
	Unit         string  `json:"unit,omitempty"`
	Confidence   float64 `json:"confidence"` // 0-1
}

// ModelInfo contains information about the model used for extraction.
type ModelInfo struct {
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	LatencyMs int    `json:"latencyMs"`
}

// EnhanceExtractionRequest is the request body for POST /v1/extraction/enhance.
type EnhanceExtractionRequest struct {
	Source  SOPSource          `json:"source"`
	Options *ExtractionOptions `json:"options,omitempty"`
}

// EnhancedExtractionResponse is the response for enhanced extraction.
type EnhancedExtractionResponse struct {
	Constraints []Constraint        `json:"constraints"`
	Decisions   []Decision          `json:"decisions"`
	Roles       []RoleStat          `json:"roles"`
	Boundaries  []BoundaryParameter `json:"boundaries"`
	ModelInfo   ModelInfo           `json:"modelInfo"`
}

// CompareExtractionRequest is the request body for POST /v1/extraction/compare.
type CompareExtractionRequest struct {
	Local  EnhancedExtractionResponse `json:"local"`
	Remote EnhancedExtractionResponse `json:"remote"`
}

// ExtractionDiff represents a single diff between extractions.
type ExtractionDiff struct {
	Type   string `json:"type"` // constraint, decision, role, boundary
	Detail string `json:"detail"`
}

// ExtractionComparisonSummary contains delta summaries.
type ExtractionComparisonSummary struct {
	ConstraintDelta int `json:"constraintDelta"`
	DecisionDelta   int `json:"decisionDelta"`
	RoleDelta       int `json:"roleDelta"`
}

// ExtractionComparisonResponse is the response for extraction comparison.
type ExtractionComparisonResponse struct {
	Summary ExtractionComparisonSummary `json:"summary"`
	Diffs   []ExtractionDiff            `json:"diffs"`
}

// CreateSkillRequest is the request body for POST /v1/skills.
type CreateSkillRequest struct {
	SkillID string   `json:"skillId"`
	Name    string   `json:"name"`
	Owner   string   `json:"owner,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

// Skill represents a skill in the registry.
type Skill struct {
	SkillID   string    `json:"skillId"`
	Name      string    `json:"name"`
	Owner     string    `json:"owner,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// CreateSkillVersionRequest is the request body for POST /v1/skills/{skillId}/versions.
type CreateSkillVersionRequest struct {
	Version        string         `json:"version"`
	SourceHash     string         `json:"sourceHash"`
	SchemaHash     string         `json:"schemaHash"`
	ConfigSnapshot map[string]any `json:"configSnapshot,omitempty"`
	Artifacts      []ArtifactRef  `json:"artifacts"`
	RunID          string         `json:"runId,omitempty"`
}

// SkillVersion represents a specific version of a skill.
type SkillVersion struct {
	SkillID        string         `json:"skillId"`
	Version        string         `json:"version"`
	SourceHash     string         `json:"sourceHash"`
	SchemaHash     string         `json:"schemaHash"`
	ConfigSnapshot map[string]any `json:"configSnapshot,omitempty"`
	Artifacts      []ArtifactRef  `json:"artifacts"`
	RunID          string         `json:"runId,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	Promoted       bool           `json:"promoted,omitempty"`
}

// SkillVersionList is the response for GET /v1/skills/{skillId}/versions.
type SkillVersionList struct {
	SkillID  string         `json:"skillId"`
	Versions []SkillVersion `json:"versions"`
}

// VersionDiffDetail represents a single change in a version diff.
type VersionDiffDetail struct {
	Path   string `json:"path,omitempty"`
	Before any    `json:"before,omitempty"`
	After  any    `json:"after,omitempty"`
}

// VersionDiffSummary contains summary statistics for a diff.
type VersionDiffSummary struct {
	ConstraintsChanged int `json:"constraintsChanged"`
	DecisionsChanged   int `json:"decisionsChanged"`
	StepsChanged       int `json:"stepsChanged"`
}

// VersionDiffResponse is the response for GET /v1/skills/{skillId}/diff.
type VersionDiffResponse struct {
	SkillID string              `json:"skillId"`
	From    string              `json:"from"`
	To      string              `json:"to"`
	Summary VersionDiffSummary  `json:"summary"`
	Details []VersionDiffDetail `json:"details,omitempty"`
}

// RollbackRequest is the request body for POST /v1/skills/{skillId}/rollback.
type RollbackRequest struct {
	TargetVersion string `json:"targetVersion"`
	Reason        string `json:"reason"`
}

// PromoteRequest is the request body for POST /v1/skills/{skillId}/promote.
type PromoteRequest struct {
	Version string `json:"version"`
	Channel string `json:"channel"` // dev, staging, prod
}

// ActionAccepted is the 202 response for rollback/promote actions.
type ActionAccepted struct {
	ActionID   string    `json:"actionId"`
	Status     string    `json:"status"` // "accepted"
	AcceptedAt time.Time `json:"acceptedAt"`
}

// TestRunner represents a test runner configuration.
type TestRunner struct {
	Type    string   `json:"type"` // "cli"
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// CreateTestRunRequest is the request body for POST /v1/tests/runs.
type CreateTestRunRequest struct {
	SkillID    string     `json:"skillId"`
	Version    string     `json:"version"`
	Runner     TestRunner `json:"runner"`
	DatasetRef string     `json:"datasetRef,omitempty"`
}

// TestRunAccepted is the 202 response for POST /v1/tests/runs.
type TestRunAccepted struct {
	TestRunID string `json:"testRunId"`
	Status    string `json:"status"` // "queued"
}

// TestRun represents a test run.
type TestRun struct {
	TestRunID  string     `json:"testRunId"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt,omitempty"`
	ExitCode   int        `json:"exitCode,omitempty"`
	SkillID    string     `json:"-"`
	Version    string     `json:"-"`
	Runner     TestRunner `json:"-"`
	DatasetRef string     `json:"-"`
}

// TestFailure represents a single test failure.
type TestFailure struct {
	CaseID string `json:"caseId"`
	Reason string `json:"reason"`
}

// TestMetrics contains test run metrics.
type TestMetrics struct {
	TotalCases  int     `json:"totalCases"`
	PassRate    float64 `json:"passRate"` // 0-1
	Regressions int     `json:"regressions,omitempty"`
}

// TestReport is the response for GET /v1/tests/runs/{testRunId}/report.
type TestReport struct {
	TestRunID string        `json:"testRunId"`
	Passed    bool          `json:"passed"`
	Metrics   TestMetrics   `json:"metrics"`
	Failures  []TestFailure `json:"failures,omitempty"`
}

// GateEvaluateRequest is the request body for POST /v1/gates/evaluate.
type GateEvaluateRequest struct {
	SkillID      string         `json:"skillId"`
	Version      string         `json:"version"`
	PolicyID     string         `json:"policyId"`
	TestRunID    string         `json:"testRunId,omitempty"`
	ExtraSignals map[string]any `json:"extraSignals,omitempty"`
}

// GateEvaluateResponse is the response for gate evaluation.
type GateEvaluateResponse struct {
	Passed     bool      `json:"passed"`
	PolicyID   string    `json:"policyId"`
	Reasons    []string  `json:"reasons"`
	DecisionAt time.Time `json:"decisionAt"`
}

// PolicyRule represents a single rule in a policy.
type PolicyRule struct {
	Metric string `json:"metric"`
	Op     string `json:"op"` // gte, lte, eq, neq
	Value  any    `json:"value"`
}

// Policy represents a gate policy.
type Policy struct {
	PolicyID string       `json:"policyId"`
	Name     string       `json:"name"`
	Rules    []PolicyRule `json:"rules"`
}
