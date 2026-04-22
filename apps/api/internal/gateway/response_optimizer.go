package gateway

import (
	"encoding/json"
	"fmt"
)

type ResponseOptimizer struct {
	defaultMaxLength int
}

func NewResponseOptimizer() *ResponseOptimizer {
	return &ResponseOptimizer{
		defaultMaxLength: 4000,
	}
}

type OptimizationResult struct {
	Optimized    bool                   `json:"optimized"`
	OriginalSize int                    `json:"original_size"`
	OptimizedSize int                   `json:"optimized_size"`
	Data         map[string]interface{} `json:"data"`
	Truncated    []string               `json:"truncated_fields,omitempty"`
	AppliedRules []string               `json:"applied_rules,omitempty"`
}

func (o *ResponseOptimizer) Optimize(data map[string]interface{}, rules []OutputRule) (*OptimizationResult, error) {
	result := &OptimizationResult{
		OriginalSize: estimateSize(data),
		Data:         data,
		Truncated:    []string{},
		AppliedRules: []string{},
	}

	for _, rule := range rules {
		switch rule.Type {
		case "pick":
			result.Data = o.pickFields(result.Data, rule.Fields)
			result.AppliedRules = append(result.AppliedRules, fmt.Sprintf("pick:%v", rule.Fields))
		case "omit":
			result.Data = o.omitFields(result.Data, rule.Fields)
			result.AppliedRules = append(result.AppliedRules, fmt.Sprintf("omit:%v", rule.Fields))
		case "truncate":
			o.truncateFields(result.Data, rule.Fields, rule.MaxLength)
			result.Truncated = append(result.Truncated, rule.Fields...)
			result.AppliedRules = append(result.AppliedRules, fmt.Sprintf("truncate:%v:%d", rule.Fields, rule.MaxLength))
		case "summarize":
			o.summarize(result.Data, rule.MaxLength)
			result.AppliedRules = append(result.AppliedRules, fmt.Sprintf("summarize:%d", rule.MaxLength))
		}
	}

	result.OptimizedSize = estimateSize(result.Data)
	result.Optimized = result.OptimizedSize < result.OriginalSize

	return result, nil
}

func (o *ResponseOptimizer) pickFields(data map[string]interface{}, fields []string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, field := range fields {
		if v, ok := data[field]; ok {
			result[field] = v
		}
	}
	if len(result) == 0 {
		return data
	}
	return result
}

func (o *ResponseOptimizer) omitFields(data map[string]interface{}, fields []string) map[string]interface{} {
	result := make(map[string]interface{})
	fieldSet := make(map[string]bool)
	for _, f := range fields {
		fieldSet[f] = true
	}
	for k, v := range data {
		if !fieldSet[k] {
			result[k] = v
		}
	}
	return result
}

func (o *ResponseOptimizer) truncateFields(data map[string]interface{}, fields []string, maxLength int) {
	for _, field := range fields {
		if v, ok := data[field]; ok {
			if str, ok := v.(string); ok {
				if len(str) > maxLength {
					data[field] = str[:maxLength] + "..."
				}
			}
		}
	}
}

func (o *ResponseOptimizer) summarize(data map[string]interface{}, maxItems int) {
	for k, v := range data {
		if arr, ok := v.([]interface{}); ok {
			if len(arr) > maxItems {
				data[k] = append(arr[:maxItems], fmt.Sprintf("... %d more items", len(arr)-maxItems))
			}
		}
	}
}

func estimateSize(data map[string]interface{}) int {
	b, _ := json.Marshal(data)
	return len(b)
}

type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
}

func (o *ResponseOptimizer) OptimizeHTTPResponse(resp *HTTPResponse, rules []OutputRule) (*HTTPResponse, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return resp, nil
	}

	result, err := o.Optimize(data, rules)
	if err != nil {
		return resp, err
	}

	optimizedBody, err := json.Marshal(result.Data)
	if err != nil {
		return resp, err
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       optimizedBody,
		Headers:    resp.Headers,
	}, nil
}