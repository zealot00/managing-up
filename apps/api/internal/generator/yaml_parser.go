package generator

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type YAMLParser struct{}

func NewYAMLParser() *YAMLParser {
	return &YAMLParser{}
}

func Sanitize(input string) string {
	if input == "" {
		return ""
	}

	codeFenceRegex := regexp.MustCompile("(?m)^[\\t ]*```[a-zA-Z]*\\s*")
	input = codeFenceRegex.ReplaceAllString(input, "")
	input = strings.TrimSpace(input)

	// Remove reasoning tags like <think> and </think>
	reasoningRegex := regexp.MustCompile("(?s)<[^>]+>")
	input = reasoningRegex.ReplaceAllString(input, "")
	input = strings.TrimSpace(input)

	lines := strings.Split(input, "\n")
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" || strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if idx := strings.Index(line, ":"); idx > 0 {
			afterColon := strings.TrimSpace(line[idx+1:])
			if afterColon != "" &&
				!strings.HasPrefix(afterColon, "\"") &&
				!strings.HasPrefix(afterColon, "'") &&
				!strings.HasPrefix(afterColon, "[") &&
				!strings.HasPrefix(afterColon, "{") {
				boolRegex := regexp.MustCompile("^(true|false|null|~)$")
				if !boolRegex.MatchString(afterColon) && strings.Contains(afterColon, ":") {
					quoted := fmt.Sprintf("%s: \"%s\"", line[:idx], strings.ReplaceAll(afterColon, "\"", "\\\""))
					lines[i] = quoted
				}
			}
		}
	}
	input = strings.Join(lines, "\n")

	input = strings.ReplaceAll(input, "\t", "  ")
	indentRegex := regexp.MustCompile("(?m)^( {4,})")
	input = indentRegex.ReplaceAllStringFunc(input, func(match string) string {
		spaces := len(match)
		return strings.Repeat(" ", spaces/2*2)
	})

	return input
}

func (p *YAMLParser) Parse(yamlContent string) (*SkillSpec, error) {
	if yamlContent == "" {
		return nil, fmt.Errorf("%w: empty yaml content", ErrInvalidSpec)
	}

	sanitized := Sanitize(yamlContent)

	var spec SkillSpec
	if err := yaml.Unmarshal([]byte(sanitized), &spec); err != nil {
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
