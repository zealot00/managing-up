package optimizer

import (
	"encoding/json"
	"testing"
)

func TestResponseOptimizer_ApplyBlacklist(t *testing.T) {
	opt := NewResponseOptimizer(OptimizeOptions{
		Blacklist: []string{"password", "secret", "token"},
	})

	data := map[string]any{
		"username": "john",
		"password": "secret123",
		"email":    "john@example.com",
		"metadata": map[string]any{
			"token": "abc123",
			"role":  "admin",
		},
	}

	result := opt.applyBlacklist(data)

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected map result")
	}

	if _, exists := resultMap["password"]; exists {
		t.Error("password should be blacklisted")
	}

	if _, exists := resultMap["token"]; exists {
		t.Error("token should be blacklisted")
	}

	if _, exists := resultMap["username"]; !exists {
		t.Error("username should be present")
	}

	if _, exists := resultMap["email"]; !exists {
		t.Error("email should be present")
	}

	meta, ok := resultMap["metadata"].(map[string]any)
	if !ok {
		t.Fatal("expected nested metadata")
	}

	if _, exists := meta["token"]; exists {
		t.Error("nested token should be blacklisted")
	}
}

func TestResponseOptimizer_ApplyWhitelist(t *testing.T) {
	opt := NewResponseOptimizer(OptimizeOptions{
		Whitelist: []string{"id", "name", "email"},
	})

	data := map[string]any{
		"id":       "123",
		"name":     "John",
		"password": "secret",
		"internal": "hidden",
	}

	result := opt.applyWhitelist(data)

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected map result")
	}

	if _, exists := resultMap["id"]; !exists {
		t.Error("id should be whitelisted")
	}

	if _, exists := resultMap["name"]; !exists {
		t.Error("name should be whitelisted")
	}

	if _, exists := resultMap["password"]; exists {
		t.Error("password should not be whitelisted")
	}

	if _, exists := resultMap["internal"]; exists {
		t.Error("internal should not be whitelisted")
	}
}

func TestResponseOptimizer_TruncateToSize(t *testing.T) {
	opt := NewResponseOptimizer(OptimizeOptions{
		MaxBytes: 100,
	})

	data := map[string]any{
		"short": "value",
	}

	result := opt.truncateToSize(data, 100)

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected map result")
	}

	if _, exists := resultMap["short"]; !exists {
		t.Error("short field should be present")
	}
}

func TestFieldSelector_Select(t *testing.T) {
	fs := NewFieldSelector([]string{"id", "name", "email"})

	data := []byte(`{
		"id": "123",
		"name": "John",
		"email": "john@example.com",
		"password": "secret",
		"internal": "hidden"
	}`)

	result, err := fs.Select(data)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	var resultMap map[string]any
	if err := json.Unmarshal(result, &resultMap); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if _, exists := resultMap["id"]; !exists {
		t.Error("id should be selected")
	}

	if _, exists := resultMap["password"]; exists {
		t.Error("password should not be selected")
	}
}

func TestFieldSelector_SelectNested(t *testing.T) {
	fs := NewFieldSelector([]string{"id", "name", "items"})

	data := []byte(`{
		"id": "123",
		"name": "Order",
		"password": "secret",
		"items": [
			{"id": 1, "name": "Item 1", "price": 10},
			{"id": 2, "name": "Item 2", "price": 20}
		]
	}`)

	result, err := fs.Select(data)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	var resultMap map[string]any
	if err := json.Unmarshal(result, &resultMap); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if _, exists := resultMap["items"]; !exists {
		t.Error("items should be selected")
	}

	items, ok := resultMap["items"].([]any)
	if !ok {
		t.Fatal("items should be an array")
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestRequestBuilder_Build(t *testing.T) {
	builder := NewRequestBuilder("https://api.example.com", "/users/{user_id}/posts/{post_id}", "GET")
	builder.AddHeader("Authorization", "Bearer token")

	params := map[string]string{
		"user_id": "123",
		"post_id": "456",
	}

	req, err := builder.Build(params, nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if req.URL.Path != "/users/123/posts/456" {
		t.Errorf("expected path /users/123/posts/456, got %s", req.URL.Path)
	}

	if req.Header.Get("Authorization") != "Bearer token" {
		t.Errorf("expected Authorization header, got %s", req.Header.Get("Authorization"))
	}
}

func TestRequestBuilder_InterpolateValue(t *testing.T) {
	builder := NewRequestBuilder("", "/users/{user_id}", "GET")

	params := map[string]string{
		"user_id": "123",
		"filter":  "active",
	}

	value := "filter={{filter}}&sort=name"
	result := builder.interpolateValue(value, params)
	expected := "filter=active&sort=name"

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestMeasureSize(t *testing.T) {
	tests := []struct {
		name    string
		data    any
		minSize int64
		maxSize int64
	}{
		{
			name:    "string",
			data:    "hello",
			minSize: 5,
			maxSize: 6,
		},
		{
			name:    "map",
			data:    map[string]any{"key": "value"},
			minSize: 3,
			maxSize: 20,
		},
		{
			name:    "array",
			data:    []any{1, 2, 3},
			minSize: 1,
			maxSize: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := measureSize(tt.data)
			if size < tt.minSize || size > tt.maxSize {
				t.Errorf("measureSize() = %d, want between %d and %d", size, tt.minSize, tt.maxSize)
			}
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	if !contains(slice, "banana") {
		t.Error("should contain banana")
	}

	if contains(slice, "orange") {
		t.Error("should not contain orange")
	}
}

func TestFilterFields(t *testing.T) {
	data := map[string]any{
		"id":       "123",
		"name":     "John",
		"password": "secret",
		"email":    "john@example.com",
	}

	fields := []string{"id", "name", "email"}
	result := filterFields(data, fields)

	if len(result) != 3 {
		t.Errorf("expected 3 fields, got %d", len(result))
	}

	if _, exists := result["password"]; exists {
		t.Error("password should not be in result")
	}
}

func TestOptimizeResponse_EmptyBody(t *testing.T) {
	opt := NewResponseOptimizer(OptimizeOptions{})

	result, err := opt.OptimizeResponse([]byte{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty result, got %d bytes", len(result))
	}
}

func TestOptimizeResponse_InvalidJSON(t *testing.T) {
	opt := NewResponseOptimizer(OptimizeOptions{
		Whitelist: []string{"id"},
	})

	result, err := opt.OptimizeResponse([]byte("not json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(result) != "not json" {
		t.Errorf("expected original for invalid JSON, got %s", string(result))
	}
}

func TestOptimizeResponse_ArrayData(t *testing.T) {
	opt := NewResponseOptimizer(OptimizeOptions{
		Whitelist: []string{"id", "name"},
	})

	data := []byte(`[{"id":1,"name":"A","secret":"X"},{"id":2,"name":"B","secret":"Y"}]`)

	result, err := opt.OptimizeResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []map[string]any
	if err := json.Unmarshal(result, &items); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}

	for _, item := range items {
		if _, exists := item["secret"]; exists {
			t.Error("secret should be filtered out")
		}
	}
}