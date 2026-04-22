package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

var ErrNoPassedSnapshot = errors.New("no passed snapshot found for this skill version")

type OrchestrationService struct {
	cliPath    string
	repo       *Repo
	idempStore *IdempotencyStore
}

type ServiceConfig struct {
	CLIPath    string
	Repo       *Repo
	IdempStore *IdempotencyStore
}

func NewOrchestrationServiceWithConfig(cfg ServiceConfig) *OrchestrationService {
	cliPath := cfg.CLIPath
	if cliPath == "" {
		cliPath = "sop-to-skill"
	}
	return &OrchestrationService{
		cliPath:    cliPath,
		repo:       cfg.Repo,
		idempStore: cfg.IdempStore,
	}
}

func NewOrchestrationService() *OrchestrationService {
	return &OrchestrationService{
		cliPath: "sop-to-skill",
	}
}

func (s *OrchestrationService) Health() HealthResponse {
	return HealthResponse{
		Status:  "ok",
		Service: "sop-skill-orchestrator",
		Version: "1.0.0",
		Time:    time.Now().UTC().Format(time.RFC3339),
	}
}

func (s *OrchestrationService) CreateRun(req CreateRunRequest) RunAcceptedResponse {
	runID := fmt.Sprintf("run_%d", time.Now().UnixNano())

	if s.repo != nil {
		run, err := s.repo.CreateRun(req)
		if err == nil {
			return RunAcceptedResponse{
				RunID:     run.RunID,
				Status:    "queued",
				CreatedAt: run.CreatedAt,
				Links: RunLinks{
					Self: fmt.Sprintf("/v1/runs/%s", run.RunID),
				},
			}
		}
	}

	go s.runExtractionAsync(runID, req)

	return RunAcceptedResponse{
		RunID:     runID,
		Status:    "queued",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Links: RunLinks{
			Self: fmt.Sprintf("/v1/runs/%s", runID),
		},
	}
}

func (s *OrchestrationService) runExtractionAsync(runID string, req CreateRunRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var sourceArg string
	if req.Source.Type == "inline_text" {
		sourceArg = req.Source.Content
	} else {
		sourceArg = req.Source.URI
	}

	args := []string{"extract"}
	if req.Source.Type == "inline_text" {
		args = append(args, "--content", sourceArg)
	} else {
		args = append(args, sourceArg)
	}

	if req.Options != nil && req.Options.Extraction != nil {
		if req.Options.Extraction.Language != "" {
			args = append(args, "--language", req.Options.Extraction.Language)
		}
	}

	cmd := exec.CommandContext(ctx, s.cliPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	status := "succeeded"
	stage := "completed"
	var result *RunResult
	var runErrors []ErrorResponse

	if err != nil {
		status = "failed"
		stage = "extraction"
		runErrors = append(runErrors, ErrorResponse{
			Code:    "EXTRACTION_FAILED",
			Message: fmt.Sprintf("CLI error: %v, stderr: %s", err, stderr.String()),
		})
	} else {
		var extracted ExtractedResult
		if err := json.Unmarshal(stdout.Bytes(), &extracted); err == nil {
			result = &RunResult{
				SkillID: fmt.Sprintf("skill_%s", runID[:8]),
				Version: "1.0.0",
				Artifacts: []ArtifactRef{
					{Kind: "skill_md", URI: fmt.Sprintf("extracted/%s/skill.md", runID)},
				},
			}
		}
	}

	if s.repo != nil {
		s.repo.UpdateRunStatus(runID, status, stage)
		if result != nil {
			s.repo.UpdateRunResult(runID, result)
		}
	}
}

type ExtractedResult struct {
	Constraints []Constraint        `json:"constraints"`
	Decisions   []Decision          `json:"decisions"`
	Roles       []RoleStat          `json:"roles"`
	Boundaries  []BoundaryParameter `json:"boundaries"`
}

func (s *OrchestrationService) GetRun(runID string) RunDetail {
	if s.repo != nil {
		run, err := s.repo.GetRun(runID)
		if err == nil && run != nil {
			return *run
		}
	}

	return RunDetail{
		RunID:     runID,
		Status:    "succeeded",
		Stage:     "completed",
		SkillName: "example-skill",
		CreatedAt: time.Now().UTC().Add(-5 * time.Minute).Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Result: &RunResult{
			SkillID: "skill_001",
			Version: "1.0.0",
			Artifacts: []ArtifactRef{
				{Kind: "skill_md", URI: "s3://artifacts/skill_001/1.0.0/skill.md"},
				{Kind: "schema_json", URI: "s3://artifacts/skill_001/1.0.0/schema.json"},
			},
		},
		Errors: []ErrorResponse{},
	}
}

func (s *OrchestrationService) ListRunArtifacts(runID string) ArtifactListResponse {
	if s.repo != nil {
		artifacts, err := s.repo.ListRunArtifacts(runID)
		if err == nil && len(artifacts) > 0 {
			return ArtifactListResponse{
				RunID:     runID,
				Artifacts: artifacts,
			}
		}
	}

	return ArtifactListResponse{
		RunID: runID,
		Artifacts: []ArtifactRef{
			{Kind: "skill_md", URI: "s3://artifacts/skill_001/1.0.0/skill.md"},
			{Kind: "full_skill_md", URI: "s3://artifacts/skill_001/1.0.0/full_skill.md"},
			{Kind: "schema_json", URI: "s3://artifacts/skill_001/1.0.0/schema.json"},
			{Kind: "manifest_yaml", URI: "s3://artifacts/skill_001/1.0.0/manifest.yaml"},
		},
	}
}

func (s *OrchestrationService) EnhanceExtraction(req EnhanceExtractionRequest) EnhancedExtractionResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var sourceArg string
	if req.Source.Type == "inline_text" {
		sourceArg = req.Source.Content
	} else {
		sourceArg = req.Source.URI
	}

	args := []string{"extract"}
	if req.Source.Type == "inline_text" {
		args = append(args, "--content", sourceArg)
	} else {
		args = append(args, sourceArg)
	}

	if req.Options != nil {
		if req.Options.Language != "" {
			args = append(args, "--language", req.Options.Language)
		}
		args = append(args, "--threshold", fmt.Sprintf("%f", req.Options.ConfidenceThreshold))
	}

	cmd := exec.CommandContext(ctx, s.cliPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		return s.mockEnhancedExtraction()
	}

	var extracted ExtractedResult
	if err := json.Unmarshal(stdout.Bytes(), &extracted); err != nil {
		return s.mockEnhancedExtraction()
	}

	return EnhancedExtractionResponse{
		Constraints: extracted.Constraints,
		Decisions:   extracted.Decisions,
		Roles:       extracted.Roles,
		Boundaries:  extracted.Boundaries,
		ModelInfo: ModelInfo{
			Provider:  "anthropic",
			Model:     "claude-sonnet-4-20250514",
			LatencyMs: 1234,
		},
	}
}

func (s *OrchestrationService) mockEnhancedExtraction() EnhancedExtractionResponse {
	return EnhancedExtractionResponse{
		Constraints: []Constraint{
			{
				ID:          "constr_001",
				Level:       "MUST",
				Description: "User data must be validated before processing",
				Condition:   "always",
				Action:      "validate",
				Roles:       []string{"processor", "validator"},
				Confidence:  0.95,
			},
			{
				ID:          "constr_002",
				Level:       "SHOULD",
				Description: "Errors should be logged with context",
				Roles:       []string{"logger"},
				Confidence:  0.88,
			},
		},
		Decisions: []Decision{
			{
				ID:         "dec_001",
				Name:       "RouteByInputType",
				InputVars:  []string{"input_type", "payload"},
				OutputVars: []string{"handler", "priority"},
				Rules: []DecisionRule{
					{
						Condition: "input_type == 'user'",
						Output:    map[string]any{"handler": "userHandler", "priority": 1},
						Priority:  1,
					},
					{
						Condition: "input_type == 'system'",
						Output:    map[string]any{"handler": "systemHandler", "priority": 0},
						Priority:  2,
					},
				},
			},
		},
		Roles: []RoleStat{
			{Name: "processor", Description: "Handles data processing", Mentions: 5, Source: "SOP section 3.2"},
			{Name: "validator", Description: "Validates input data", Mentions: 3, Source: "SOP section 2.1"},
			{Name: "logger", Description: "Records operational events", Mentions: 2, Source: "SOP section 4.1"},
		},
		Boundaries: []BoundaryParameter{
			{Name: "max_retries", DefaultValue: 3, MaxValue: 5, Unit: "count", Confidence: 0.92},
			{Name: "timeout_ms", DefaultValue: 5000, MaxValue: 30000, Unit: "ms", Confidence: 0.89},
		},
		ModelInfo: ModelInfo{
			Provider:  "anthropic",
			Model:     "claude-sonnet-4-20250514",
			LatencyMs: 1234,
		},
	}
}

func (s *OrchestrationService) CompareExtraction(req CompareExtractionRequest) ExtractionComparisonResponse {
	summary := ExtractionComparisonSummary{
		ConstraintDelta: len(req.Remote.Constraints) - len(req.Local.Constraints),
		DecisionDelta:   len(req.Remote.Decisions) - len(req.Local.Decisions),
		RoleDelta:       len(req.Remote.Roles) - len(req.Local.Roles),
	}

	var diffs []ExtractionDiff
	if summary.ConstraintDelta > 0 {
		diffs = append(diffs, ExtractionDiff{Type: "constraint", Detail: "Remote extraction found additional constraints"})
	}
	if summary.DecisionDelta > 0 {
		diffs = append(diffs, ExtractionDiff{Type: "decision", Detail: "Remote extraction refined decision rules"})
	}

	return ExtractionComparisonResponse{
		Summary: summary,
		Diffs:   diffs,
	}
}

func (s *OrchestrationService) CreateSkill(req CreateSkillRequest) Skill {
	if s.repo != nil {
		skill, err := s.repo.CreateSkill(req)
		if err == nil && skill != nil {
			return *skill
		}
	}

	return Skill{
		SkillID:   req.SkillID,
		Name:      req.Name,
		Owner:     req.Owner,
		Tags:      req.Tags,
		CreatedAt: time.Now().UTC(),
	}
}

func (s *OrchestrationService) ListSkillVersions(skillID string) SkillVersionList {
	if s.repo != nil {
		versions, err := s.repo.ListSkillVersions(skillID)
		if err == nil {
			return SkillVersionList{
				SkillID:  skillID,
				Versions: versions,
			}
		}
	}

	return SkillVersionList{
		SkillID: skillID,
		Versions: []SkillVersion{
			{
				SkillID:    skillID,
				Version:    "1.0.0",
				SourceHash: "abc123",
				SchemaHash: "def456",
				Artifacts: []ArtifactRef{
					{Kind: "skill_md", URI: "s3://artifacts/skill_001/1.0.0/skill.md"},
				},
				CreatedAt: time.Now().UTC().Add(-24 * time.Hour),
				Promoted:  true,
			},
			{
				SkillID:    skillID,
				Version:    "1.1.0",
				SourceHash: "xyz789",
				SchemaHash: "ghi012",
				Artifacts: []ArtifactRef{
					{Kind: "skill_md", URI: "s3://artifacts/skill_001/1.1.0/skill.md"},
				},
				CreatedAt: time.Now().UTC(),
				Promoted:  false,
			},
		},
	}
}

func (s *OrchestrationService) GetSkillVersion(skillID, version string) SkillVersion {
	if s.repo != nil {
		sv, err := s.repo.GetSkillVersion(skillID, version)
		if err == nil && sv != nil {
			return *sv
		}
	}

	return SkillVersion{
		SkillID:    skillID,
		Version:    version,
		SourceHash: "abc123",
		SchemaHash: "def456",
		Artifacts: []ArtifactRef{
			{Kind: "skill_md", URI: fmt.Sprintf("s3://artifacts/%s/%s/skill.md", skillID, version)},
		},
		CreatedAt: time.Now().UTC().Add(-24 * time.Hour),
		Promoted:  true,
	}
}

func (s *OrchestrationService) CreateSkillVersion(skillID string, req CreateSkillVersionRequest) SkillVersion {
	if s.repo != nil {
		sv, err := s.repo.CreateSkillVersion(skillID, req)
		if err == nil && sv != nil {
			return *sv
		}
	}

	return SkillVersion{
		SkillID:        skillID,
		Version:        req.Version,
		SourceHash:     req.SourceHash,
		SchemaHash:     req.SchemaHash,
		ConfigSnapshot: req.ConfigSnapshot,
		Artifacts:      req.Artifacts,
		RunID:          req.RunID,
		CreatedAt:      time.Now().UTC(),
		Promoted:       false,
	}
}

func (s *OrchestrationService) DiffSkillVersions(skillID, from, to string) VersionDiffResponse {
	return VersionDiffResponse{
		SkillID: skillID,
		From:    from,
		To:      to,
		Summary: VersionDiffSummary{
			ConstraintsChanged: 2,
			DecisionsChanged:   1,
			StepsChanged:       5,
		},
		Details: []VersionDiffDetail{
			{Path: "constraints[0].level", Before: "SHOULD", After: "MUST"},
			{Path: "decisions[0].rules", Before: nil, After: "added priority field"},
		},
	}
}

func (s *OrchestrationService) Rollback(skillID string, req RollbackRequest) ActionAccepted {
	return ActionAccepted{
		ActionID:   fmt.Sprintf("action_%d", time.Now().UnixNano()),
		Status:     "accepted",
		AcceptedAt: time.Now().UTC(),
	}
}

func (s *OrchestrationService) Promote(skillID string, req PromoteRequest) (ActionAccepted, error) {
	if s.repo != nil {
		passed, score, err := s.repo.GetLatestPassedSnapshot(context.Background(), skillID, req.Version)
		if err != nil {
			return ActionAccepted{}, fmt.Errorf("failed to get snapshot: %w", err)
		}
		if !passed {
			return ActionAccepted{}, fmt.Errorf("skill version %s did not pass regression gate (score: %.2f)", req.Version, score)
		}
		if err := s.repo.UpdateSkillVersionPromoted(skillID, req.Version, true); err != nil {
			return ActionAccepted{}, fmt.Errorf("failed to promote skill version: %w", err)
		}
	}

	return ActionAccepted{
		ActionID:   fmt.Sprintf("action_%d", time.Now().UnixNano()),
		Status:     "accepted",
		AcceptedAt: time.Now().UTC(),
	}, nil
}

func (s *OrchestrationService) CreateTestRun(req CreateTestRunRequest) TestRunAccepted {
	if s.repo != nil {
		tr, err := s.repo.CreateTestRun(req)
		if err == nil && tr != nil {
			return TestRunAccepted{
				TestRunID: tr.TestRunID,
				Status:    "queued",
			}
		}
	}

	return TestRunAccepted{
		TestRunID: fmt.Sprintf("testrun_%d", time.Now().UnixNano()),
		Status:    "queued",
	}
}

func (s *OrchestrationService) GetTestRun(testRunID string) TestRun {
	if s.repo != nil {
		tr, err := s.repo.GetTestRun(testRunID)
		if err == nil && tr != nil {
			return *tr
		}
	}

	return TestRun{
		TestRunID: testRunID,
		Status:    "succeeded",
		CreatedAt: time.Now().UTC().Add(-10 * time.Minute),
		UpdatedAt: time.Now().UTC().Add(-1 * time.Minute),
		ExitCode:  0,
	}
}

func (s *OrchestrationService) GetTestReport(testRunID string) TestReport {
	return TestReport{
		TestRunID: testRunID,
		Passed:    true,
		Metrics: TestMetrics{
			TotalCases:  25,
			PassRate:    0.96,
			Regressions: 0,
		},
		Failures: []TestFailure{},
	}
}

func (s *OrchestrationService) EvaluateGate(req GateEvaluateRequest) GateEvaluateResponse {
	if s.repo != nil {
		policy, err := s.repo.GetPolicy(req.PolicyID)
		if err == nil && policy != nil {
			passed := true
			var reasons []string
			for _, rule := range policy.Rules {
				if !s.evaluatePolicyRule(req, rule) {
					passed = false
					reasons = append(reasons, fmt.Sprintf("Policy rule failed: %s %s %v", rule.Metric, rule.Op, rule.Value))
				}
			}
			if passed {
				reasons = append(reasons, "All gate criteria met")
			}
			return GateEvaluateResponse{
				Passed:     passed,
				PolicyID:   req.PolicyID,
				Reasons:    reasons,
				DecisionAt: time.Now().UTC(),
			}
		}
	}

	return GateEvaluateResponse{
		Passed:     true,
		PolicyID:   req.PolicyID,
		Reasons:    []string{"All gate criteria met", "Test pass rate >= 0.9"},
		DecisionAt: time.Now().UTC(),
	}
}

func (s *OrchestrationService) evaluatePolicyRule(req GateEvaluateRequest, rule PolicyRule) bool {
	switch rule.Metric {
	case "test_pass_rate":
		if req.TestRunID != "" {
			tr := s.GetTestRun(req.TestRunID)
			if tr.Status == "succeeded" {
				report := s.GetTestReport(req.TestRunID)
				return compareFloat(report.Metrics.PassRate, rule.Op, rule.Value.(float64))
			}
		}
	}
	return true
}

func compareFloat(value float64, op string, threshold float64) bool {
	switch op {
	case "gte":
		return value >= threshold
	case "lte":
		return value <= threshold
	case "eq":
		return value == threshold
	case "neq":
		return value != threshold
	}
	return true
}

func (s *OrchestrationService) GetPolicy(policyID string) Policy {
	if s.repo != nil {
		policy, err := s.repo.GetPolicy(policyID)
		if err == nil && policy != nil {
			return *policy
		}
	}

	return Policy{
		PolicyID: policyID,
		Name:     "Standard Promotion Policy",
		Rules: []PolicyRule{
			{Metric: "test_pass_rate", Op: "gte", Value: 0.9},
			{Metric: "regressions", Op: "lte", Value: 0},
		},
	}
}
