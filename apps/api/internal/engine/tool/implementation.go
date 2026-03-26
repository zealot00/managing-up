package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// CLITool implements Tool for CLI commands.
type CLITool struct {
	def        *ToolDefinition
	binaryPath string
	args       []string
}

func NewCLITool(def *ToolDefinition) *CLITool {
	return &CLITool{
		def:        def,
		binaryPath: def.Impl.BinaryPath,
	}
}

func (t *CLITool) Name() string        { return t.def.Name }
func (t *CLITool) Description() string { return t.def.Description }

func (t *CLITool) Execute(ctx context.Context, args map[string]any) (any, error) {
	cliArgs, err := t.buildArgs(args)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, t.binaryPath, cliArgs...)
	output, err := cmd.Output()
	if err != nil {
		return nil, &Error{Code: "CLI_EXEC_FAILED", Message: err.Error()}
	}

	var result any
	if err := json.Unmarshal(output, &result); err != nil {
		return string(output), nil
	}
	return result, nil
}

func (t *CLITool) buildArgs(args map[string]any) ([]string, error) {
	var parts []string
	for k, v := range args {
		parts = append(parts, "--"+k)
		parts = append(parts, fmt.Sprintf("%v", v))
	}
	return parts, nil
}

// HTTPTool implements Tool for HTTP requests.
type HTTPTool struct {
	def     *ToolDefinition
	baseURL string
}

func NewHTTPTool(def *ToolDefinition) *HTTPTool {
	return &HTTPTool{
		def:     def,
		baseURL: def.Impl.BaseURL,
	}
}

func (t *HTTPTool) Name() string        { return t.def.Name }
func (t *HTTPTool) Description() string { return t.def.Description }

func (t *HTTPTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	url := t.baseURL
	for k, v := range args {
		url = strings.ReplaceAll(url, "{"+k+"}", fmt.Sprintf("%v", v))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// ScriptTool implements Tool for scripts.
type ScriptTool struct {
	def         *ToolDefinition
	interpreter string
	scriptPath  string
}

func NewScriptTool(def *ToolDefinition) *ScriptTool {
	return &ScriptTool{
		def:         def,
		interpreter: def.Impl.Interpreter,
		scriptPath:  def.Impl.ScriptPath,
	}
}

func (t *ScriptTool) Name() string        { return t.def.Name }
func (t *ScriptTool) Description() string { return t.def.Description }

func (t *ScriptTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	cmd := exec.CommandContext(ctx, t.interpreter, t.scriptPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(output, &result); err != nil {
		return string(output), nil
	}
	return result, nil
}
