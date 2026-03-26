package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPAdapter struct {
	def     *ToolDefinition
	baseURL string
}

func NewHTTPAdapter(def *ToolDefinition) *HTTPAdapter {
	return &HTTPAdapter{def: def, baseURL: def.Impl.BaseURL}
}

func (t *HTTPAdapter) Name() string        { return t.def.Name }
func (t *HTTPAdapter) Description() string { return t.def.Description }

func (t *HTTPAdapter) Execute(ctx context.Context, args map[string]any) (any, error) {
	method := t.getMethod(args)
	url := t.getURL(args)
	headers := t.getHeaders(args)
	body := t.getBody(args)
	timeout := t.getTimeout(args)

	fullURL := t.buildURL(url, args)

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, &Error{Code: "INVALID_BODY", Message: "failed to marshal request body", Err: err}
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, &Error{Code: "INVALID_REQUEST", Message: err.Error()}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	if timeout > 0 && timeout < 30 {
		client.Timeout = time.Duration(timeout) * time.Second
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, &Error{Code: "REQUEST_FAILED", Message: err.Error()}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &Error{Code: "READ_RESPONSE_FAILED", Message: err.Error()}
	}

	var result any
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			result = string(respBody)
		}
	}

	return map[string]any{
		"status":     resp.Status,
		"statusCode": resp.StatusCode,
		"headers":    resp.Header,
		"body":       result,
	}, nil
}

func (t *HTTPAdapter) getMethod(args map[string]any) string {
	if m, ok := args["method"].(string); ok && m != "" {
		return strings.ToUpper(m)
	}
	return http.MethodGet
}

func (t *HTTPAdapter) getURL(args map[string]any) string {
	if u, ok := args["url"].(string); ok && u != "" {
		return u
	}
	return ""
}

func (t *HTTPAdapter) buildURL(url string, args map[string]any) string {
	fullURL := url
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		fullURL = strings.TrimSuffix(t.baseURL, "/") + "/" + strings.TrimPrefix(url, "/")
	}
	for k, v := range args {
		fullURL = strings.ReplaceAll(fullURL, "{"+k+"}", fmt.Sprintf("%v", v))
	}
	return fullURL
}

func (t *HTTPAdapter) getHeaders(args map[string]any) map[string]string {
	headers := make(map[string]string)
	if h, ok := args["headers"].(map[string]any); ok {
		for k, v := range h {
			if vs, ok := v.(string); ok {
				headers[k] = vs
			} else {
				headers[k] = fmt.Sprintf("%v", v)
			}
		}
	}
	return headers
}

func (t *HTTPAdapter) getBody(args map[string]any) any {
	if b, ok := args["body"]; ok && b != nil {
		return b
	}
	return nil
}

func (t *HTTPAdapter) getTimeout(args map[string]any) int {
	if to, ok := args["timeout_seconds"]; ok {
		switch v := to.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			var i int
			fmt.Sscanf(v, "%d", &i)
			return i
		}
	}
	return 0
}
