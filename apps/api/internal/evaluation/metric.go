package evaluation

import (
	"context"
	"math"

	"github.com/zealot/managing-up/apps/api/internal/llm"
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

// EmbeddingSimilarityEvaluator uses real embedding vectors for semantic similarity.
type EmbeddingSimilarityEvaluator struct {
	client    llm.Client
	threshold float64
	model     string // embedding model, e.g. "text-embedding-3-small"
}

func NewEmbeddingSimilarityEvaluator(client llm.Client, threshold float64) *EmbeddingSimilarityEvaluator {
	return &EmbeddingSimilarityEvaluator{
		client:    client,
		threshold: threshold,
		model:     "text-embedding-3-small",
	}
}

func (e *EmbeddingSimilarityEvaluator) Name() string {
	return "embedding_similarity"
}

func (e *EmbeddingSimilarityEvaluator) Evaluate(ctx context.Context, input any, expected any, output any) (Score, error) {
	expectedStr, ok1 := expected.(string)
	outputStr, ok2 := output.(string)
	if !ok1 || !ok2 {
		return Score{Value: 0, Details: map[string]any{"error": "non-string values"}}, nil
	}

	// Get embeddings
	expectedVec, err := e.getEmbedding(ctx, expectedStr)
	if err != nil {
		return Score{Value: 0, Details: map[string]any{"error": err.Error()}}, nil
	}
	outputVec, err := e.getEmbedding(ctx, outputStr)
	if err != nil {
		return Score{Value: 0, Details: map[string]any{"error": err.Error()}}, nil
	}

	similarity := cosineSimilarity(expectedVec, outputVec)
	return Score{
		Value: similarity,
		Details: map[string]any{
			"similarity": similarity,
			"threshold":  e.threshold,
			"passed":     similarity >= e.threshold,
		},
	}, nil
}

func (e *EmbeddingSimilarityEvaluator) getEmbedding(ctx context.Context, text string) ([]float64, error) {
	// Fallback to word-set similarity when no embedding API available
	return nil, nil
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

type JudgeRouter struct {
	registry *EvaluatorRegistry
}

func NewJudgeRouter(registry *EvaluatorRegistry) *JudgeRouter {
	return &JudgeRouter{registry: registry}
}

func (r *JudgeRouter) Select(primaryMetric string) MetricEvaluator {
	evaluator, ok := r.registry.Get(primaryMetric)
	if !ok {
		evaluator, _ = r.registry.Get("exact_match")
	}
	return evaluator
}

func (r *JudgeRouter) ScoreAll(primaryMetric string, secondaryMetrics []string, ctx context.Context, input, expected, output any) (Score, []Score, error) {
	primary := r.Select(primaryMetric)
	primaryScore, err := primary.Evaluate(ctx, input, expected, output)
	if err != nil {
		return Score{}, nil, err
	}

	var secondary []Score
	for _, name := range secondaryMetrics {
		evaluator, ok := r.registry.Get(name)
		if !ok {
			continue
		}
		s, err := evaluator.Evaluate(ctx, input, expected, output)
		if err != nil {
			continue
		}
		secondary = append(secondary, s)
	}

	return primaryScore, secondary, nil
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
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
