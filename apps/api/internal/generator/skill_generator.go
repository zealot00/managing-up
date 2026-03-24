package generator

import (
	"context"
	"fmt"

	"github.com/zealot/managing-up/apps/api/internal/llm"
)

type Generator struct {
	llmClient llm.Client
	parser    *YAMLParser
}

func NewGenerator(llmClient llm.Client) *Generator {
	return &Generator{
		llmClient: llmClient,
		parser:    NewYAMLParser(),
	}
}

func (g *Generator) GenerateFromSOP(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	prompt := buildSOPPrompt(req.SOPText, req.Description)

	resp, err := g.llmClient.Generate(ctx, []llm.Message{
		{Role: "system", Content: systemPrompt()},
		{Role: "user", Content: prompt},
	}, llm.WithTemperature(0.3), llm.WithJSONResponse())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLMCallFailed, err)
	}

	spec, err := g.parser.Parse(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse skill spec: %w", ErrInvalidSpec, err)
	}

	if err := g.validate(spec); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}

	return &GenerateResponse{
		SpecYAML: resp.Content,
		Spec:     spec,
		Provider: resp.Provider,
		Model:    resp.Model,
		Usage:    &resp.Usage,
	}, nil
}

type GenerateRequest struct {
	SOPText     string
	Description string
}

type GenerateResponse struct {
	SpecYAML string
	Spec     *SkillSpec
	Provider llm.Provider
	Model    llm.Model
	Usage    *llm.Usage
}
