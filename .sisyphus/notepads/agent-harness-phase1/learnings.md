## Agent Harness Phase 1 Notepad
**Created:** 2026-03-25
**Plan:** docs/todo-plan.md

## Conventions

## Gotchas

## Decisions

## Problems

## Judge Model Implementation (2026-03-25)

### Pattern: PromptBasedJudge Function Type
- `PromptBasedJudge` is a function type: `func(ctx context.Context, input any, expected any, output any) (float64, error)`
- `JudgeModelEvaluator` wraps a `PromptBasedJudge` and implements `MetricEvaluator` interface
- `NewPromptBasedJudge(client llm.Client, config *JudgePromptConfig) PromptBasedJudge` creates the judge function

### Judge Prompt Template
- Uses Go template format with `{{.TaskDescription}}`, `{{.ExpectedOutput}}`, `{{.ActualOutput}}`
- Prompts LLM to return JSON with `score` (0-1) and `reasoning`
- Score is parsed using regex extraction from JSON response

### Configuration via Environment Variables
- `JUDGE_PROVIDER`: LLM provider (default: openai)
- `JUDGE_MODEL`: Model name (default: gpt-4o-mini)
- `JUDGE_API_KEY`: API key for the judge LLM

### Integration in EvaluationRunner
- Judge is created and registered in `NewEvaluationRunner`
- Falls back gracefully if judge client creation fails (judge_model won't be available)
- Existing evaluators (exact_match, semantic_similarity, embedding_similarity) unaffected

## Task Builder Implementation (2026-03-25)

### Architecture Pattern: Service Layer + Repository Adapter
- Task service defines `TaskRepository` interface in `service/task.go`
- `repoToTaskRepoAdapter` in `server/server.go` adapts `server.Repository` to `service.TaskRepository`
- This allows service layer to be persistence-agnostic

### New Types Added
- `BuildTaskFromTraceRequest` in `service/task.go` - request type for the new method
- `TaskExecutionSource` and `TaskTraceEvent` - internal types for trace extraction
- `BuildTaskFromTraceRequest` in `server/types.go` - API request type

### Data Extraction from Trace
- `execution_started` event: extracts `input` field
- `llm_call` event: extracts `model` and `output` fields
- `tool_output` event: extracts `output` field as fallback expected output

### Repository Interface Evolution
- When adding methods to `TaskRepository`, must also:
  1. Add to `server.Repository` interface
  2. Implement in `store.go` (in-memory)
  3. Implement in `postgres/repository.go`
  4. Add adapter methods in `repoToTaskRepoAdapter`

### Handler Pattern for New Endpoints
- Add route: `mux.HandleFunc("/api/v1/tasks/from-trace", srv.handleTaskFromTrace)`
- Handler goes in `server/server.go` with error mapping from service errors to HTTP responses

## Basic Sweep Engine Implementation (2026-03-25)

### New Types Added

**server/types.go:**
- `Variant` struct: defines a single sweep variant with Model, Prompt, Temperature, MaxTokens, Seed, SkillConfig
- `Variants []Variant` added to `CreateExperimentRequest`
- `Variants []Variant` added to `Experiment` 
- `VariantID string` added to `ExperimentRun`

**service/experiment.go:**
- `Variant` struct (mirrors server type)
- `Variants []Variant` added to `Experiment` and `CreateExperimentRequest`
- `VariantID` added to `ExperimentRun`
- Modified `RunExperiment()` to handle sweep mode (task×variant pairs) vs traditional mode (task×agent pairs)
- Modified `runSingleTask()` to accept variant parameter and set VariantID on run

### Sweep Execution Flow
1. `POST /api/v1/experiments` accepts optional `variants[]` array
2. If variants provided → create task×variant pairs for execution
3. If no variants → create task×agent pairs (traditional behavior)
4. Each (task, variant) pair runs independently with that variant's config
5. Results stored with VariantID to distinguish sweep runs

### VariantID Generation
- Uses variant.Name if provided
- Falls back to variant.Model if Name is empty
- Empty string for traditional (non-sweep) experiments

### Backward Compatibility
- Existing experiments without variants continue to work unchanged
- AgentIDs field still used for traditional experiments
- Variants field is optional (omitempty in JSON)

## Capability Diff Engine Implementation (2026-03-25)

### New Types Added

**server/types.go:**
- `CapabilityDiffRequest` - query params for capability diff API
- `CapabilityScore` - aggregated score per capability dimension
- `DiffResult` - complete diff result with scores, change, CI, p-value, verdict

**service/experiment.go:**
- `CapabilityDiff` struct - mirrors DiffResult for service layer
- `CalculateCapabilityDiff()` method - computes capability comparison between two experiments
- Helper functions: `aggregateScoresByCapability`, `hasCapability`, `computeMean`, `computeVariance`
- Statistical functions: `computePValue`, `computeConfidenceInterval`, `tDistCDF`, `tInv`

### API Endpoint
- Route: `GET /api/v1/capabilities/{name}/diff?experiment_a=...&experiment_b=...`
- Returns diff for a single capability (specified by {name} path param)
- experiment_a and experiment_b are experiment IDs to compare

### Diff Result Structure
```json
{
  "capability": "planning",
  "experiment_a": "exp_v1",
  "experiment_b": "exp_v2", 
  "score_a": 0.85,
  "score_b": 0.72,
  "score_change": 0.13,
  "percent_change": 18.06,
  "confidence_interval": [-0.05, 0.31],
  "p_value": 0.023,
  "verdict": "smarter",
  "summary": "Experiment A scored 0.85 vs B scored 0.72 (p=0.023)"
}
```

### Statistical Methods
- Uses two-sample t-test approximation for p-value
- 95% confidence interval using t-distribution critical values
- Verdict logic: p < 0.05 → statistically significant → smarter/dumber; else no_significant_change

### Implementation Notes
- Aggregates scores by matching task tags against capability name
- Tasks with matching tags contribute their OverallScore to the aggregate
- Requires at least 2 samples in each group for statistical significance
- Falls back to "no_significant_change" verdict when insufficient data
