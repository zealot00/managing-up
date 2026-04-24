package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/zealot/managing-up/apps/api/internal/service"
)

type SweepHandler struct {
	repo SweepRepository
}

type SweepRepository interface {
	CreateSweepConfig(cfg SweepConfig) (SweepConfig, error)
	GetSweepConfig(id string) (SweepConfig, bool)
	ListSweepConfigs() ([]SweepConfig, error)
	UpdateSweepConfig(cfg SweepConfig) error
	DeleteSweepConfig(id string) error
	CreateSweepRuns(runs []SweepRun) error
	GetSweepRunsByConfigID(configID string) ([]SweepRun, error)
	UpdateSweepRun(run SweepRun) error
	GetTask(id string) (service.Task, bool)
}

type SweepConfig struct {
	ID          string          `json:"id"`
	Name       string          `json:"name"`
	Description string         `json:"description"`
	TaskID      string         `json:"task_id"`
	Parameters  SweepParameters `json:"parameters"`
	Status     string         `json:"status"`
	TotalRuns  int            `json:"total_runs"`
	Completed  int            `json:"completed"`
	CreatedBy  string         `json:"created_by"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type SweepParameters struct {
	Models       []string            `json:"models"`
	Temperatures []float64           `json:"temperatures"`
	MaxTokens    []int               `json:"max_tokens"`
	Prompts      []SweepPromptVariant `json:"prompts"`
}

type SweepPromptVariant struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Content string `json:"content"`
}

type SweepRun struct {
	ID              string     `json:"id"`
	SweepConfigID   string     `json:"sweep_config_id"`
	VariantIndex    int        `json:"variant_index"`
	Model           string     `json:"model"`
	Temperature     float64    `json:"temperature"`
	MaxTokens       int        `json:"max_tokens"`
	PromptID        string     `json:"prompt_id"`
	PromptLabel     string     `json:"prompt_label"`
	Status          string     `json:"status"`
	TaskExecutionID string     `json:"task_execution_id,omitempty"`
	Score           float64    `json:"score,omitempty"`
	DurationMs      int64      `json:"duration_ms,omitempty"`
	Error           string     `json:"error,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type CreateSweepRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	TaskID     string          `json:"task_id"`
	Parameters  SweepParameters `json:"parameters"`
}

type SweepMatrixCell struct {
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	PromptID    string  `json:"prompt_id"`
	PromptLabel string  `json:"prompt_label"`
	RunID       string  `json:"run_id,omitempty"`
	Status      string  `json:"status"`
	Score       float64 `json:"score,omitempty"`
}

func NewSweepHandler(repo SweepRepository) *SweepHandler {
	return &SweepHandler{repo: repo}
}

func (h *SweepHandler) ListSweeps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	configs, err := h.repo.ListSweepConfigs()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeEnvelope(w, http.StatusOK, "sweeps", map[string]any{
		"items": configs,
	})
}

func (h *SweepHandler) GetSweep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	id := extractSweepID(r.URL.Path)
	if id == "" {
		writeError(w, http.StatusBadRequest, "ID_REQUIRED", "sweep id is required")
		return
	}

	config, ok := h.repo.GetSweepConfig(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "sweep not found")
		return
	}

	writeEnvelope(w, http.StatusOK, "sweep", config)
}

func (h *SweepHandler) CreateSweep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	if !isJSONRequest(r) {
		writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json")
		return
	}

	var req CreateSweepRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "NAME_REQUIRED", "name is required")
		return
	}
	if req.TaskID == "" {
		writeError(w, http.StatusBadRequest, "TASK_ID_REQUIRED", "task_id is required")
		return
	}

	_, ok := h.repo.GetTask(req.TaskID)
	if !ok {
		writeError(w, http.StatusNotFound, "TASK_NOT_FOUND", "task not found")
		return
	}

	totalRuns := calculateTotalRuns(req.Parameters)
	now := time.Now()

	cfg := SweepConfig{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		TaskID:      req.TaskID,
		Parameters:  req.Parameters,
		Status:      "pending",
		TotalRuns:   totalRuns,
		Completed:   0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	runs := generateSweepRuns(cfg.ID, req.Parameters)

	if _, err := h.repo.CreateSweepConfig(cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}

	if err := h.repo.CreateSweepRuns(runs); err != nil {
		writeError(w, http.StatusInternalServerError, "CREATE_RUNS_FAILED", err.Error())
		return
	}

	writeEnvelope(w, http.StatusCreated, "sweep_created", map[string]any{
		"id":         cfg.ID,
		"name":       cfg.Name,
		"total_runs": totalRuns,
	})
}

func (h *SweepHandler) DeleteSweep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	id := extractSweepID(r.URL.Path)
	if id == "" {
		writeError(w, http.StatusBadRequest, "ID_REQUIRED", "sweep id is required")
		return
	}

	if err := h.repo.DeleteSweepConfig(id); err != nil {
		writeError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	writeEnvelope(w, http.StatusOK, "sweep_deleted", map[string]string{"status": "deleted"})
}

func (h *SweepHandler) GetSweepMatrix(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	id := extractSweepID(r.URL.Path)
	if id == "" {
		writeError(w, http.StatusBadRequest, "ID_REQUIRED", "sweep id is required")
		return
	}

	_, ok := h.repo.GetSweepConfig(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "sweep not found")
		return
	}

	runs, err := h.repo.GetSweepRunsByConfigID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	matrix := buildMatrix(runs)

	writeEnvelope(w, http.StatusOK, "sweep_matrix", map[string]any{
		"sweep_id": id,
		"matrix":    matrix,
		"summary":  buildMatrixSummary(runs),
	})
}

func extractSweepID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, p := range parts {
		if p == "sweeps" && i+1 < len(parts) {
			next := parts[i+1]
			if next == "matrix" || next == "delete" {
				if i+2 < len(parts) {
					return parts[i+2]
				}
			} else {
				return next
			}
		}
	}
	return ""
}

func calculateTotalRuns(params SweepParameters) int {
	total := len(params.Models) * len(params.Temperatures) * len(params.MaxTokens) * len(params.Prompts)
	if total == 0 {
		total = 1
	}
	return total
}

func generateSweepRuns(configID string, params SweepParameters) []SweepRun {
	var runs []SweepRun
	variantIndex := 0
	now := time.Now()

	for _, model := range params.Models {
		for _, temp := range params.Temperatures {
			for _, tokens := range params.MaxTokens {
				for _, prompt := range params.Prompts {
					runs = append(runs, SweepRun{
						ID:            uuid.New().String(),
						SweepConfigID: configID,
						VariantIndex:  variantIndex,
						Model:         model,
						Temperature:   temp,
						MaxTokens:     tokens,
						PromptID:      prompt.ID,
						PromptLabel:   prompt.Label,
						Status:        "pending",
						CreatedAt:     now,
					})
					variantIndex++
				}
			}
		}
	}

	return runs
}

func buildMatrix(runs []SweepRun) [][]SweepMatrixCell {
	if len(runs) == 0 {
		return [][]SweepMatrixCell{}
	}

	type key struct {
		model string
		temp  float64
	}
	modelTempRuns := make(map[key][]SweepRun)
	for _, run := range runs {
		k := key{model: run.Model, temp: run.Temperature}
		modelTempRuns[k] = append(modelTempRuns[k], run)
	}

	var matrix [][]SweepMatrixCell
	for k, runsList := range modelTempRuns {
		row := make([]SweepMatrixCell, 0, len(runsList))
		for _, run := range runsList {
			cell := SweepMatrixCell{
				Model:       run.Model,
				Temperature: run.Temperature,
				MaxTokens:   run.MaxTokens,
				PromptID:    run.PromptID,
				PromptLabel: run.PromptLabel,
				RunID:       run.ID,
				Status:      run.Status,
				Score:       run.Score,
			}
			row = append(row, cell)
		}
		_ = k
		matrix = append(matrix, row)
	}

	return matrix
}

func buildMatrixSummary(runs []SweepRun) map[string]any {
	total := len(runs)
	completed := 0
	var totalScore float64
	var maxScore float64 = -1
	var minScore float64 = 2

	for _, run := range runs {
		if run.Status == "completed" {
			completed++
			totalScore += run.Score
			if run.Score > maxScore {
				maxScore = run.Score
			}
			if run.Score < minScore {
				minScore = run.Score
			}
		}
	}

	avgScore := 0.0
	if completed > 0 {
		avgScore = totalScore / float64(completed)
	}

	return map[string]any{
		"total":      total,
		"completed":  completed,
		"pending":    total - completed,
		"avg_score":  avgScore,
		"max_score":  maxScore,
		"min_score":  minScore,
	}
}