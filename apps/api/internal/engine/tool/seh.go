package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type SEHTool struct {
	def        *ToolDefinition
	binaryPath string
}

func NewSEHTool(binaryPath string) *SEHTool {
	return &SEHTool{
		binaryPath: binaryPath,
		def: &ToolDefinition{
			Name:     "seh",
			Category: CategoryCLI,
			Version:  "1.0",
			Impl: ToolImplementation{
				Type:       "cli",
				BinaryPath: binaryPath,
				Command:    "seh",
			},
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]SchemaProp{
					"action": {
						Type:        "string",
						Description: "Action to perform: run, score, gate, compare, report, drift, matrix",
					},
					"skill": {
						Type:        "string",
						Description: "Skill name to execute (for run action)",
					},
					"cases": {
						Type:        "string",
						Description: "Path to evaluation cases directory (for run action)",
					},
					"out": {
						Type:        "string",
						Description: "Output file path",
					},
					"run": {
						Type:        "string",
						Description: "Path to run result JSON (for score/gate/report actions)",
					},
					"baseline": {
						Type:        "string",
						Description: "Path to baseline run JSON (for compare action)",
					},
					"policy": {
						Type:        "string",
						Description: "Path to policy YAML file (for gate action)",
					},
					"in": {
						Type:        "string",
						Description: "Input file path (for report action)",
					},
					"workers": {
						Type:        "int",
						Description: "Number of concurrent workers",
					},
					"case_timeout": {
						Type:        "string",
						Description: "Per-case timeout (e.g., 250ms, 2s)",
					},
					"max_retries": {
						Type:        "int",
						Description: "Maximum retry attempts",
					},
					"tag": {
						Type:        "string",
						Description: "Tag to filter cases",
					},
					"sample": {
						Type:        "float",
						Description: "Random sample ratio (0.0-1.0)",
					},
					"seed": {
						Type:        "int",
						Description: "Deterministic replay seed",
					},
				},
				Required: []string{"action"},
			},
			Description: "Skill Eval Harness - 执行评测、评分、门禁判断、回归比较的 CLI 工具",
		},
	}
}

func (t *SEHTool) Name() string        { return "seh" }
func (t *SEHTool) Description() string { return t.def.Description }

func (t *SEHTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	action, _ := args["action"].(string)
	if action == "" {
		return nil, fmt.Errorf("action is required")
	}

	cmdArgs, err := t.buildArgs(action, args)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, t.binaryPath, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   string(output),
			"action":  action,
		}, nil
	}

	return t.parseOutput(action, output)
}

func (t *SEHTool) buildArgs(action string, args map[string]any) ([]string, error) {
	var parts []string

	switch action {
	case "run":
		parts = append(parts, "run")
		if skill, ok := args["skill"].(string); ok && skill != "" {
			parts = append(parts, "--skill", skill)
		}
		if cases, ok := args["cases"].(string); ok && cases != "" {
			parts = append(parts, "--cases", cases)
		}
		if out, ok := args["out"].(string); ok && out != "" {
			parts = append(parts, "--out", out)
		}

	case "score":
		parts = append(parts, "score")
		if run, ok := args["run"].(string); ok && run != "" {
			parts = append(parts, "--run", run)
		}
		if out, ok := args["out"].(string); ok && out != "" {
			parts = append(parts, "--out", out)
		}

	case "gate":
		parts = append(parts, "gate")
		if report, ok := args["report"].(string); ok && report != "" {
			parts = append(parts, "--report", report)
		}
		if policy, ok := args["policy"].(string); ok && policy != "" {
			parts = append(parts, "--policy", policy)
		}

	case "compare":
		parts = append(parts, "compare")
		if run, ok := args["run"].(string); ok && run != "" {
			parts = append(parts, "--run", run)
		}
		if baseline, ok := args["baseline"].(string); ok && baseline != "" {
			parts = append(parts, "--baseline", baseline)
		}

	case "report":
		parts = append(parts, "report")
		if in, ok := args["in"].(string); ok && in != "" {
			parts = append(parts, "--in", in)
		}
		if out, ok := args["out"].(string); ok && out != "" {
			parts = append(parts, "--out", out)
		}

	case "drift":
		parts = append(parts, "drift")
		if run, ok := args["run"].(string); ok && run != "" {
			parts = append(parts, "--run", run)
		}
		if baseline, ok := args["baseline"].(string); ok && baseline != "" {
			parts = append(parts, "--baseline", baseline)
		}
		if out, ok := args["out"].(string); ok && out != "" {
			parts = append(parts, "--out", out)
		}

	case "matrix":
		parts = append(parts, "matrix")
		if runtimes, ok := args["runtimes"].(string); ok && runtimes != "" {
			parts = append(parts, "--runtimes", runtimes)
		}
		if cases, ok := args["cases"].(string); ok && cases != "" {
			parts = append(parts, "--cases", cases)
		}
		if out, ok := args["out"].(string); ok && out != "" {
			parts = append(parts, "--out", out)
		}

	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}

	if workers, ok := args["workers"].(int); ok && workers > 0 {
		parts = append(parts, "--workers", fmt.Sprintf("%d", workers))
	}
	if timeout, ok := args["case_timeout"].(string); ok && timeout != "" {
		parts = append(parts, "--case-timeout", timeout)
	}
	if retries, ok := args["max_retries"].(int); ok && retries > 0 {
		parts = append(parts, "--max-retries", fmt.Sprintf("%d", retries))
	}
	if tag, ok := args["tag"].(string); ok && tag != "" {
		parts = append(parts, "--tag", tag)
	}
	if sample, ok := args["sample"].(float64); ok && sample > 0 {
		parts = append(parts, "--sample", fmt.Sprintf("%f", sample))
	}
	if seed, ok := args["seed"].(int); ok && seed > 0 {
		parts = append(parts, "--seed", fmt.Sprintf("%d", seed))
	}

	return parts, nil
}

func (t *SEHTool) parseOutput(action string, output []byte) (any, error) {
	var result any

	if err := json.Unmarshal(output, &result); err != nil {
		text := strings.TrimSpace(string(output))
		return map[string]any{
			"success": true,
			"action":  action,
			"output":  text,
		}, nil
	}

	return map[string]any{
		"success": true,
		"action":  action,
		"result":  result,
		"raw":     string(output),
	}, nil
}

func (t *SEHTool) Definition() *ToolDefinition {
	return t.def
}

func RegisterSEHTool(registry *Registry, binaryPath string) error {
	tool := NewSEHTool(binaryPath)
	return registry.Register(tool, tool.Definition())
}
