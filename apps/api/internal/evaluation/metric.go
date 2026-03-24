package evaluation

import (
	"context"
)

type Score struct {
	Value   float64
	Details map[string]any
}

type MetricEvaluator interface {
	Name() string
	Evaluate(ctx context.Context, input any, expected any, output any) (Score, error)
}

type ExactMatchEvaluator struct{}

func (e *ExactMatchEvaluator) Name() string {
	return "exact_match"
}

func (e *ExactMatchEvaluator) Evaluate(ctx context.Context, input any, expected any, output any) (Score, error) {
	details := map[string]any{
		"expected": expected,
		"actual":   output,
	}

	match := output == expected

	return Score{
		Value:   boolToFloat(match),
		Details: details,
	}, nil
}

type SemanticSimilarityEvaluator struct {
	threshold float64
}

func NewSemanticSimilarityEvaluator(threshold float64) *SemanticSimilarityEvaluator {
	return &SemanticSimilarityEvaluator{threshold: threshold}
}

func (e *SemanticSimilarityEvaluator) Name() string {
	return "semantic_similarity"
}

func (e *SemanticSimilarityEvaluator) Evaluate(ctx context.Context, input any, expected any, output any) (Score, error) {
	expectedStr, ok1 := expected.(string)
	outputStr, ok2 := output.(string)

	if !ok1 || !ok2 {
		return Score{Value: 0, Details: map[string]any{"error": "non-string values provided"}}, nil
	}

	similarity := calculateCosineSimilarity(expectedStr, outputStr)
	pass := similarity >= e.threshold

	return Score{
		Value: similarity,
		Details: map[string]any{
			"similarity": similarity,
			"threshold":  e.threshold,
			"passed":     pass,
		},
	}, nil
}

type JudgeModelEvaluator struct {
	judgeFn PromptBasedJudge
}

func NewJudgeModelEvaluator(judgeFn PromptBasedJudge) *JudgeModelEvaluator {
	return &JudgeModelEvaluator{judgeFn: judgeFn}
}

func (e *JudgeModelEvaluator) Name() string {
	return "judge_model"
}

func (e *JudgeModelEvaluator) Evaluate(ctx context.Context, input any, expected any, output any) (Score, error) {
	score, err := e.judgeFn(ctx, input, expected, output)
	if err != nil {
		return Score{}, err
	}
	return Score{Value: score, Details: map[string]any{"method": "judge_model"}}, nil
}

type PromptBasedJudge func(ctx context.Context, input any, expected any, output any) (float64, error)

type EvaluatorRegistry struct {
	evaluators map[string]MetricEvaluator
}

func NewEvaluatorRegistry() *EvaluatorRegistry {
	return &EvaluatorRegistry{
		evaluators: make(map[string]MetricEvaluator),
	}
}

func (r *EvaluatorRegistry) Register(evaluator MetricEvaluator) {
	r.evaluators[evaluator.Name()] = evaluator
}

func (r *EvaluatorRegistry) Get(name string) (MetricEvaluator, bool) {
	evaluator, ok := r.evaluators[name]
	return evaluator, ok
}

func (r *EvaluatorRegistry) List() []MetricEvaluator {
	evaluators := make([]MetricEvaluator, 0, len(r.evaluators))
	for _, e := range r.evaluators {
		evaluators = append(evaluators, e)
	}
	return evaluators
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func calculateCosineSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	if s1 == "" || s2 == "" {
		return 0.0
	}

	words1 := wordSet(s1)
	words2 := wordSet(s2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	intersection := 0
	for w := range words1 {
		if words2[w] {
			intersection++
		}
	}

	norm1 := float64(len(words1))
	norm2 := float64(len(words2))

	return float64(intersection) / (norm1 * norm2 / float64(intersection+1))
}

func wordSet(s string) map[string]bool {
	words := make(map[string]bool)
	current := ""
	for _, c := range s {
		if c == ' ' || c == ',' || c == '.' || c == '\n' || c == '\t' {
			if current != "" {
				words[current] = true
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		words[current] = true
	}
	return words
}
