## Agent Harness Phase 1 Decisions

### Judge Model Evaluator Architecture (2026-03-25)
**Decision:** Implement judge_model as a `PromptBasedJudge` function closure created by `NewPromptBasedJudge`

**Rationale:**
- Allows flexible LLM judge implementation without coupling to specific providers
- The `PromptBasedJudge` function type is already defined in the existing `JudgeModelEvaluator`
- Factory pattern (`NewPromptBasedJudge`) enables dependency injection of LLM client

**Alternative considered:** Passing metric config at evaluation time
- Rejected because `MetricEvaluator.Evaluate()` interface doesn't support config parameter
- Would require interface change affecting all evaluators

**Result:** Judge is registered at runner initialization with default/configured LLM client

