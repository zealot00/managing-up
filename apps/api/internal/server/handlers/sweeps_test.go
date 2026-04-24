package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/service"
)

func TestSweepHandler_ListSweeps(t *testing.T) {
	t.Parallel()

	handler := NewSweepHandler(&mockSweepRepo{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sweeps", nil)
	rec := httptest.NewRecorder()

	handler.ListSweeps(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json response: %v", err)
	}

	_, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data payload")
	}
}

func TestSweepHandler_ListSweeps_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := NewSweepHandler(&mockSweepRepo{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sweeps", nil)
	rec := httptest.NewRecorder()

	handler.ListSweeps(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestSweepHandler_GetSweep(t *testing.T) {
	t.Parallel()

	handler := NewSweepHandler(&mockSweepRepo{configs: []SweepConfig{
		{
			ID:          "sweep_001",
			Name:        "test_sweep",
			Description: "Test sweep",
			TaskID:      "task_001",
			Parameters: SweepParameters{
				Models:       []string{"gpt-4o"},
				Temperatures: []float64{0.0},
				MaxTokens:    []int{512},
				Prompts:     []SweepPromptVariant{{ID: "p1", Label: "Default", Content: "Hello"}},
			},
			Status:     "pending",
			TotalRuns:  1,
			Completed:  0,
			CreatedBy:  "test_user",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sweeps/sweep_001", nil)
	rec := httptest.NewRecorder()

	handler.GetSweep(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json response: %v", err)
	}

	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data payload")
	}

	if data["id"] != "sweep_001" {
		t.Fatalf("expected sweep id sweep_001, got %v", data["id"])
	}
}

func TestSweepHandler_GetSweep_NotFound(t *testing.T) {
	t.Parallel()

	handler := NewSweepHandler(&mockSweepRepo{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sweeps/nonexistent", nil)
	rec := httptest.NewRecorder()

	handler.GetSweep(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestSweepHandler_CreateSweep(t *testing.T) {
	t.Parallel()

	handler := NewSweepHandler(&mockSweepRepo{
		tasks: map[string]service.Task{
			"task_001": {ID: "task_001", Name: "Test Task"},
		},
	})

	body := []byte(`{
		"name": "test_sweep",
		"description": "A test sweep",
		"task_id": "task_001",
		"parameters": {
			"models": ["gpt-4o", "claude-3-5-sonnet"],
			"temperatures": [0.0, 0.7],
			"max_tokens": [256, 1024],
			"prompts": [
				{"id": "p1", "label": "Default", "content": "Hello"},
				{"id": "p2", "label": "Formal", "content": "Hello, how can I help?"}
			]
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sweeps/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateSweep(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var bodyResp Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &bodyResp); err != nil {
		t.Fatalf("expected valid json response: %v", err)
	}

	data, ok := bodyResp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data payload")
	}

	if data["name"] != "test_sweep" {
		t.Fatalf("expected name test_sweep, got %v", data["name"])
	}

	totalRuns, ok := data["total_runs"].(float64)
	if !ok || totalRuns != 16 {
		t.Fatalf("expected total_runs 16, got %v", data["total_runs"])
	}
}

func TestSweepHandler_CreateSweep_ValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		body   string
		errMsg string
		status int
	}{
		{
			name:   "missing name",
			body:   `{"task_id": "task_001", "parameters": {"models": ["gpt-4o"], "temperatures": [0.0], "max_tokens": [256], "prompts": [{"id": "p1", "label": "D", "content": "H"}]}}`,
			errMsg: "NAME_REQUIRED",
			status: http.StatusBadRequest,
		},
		{
			name:   "missing task_id",
			body:   `{"name": "test_sweep", "parameters": {"models": ["gpt-4o"], "temperatures": [0.0], "max_tokens": [256], "prompts": [{"id": "p1", "label": "D", "content": "H"}]}}`,
			errMsg: "TASK_ID_REQUIRED",
			status: http.StatusBadRequest,
		},
		{
			name:   "task not found",
			body:   `{"name": "test_sweep", "task_id": "nonexistent", "parameters": {"models": ["gpt-4o"], "temperatures": [0.0], "max_tokens": [256], "prompts": [{"id": "p1", "label": "D", "content": "H"}]}}`,
			errMsg: "TASK_NOT_FOUND",
			status: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewSweepHandler(&mockSweepRepo{})

			req := httptest.NewRequest(http.MethodPost, "/api/v1/sweeps/create", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.CreateSweep(rec, req)

			if rec.Code != tt.status {
				t.Fatalf("expected status %d, got %d", tt.status, rec.Code)
			}

			var bodyResp Envelope
			if err := json.Unmarshal(rec.Body.Bytes(), &bodyResp); err != nil {
				t.Fatalf("expected valid json response: %v", err)
			}

			if bodyResp.Error == nil {
				t.Fatalf("expected error in response")
			}

			if bodyResp.Error.Code != tt.errMsg {
				t.Fatalf("expected error code %s, got %v", tt.errMsg, bodyResp.Error.Code)
			}
		})
	}
}

func TestSweepHandler_CreateSweep_InvalidContentType(t *testing.T) {
	t.Parallel()

	handler := NewSweepHandler(&mockSweepRepo{})

	body := []byte(`{"name": "test_sweep", "task_id": "task_001", "parameters": {"models": ["gpt-4o"], "temperatures": [0.0], "max_tokens": [256], "prompts": [{"id": "p1", "label": "D", "content": "H"}]}}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sweeps/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	handler.CreateSweep(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, rec.Code)
	}
}

func TestSweepHandler_CreateSweep_InvalidJSON(t *testing.T) {
	t.Parallel()

	handler := NewSweepHandler(&mockSweepRepo{})

	body := []byte(`{"name":}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sweeps/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateSweep(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSweepHandler_DeleteSweep(t *testing.T) {
	t.Parallel()

	repo := &mockSweepRepo{}
	handler := NewSweepHandler(repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/sweeps/delete/sweep_001", nil)
	rec := httptest.NewRecorder()

	handler.DeleteSweep(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !repo.deleted {
		t.Fatalf("expected DeleteSweepConfig to be called")
	}
}

func TestSweepHandler_DeleteSweep_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := NewSweepHandler(&mockSweepRepo{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sweeps/delete/sweep_001", nil)
	rec := httptest.NewRecorder()

	handler.DeleteSweep(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestSweepHandler_GetSweepMatrix(t *testing.T) {
	t.Parallel()

	handler := NewSweepHandler(&mockSweepRepo{
		configs: []SweepConfig{
			{
				ID:          "sweep_001",
				Name:        "test_sweep",
				Description: "Test sweep",
				TaskID:      "task_001",
				Parameters: SweepParameters{
					Models:       []string{"gpt-4o"},
					Temperatures: []float64{0.0},
					MaxTokens:    []int{512},
					Prompts:     []SweepPromptVariant{{ID: "p1", Label: "Default", Content: "Hello"}},
				},
				Status:     "pending",
				TotalRuns:  1,
				Completed:  0,
				CreatedBy:  "test_user",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
		runs: []SweepRun{
			{
				ID:            "run_001",
				SweepConfigID: "sweep_001",
				VariantIndex:  0,
				Model:         "gpt-4o",
				Temperature:   0.0,
				MaxTokens:     512,
				PromptID:      "p1",
				PromptLabel:   "Default",
				Status:        "completed",
				Score:         0.85,
				CreatedAt:     time.Now(),
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sweeps/sweep_001", nil)
	rec := httptest.NewRecorder()

	handler.GetSweepMatrix(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json response: %v", err)
	}

	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data payload")
	}

	if data["sweep_id"] != "sweep_001" {
		t.Fatalf("expected sweep_id sweep_001, got %v", data["sweep_id"])
	}

	summary, ok := data["summary"].(map[string]any)
	if !ok {
		t.Fatalf("expected summary in response")
	}

	if summary["total"].(float64) != 1 {
		t.Fatalf("expected total 1, got %v", summary["total"])
	}

	if summary["completed"].(float64) != 1 {
		t.Fatalf("expected completed 1, got %v", summary["completed"])
	}

	if summary["avg_score"].(float64) != 0.85 {
		t.Fatalf("expected avg_score 0.85, got %v", summary["avg_score"])
	}
}

func TestSweepHandler_GetSweepMatrix_NotFound(t *testing.T) {
	t.Parallel()

	handler := NewSweepHandler(&mockSweepRepo{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sweeps/matrix/nonexistent", nil)
	rec := httptest.NewRecorder()

	handler.GetSweepMatrix(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestCalculateTotalRuns(t *testing.T) {
	tests := []struct {
		name     string
		params   SweepParameters
		expected int
	}{
		{
			name: "basic calculation",
			params: SweepParameters{
				Models:       []string{"gpt-4o", "claude-3-5-sonnet"},
				Temperatures: []float64{0.0, 0.7, 1.0},
				MaxTokens:    []int{256, 1024},
				Prompts:     []SweepPromptVariant{{ID: "p1"}, {ID: "p2"}},
			},
			expected: 2 * 3 * 2 * 2,
		},
		{
			name: "single values",
			params: SweepParameters{
				Models:       []string{"gpt-4o"},
				Temperatures: []float64{0.5},
				MaxTokens:    []int{512},
				Prompts:     []SweepPromptVariant{{ID: "p1"}},
			},
			expected: 1,
		},
		{
			name: "empty params returns 1",
			params: SweepParameters{
				Models:       []string{},
				Temperatures: []float64{},
				MaxTokens:    []int{},
				Prompts:     []SweepPromptVariant{},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateTotalRuns(tt.params)
			if result != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGenerateSweepRuns(t *testing.T) {
	params := SweepParameters{
		Models:       []string{"gpt-4o", "claude-3-5-sonnet"},
		Temperatures: []float64{0.0, 0.7},
		MaxTokens:    []int{256, 1024},
		Prompts:     []SweepPromptVariant{
			{ID: "p1", Label: "Default", Content: "Hello"},
			{ID: "p2", Label: "Formal", Content: "Hello!"},
		},
	}

	runs := generateSweepRuns("sweep_001", params)

	expected := 2 * 2 * 2 * 2
	if len(runs) != expected {
		t.Fatalf("expected %d runs, got %d", expected, len(runs))
	}

	for i, run := range runs {
		if run.SweepConfigID != "sweep_001" {
			t.Fatalf("run %d: expected sweep_config_id sweep_001, got %s", i, run.SweepConfigID)
		}
		if run.Status != "pending" {
			t.Fatalf("run %d: expected status pending, got %s", i, run.Status)
		}
		if run.VariantIndex != i {
			t.Fatalf("run %d: expected variant_index %d, got %d", i, i, run.VariantIndex)
		}
	}
}

func getFloat(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}

func getInt(v any) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	default:
		return 0
	}
}

func approxEqual(a, b float64) bool {
	const epsilon = 0.001
	return a > b-epsilon && a < b+epsilon
}

func TestBuildMatrixSummary(t *testing.T) {
	now := time.Now()
	completedAt := now

	tests := []struct {
		name     string
		runs     []SweepRun
		expected map[string]any
	}{
		{
			name: "all completed with scores",
			runs: []SweepRun{
				{ID: "r1", Status: "completed", Score: 0.9, CreatedAt: now, CompletedAt: &completedAt},
				{ID: "r2", Status: "completed", Score: 0.7, CreatedAt: now, CompletedAt: &completedAt},
				{ID: "r3", Status: "completed", Score: 0.8, CreatedAt: now, CompletedAt: &completedAt},
			},
			expected: map[string]any{
				"total":      3,
				"completed":  3,
				"pending":    0,
				"avg_score":  0.8,
				"max_score":  0.9,
				"min_score":  0.7,
			},
		},
		{
			name: "mixed status",
			runs: []SweepRun{
				{ID: "r1", Status: "completed", Score: 0.9, CreatedAt: now, CompletedAt: &completedAt},
				{ID: "r2", Status: "pending", Score: 0, CreatedAt: now},
				{ID: "r3", Status: "failed", Score: 0, CreatedAt: now},
			},
			expected: map[string]any{
				"total":      3,
				"completed":  1,
				"pending":    2,
				"avg_score":  0.9,
				"max_score":  0.9,
				"min_score":  0.9,
			},
		},
		{
			name:     "empty runs",
			runs:     []SweepRun{},
			expected: map[string]any{
				"total":      0,
				"completed":  0,
				"pending":    0,
				"avg_score":  0.0,
				"max_score":  -1,
				"min_score":  2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := buildMatrixSummary(tt.runs)

			if getInt(summary["total"]) != tt.expected["total"] {
				t.Fatalf("expected total %v, got %v", tt.expected["total"], summary["total"])
			}
			if getInt(summary["completed"]) != tt.expected["completed"] {
				t.Fatalf("expected completed %v, got %v", tt.expected["completed"], summary["completed"])
			}
			if getInt(summary["pending"]) != tt.expected["pending"] {
				t.Fatalf("expected pending %v, got %v", tt.expected["pending"], summary["pending"])
			}
			if !approxEqual(getFloat(summary["avg_score"]), getFloat(tt.expected["avg_score"])) {
				t.Fatalf("expected avg_score %v, got %v", tt.expected["avg_score"], summary["avg_score"])
			}
			if !approxEqual(getFloat(summary["max_score"]), getFloat(tt.expected["max_score"])) {
				t.Fatalf("expected max_score %v, got %v", tt.expected["max_score"], summary["max_score"])
			}
			if !approxEqual(getFloat(summary["min_score"]), getFloat(tt.expected["min_score"])) {
				t.Fatalf("expected min_score %v, got %v", tt.expected["min_score"], summary["min_score"])
			}
		})
	}
}

type mockSweepRepo struct {
	configs  []SweepConfig
	runs     []SweepRun
	tasks    map[string]service.Task
	deleted  bool
}

func (m *mockSweepRepo) CreateSweepConfig(cfg SweepConfig) (SweepConfig, error) {
	m.configs = append(m.configs, cfg)
	return cfg, nil
}

func (m *mockSweepRepo) GetSweepConfig(id string) (SweepConfig, bool) {
	for _, c := range m.configs {
		if c.ID == id {
			return c, true
		}
	}
	return SweepConfig{}, false
}

func (m *mockSweepRepo) ListSweepConfigs() ([]SweepConfig, error) {
	return m.configs, nil
}

func (m *mockSweepRepo) UpdateSweepConfig(cfg SweepConfig) error {
	for i, c := range m.configs {
		if c.ID == cfg.ID {
			m.configs[i] = cfg
			break
		}
	}
	return nil
}

func (m *mockSweepRepo) DeleteSweepConfig(id string) error {
	m.deleted = true
	return nil
}

func (m *mockSweepRepo) CreateSweepRuns(runs []SweepRun) error {
	m.runs = append(m.runs, runs...)
	return nil
}

func (m *mockSweepRepo) GetSweepRunsByConfigID(configID string) ([]SweepRun, error) {
	var result []SweepRun
	for _, r := range m.runs {
		if r.SweepConfigID == configID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockSweepRepo) UpdateSweepRun(run SweepRun) error {
	for i, r := range m.runs {
		if r.ID == run.ID {
			m.runs[i] = run
			break
		}
	}
	return nil
}

func (m *mockSweepRepo) GetTask(id string) (service.Task, bool) {
	if m.tasks == nil {
		return service.Task{}, false
	}
	t, ok := m.tasks[id]
	return t, ok
}