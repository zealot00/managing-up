package optimizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/zealot/managing-up/apps/api/internal/llm"
)

type ResponseOptimizer struct {
	whitelist    []string
	blacklist    []string
	maxBytes     int64
	useSummary   bool
	summaryModel llm.Client
	summaryThreshold int64
}

type OptimizeOptions struct {
	Whitelist   []string
	Blacklist   []string
	MaxBytes    int64
	UseSummary  bool
	SummaryModel llm.Client
	SummaryThreshold int64
}

func NewResponseOptimizer(opts OptimizeOptions) *ResponseOptimizer {
	return &ResponseOptimizer{
		whitelist:    opts.Whitelist,
		blacklist:    opts.Blacklist,
		maxBytes:     opts.MaxBytes,
		useSummary:   opts.UseSummary,
		summaryModel: opts.SummaryModel,
		summaryThreshold: opts.SummaryThreshold,
	}
}

func (o *ResponseOptimizer) OptimizeResponse(body []byte) ([]byte, error) {
	if len(body) == 0 {
		return body, nil
	}

	data, err := o.parseBody(body)
	if err != nil {
		return body, nil
	}

	if len(o.blacklist) > 0 {
		data = o.applyBlacklist(data)
	}

	if len(o.whitelist) > 0 {
		data = o.applyWhitelist(data)
	}

	if o.maxBytes > 0 && int64(len(body)) > o.maxBytes {
		data = o.truncateToSize(data, o.maxBytes)
	}

	result, err := json.Marshal(data)
	if err != nil {
		return body, nil
	}

	return result, nil
}

func (o *ResponseOptimizer) parseBody(body []byte) (any, error) {
	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (o *ResponseOptimizer) applyBlacklist(data any) any {
	switch v := data.(type) {
	case map[string]any:
		result := make(map[string]any)
		for k, val := range v {
			if !o.isBlacklisted(k) {
				result[k] = o.applyBlacklist(val)
			}
		}
		return result
	case []any:
		result := make([]any, 0, len(v))
		for _, item := range v {
			result = append(result, o.applyBlacklist(item))
		}
		return result
	default:
		return v
	}
}

func (o *ResponseOptimizer) isBlacklisted(field string) bool {
	for _, blocked := range o.blacklist {
		if field == blocked || strings.HasPrefix(field, blocked) {
			return true
		}
	}
	return false
}

func (o *ResponseOptimizer) applyWhitelist(data any) any {
	if len(o.whitelist) == 0 {
		return data
	}

	switch v := data.(type) {
	case map[string]any:
		result := make(map[string]any)
		for _, allowed := range o.whitelist {
			if val, ok := v[allowed]; ok {
				result[allowed] = o.applyWhitelist(val)
			}
		}
		return result
	case []any:
		result := make([]any, 0, len(v))
		for _, item := range v {
			result = append(result, o.applyWhitelist(item))
		}
		return result
	default:
		return v
	}
}

func (o *ResponseOptimizer) truncateToSize(data any, maxBytes int64) any {
	switch v := data.(type) {
	case string:
		if int64(len(v)) > maxBytes {
			return v[:maxBytes]
		}
		return v
	case map[string]any:
		result := make(map[string]any)
		var totalSize int64

		for k, val := range v {
			encoded, _ := json.Marshal(val)
			if totalSize+int64(len(encoded)) > maxBytes {
				break
			}
			result[k] = val
			totalSize += int64(len(encoded))
		}
		return result
	case []any:
		result := make([]any, 0, len(v))
		var totalSize int64

		for _, item := range v {
			encoded, _ := json.Marshal(item)
			if totalSize+int64(len(encoded)) > maxBytes {
				break
			}
			result = append(result, item)
			totalSize += int64(len(encoded))
		}
		return result
	default:
		return v
	}
}

func (o *ResponseOptimizer) GenerateSummary(ctx context.Context, body []byte) (string, error) {
	if o.summaryModel == nil || !o.useSummary {
		return "", nil
	}

	if int64(len(body)) < o.summaryThreshold {
		return "", nil
	}

	messages := []llm.Message{
		{Role: "user", Content: fmt.Sprintf("Summarize the following JSON response concisely (max 100 words):\n%s", string(body))},
	}

	resp, err := o.summaryModel.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

type RequestBuilder struct {
	baseURL    string
	path       string
	method     string
	headers    map[string]string
	timeout    int
}

func NewRequestBuilder(baseURL, path, method string) *RequestBuilder {
	return &RequestBuilder{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		path:    path,
		method:  method,
		headers: make(map[string]string),
	}
}

func (b *RequestBuilder) AddHeader(key, value string) *RequestBuilder {
	b.headers[key] = value
	return b
}

func (b *RequestBuilder) SetTimeout(seconds int) *RequestBuilder {
	b.timeout = seconds
	return b
}

func (b *RequestBuilder) Build(params map[string]string, body io.Reader) (*http.Request, error) {
	url := b.baseURL + b.interpolatePath(params)

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(b.method, url, body)
	} else {
		req, err = http.NewRequest(b.method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	for key, value := range b.headers {
		req.Header.Set(key, b.interpolateValue(value, params))
	}

	return req, nil
}

func (b *RequestBuilder) interpolatePath(params map[string]string) string {
	result := b.path

	for key, value := range params {
		placeholder := fmt.Sprintf("{%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

func (b *RequestBuilder) interpolateValue(value string, params map[string]string) string {
	result := value

	for key, val := range params {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, val)
	}

	return result
}

type FieldSelector struct {
	keepFields map[string]bool
}

func NewFieldSelector(fields []string) *FieldSelector {
	fs := &FieldSelector{
		keepFields: make(map[string]bool),
	}
	for _, f := range fields {
		fs.keepFields[f] = true
	}
	return fs
}

func (fs *FieldSelector) Select(data []byte) ([]byte, error) {
	var obj any
	if err := json.Unmarshal(data, &obj); err != nil {
		return data, err
	}

	selected := fs.selectRecursive(obj)
	return json.Marshal(selected)
}

func (fs *FieldSelector) selectRecursive(data any) any {
	switch v := data.(type) {
	case map[string]any:
		result := make(map[string]any)
		for key, val := range v {
			if fs.keepFields[key] {
				result[key] = fs.selectRecursive(val)
			}
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = fs.selectRecursive(item)
		}
		return result
	default:
		return v
	}
}

func measureSize(data any) int64 {
	switch v := data.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	case []any:
		var total int64
		for _, item := range v {
			total += measureSize(item)
		}
		return total
	case map[string]any:
		var total int64
		for key, val := range v {
			total += int64(len(key))
			total += measureSize(val)
		}
		return total
	default:
		dataBytes, _ := json.Marshal(v)
		return int64(len(dataBytes))
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func filterFields(data map[string]any, fields []string) map[string]any {
	result := make(map[string]any)
	for _, f := range fields {
		if v, ok := data[f]; ok {
			result[f] = v
		}
	}
	return result
}

func readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
}

func mustMarshal(data any) []byte {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.Encode(data)
	return buf.Bytes()
}