package bridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/bridge/optimizer"
	"github.com/zealot/managing-up/apps/api/internal/bridge/parser"
	"github.com/zealot/managing-up/apps/api/internal/bridge/template"
)

type BridgeAdapter struct {
	template   *template.AdapterTemplate
	parser     *parser.OpenAPISpec
	optimizer  *optimizer.ResponseOptimizer
	httpClient *http.Client
}

type AdapterConfig struct {
	Template      *template.AdapterTemplate
	OpenAPISpec   []byte
	HTTPClient    *http.Client
	OptimizerOpts optimizer.OptimizeOptions
}

func NewBridgeAdapter(cfg AdapterConfig) (*BridgeAdapter, error) {
	spec, err := parser.ParseOpenAPI(cfg.OpenAPISpec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	opt := optimizer.NewResponseOptimizer(cfg.OptimizerOpts)

	return &BridgeAdapter{
		template:   cfg.Template,
		parser:     spec,
		optimizer:  opt,
		httpClient: httpClient,
	}, nil
}

func (a *BridgeAdapter) GetTools() []template.MCPTool {
	return a.template.GenerateTools(a.parser.ParseEndpoints())
}

func (a *BridgeAdapter) GetBaseURL() string {
	return a.parser.GetBaseURL()
}

func (a *BridgeAdapter) CallTool(ctx context.Context, toolName string, input map[string]any) (map[string]any, error) {
	tools := a.GetTools()
	var selectedTool *template.MCPTool

	for i := range tools {
		if tools[i].Name == toolName {
			selectedTool = &tools[i]
			break
		}
	}

	if selectedTool == nil {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	params := a.template.ApplyInputMapping(toolName, input)

	req, err := a.buildRequest(selectedTool, params, input)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := a.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if a.optimizer != nil {
		optimized, err := a.optimizer.OptimizeResponse(body)
		if err == nil {
			body = optimized
		}
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return map[string]any{"raw": string(body)}, nil
	}

	return result, nil
}

func (a *BridgeAdapter) buildRequest(tool *template.MCPTool, params map[string]string, input map[string]any) (*http.Request, error) {
	path := tool.Endpoint.Path

	for key, value := range params {
		path = strings.ReplaceAll(path, fmt.Sprintf("{%s}", key), value)
	}

	url := combineURL(a.GetBaseURL(), path)

	var body io.Reader
	if hasBody(tool) {
		bodyBytes, _ := json.Marshal(input)
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(tool.Endpoint.Method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func combineURL(baseURL, path string) string {
	base := baseURL
	if base == "" {
		base = "http://localhost"
	}
	return fmt.Sprintf("%s%s", strings.TrimSuffix(base, "/"), path)
}

func hasBody(tool *template.MCPTool) bool {
	return tool.Endpoint.Method == "POST" || tool.Endpoint.Method == "PUT" || tool.Endpoint.Method == "PATCH"
}