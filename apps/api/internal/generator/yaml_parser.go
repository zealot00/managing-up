package generator

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type YAMLParser struct{}

func NewYAMLParser() *YAMLParser {
	return &YAMLParser{}
}

func (p *YAMLParser) Parse(yamlContent string) (*SkillSpec, error) {
	if yamlContent == "" {
		return nil, fmt.Errorf("%w: empty yaml content", ErrInvalidSpec)
	}

	var spec SkillSpec
	if err := yaml.Unmarshal([]byte(yamlContent), &spec); err != nil {
		return nil, fmt.Errorf("%w: failed to parse yaml: %v", ErrInvalidSpec, err)
	}

	if spec.Name == "" {
		return nil, fmt.Errorf("%w: missing name", ErrInvalidSpec)
	}
	if spec.Version == "" {
		return nil, fmt.Errorf("%w: missing version", ErrInvalidSpec)
	}
	if len(spec.Steps) == 0 {
		return nil, fmt.Errorf("%w: no steps defined", ErrInvalidSpec)
	}

	for i, step := range spec.Steps {
		if step.ID == "" {
			return nil, fmt.Errorf("%w: step at index %d missing id", ErrInvalidSpec, i)
		}
		if step.Type == "" {
			return nil, fmt.Errorf("%w: step %s missing type", ErrInvalidSpec, step.ID)
		}
		if step.Type == "tool" && step.ToolRef == "" {
			return nil, fmt.Errorf("%w: tool step %s missing tool_ref", ErrInvalidSpec, step.ID)
		}
		if step.Type == "approval" && step.ApproverGroup == "" {
			return nil, fmt.Errorf("%w: approval step %s missing approver_group", ErrInvalidSpec, step.ID)
		}
	}

	return &spec, nil
}
