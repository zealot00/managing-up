package service

import (
	"fmt"
	"time"
)

type Task struct {
	ID          string
	Name        string
	Description string
	Tags        []string
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// === New fields (Phase 0.1) ===
	TaskType  string
	Input     TaskInput
	Gold      GoldConfig
	Scoring   ScoringConfig
	Execution ExecutionConfig

	// === Legacy fields (kept for backward compatibility) ===
	SkillID    string
	Difficulty string
	TestCases  []TestCase
}

type TestCase struct {
	Input    map[string]any
	Expected any
}

type TaskInput struct {
	Source string
	Path   string
	Format string
}

type GoldConfig struct {
	Type string
	Data any
}

type Threshold struct {
	Pass            float64
	RegressionAlert float64
}

type ScoringConfig struct {
	PrimaryMetric    string
	SecondaryMetrics []string
	Threshold        Threshold
}

type ExecutionConfig struct {
	Model       string
	Temperature float64
	MaxTokens   int
	Seed        int64
}

var validDifficulties = []string{"easy", "medium", "hard"}

type TaskRepository interface {
	CreateTask(task Task) (Task, error)
	GetTask(id string) (Task, bool)
	ListTasks(skillID string, difficulty string) []Task
	UpdateTask(task Task) error
	DeleteTask(id string) error
}

type TaskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

type CreateTaskRequest struct {
	Name        string
	Description string
	SkillID     string
	Tags        []string
	Difficulty  string
	TestCases   []TestCase

	// === New fields (Phase 0.1) ===
	TaskType  string
	Input     TaskInput
	Gold      GoldConfig
	Scoring   ScoringConfig
	Execution ExecutionConfig
}

func (s *TaskService) CreateTask(req CreateTaskRequest) (Task, error) {
	if req.Name == "" {
		return Task{}, ErrTaskNameRequired
	}

	if req.Tags == nil {
		req.Tags = []string{}
	}

	if req.TestCases == nil {
		req.TestCases = []TestCase{}
	}

	if req.TaskType == "" {
		req.TaskType = "benchmark"
	}

	if req.Input.Source == "" {
		req.Input.Source = "inline"
	}

	if req.Gold.Type == "" {
		req.Gold.Type = "exact_match"
	}

	if req.Scoring.PrimaryMetric == "" {
		req.Scoring.PrimaryMetric = "exact_match"
	}
	if req.Scoring.Threshold.Pass == 0 {
		req.Scoring.Threshold.Pass = 0.85
	}
	if req.Scoring.Threshold.RegressionAlert == 0 {
		req.Scoring.Threshold.RegressionAlert = 0.90
	}

	if req.Execution.Model == "" {
		req.Execution.Model = "gpt-4o"
	}
	if req.Execution.Temperature == 0 {
		req.Execution.Temperature = 0.0
	}
	if req.Execution.MaxTokens == 0 {
		req.Execution.MaxTokens = 2048
	}

	if req.Difficulty == "" {
		req.Difficulty = "medium"
	}

	if !isValidDifficulty(req.Difficulty) {
		return Task{}, ErrInvalidDifficulty
	}

	now := time.Now()
	task := Task{
		ID:          fmt.Sprintf("task_%d", time.Now().UnixNano()),
		Name:        req.Name,
		Description: req.Description,
		SkillID:     req.SkillID,
		Tags:        req.Tags,
		Difficulty:  req.Difficulty,
		TestCases:   req.TestCases,
		CreatedAt:   now,
		UpdatedAt:   now,

		TaskType:  req.TaskType,
		Input:     req.Input,
		Gold:      req.Gold,
		Scoring:   req.Scoring,
		Execution: req.Execution,
	}

	return s.repo.CreateTask(task)
}

func (s *TaskService) GetTask(id string) (Task, error) {
	task, ok := s.repo.GetTask(id)
	if !ok {
		return Task{}, ErrTaskNotFound
	}
	return task, nil
}

func (s *TaskService) ListTasks(skillID string, difficulty string) []Task {
	return s.repo.ListTasks(skillID, difficulty)
}

func (s *TaskService) UpdateTask(id string, req CreateTaskRequest) (Task, error) {
	task, ok := s.repo.GetTask(id)
	if !ok {
		return Task{}, ErrTaskNotFound
	}

	if req.Name != "" {
		task.Name = req.Name
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.SkillID != "" {
		task.SkillID = req.SkillID
	}
	if len(req.Tags) > 0 {
		task.Tags = req.Tags
	}
	if req.Difficulty != "" {
		if !isValidDifficulty(req.Difficulty) {
			return Task{}, ErrInvalidDifficulty
		}
		task.Difficulty = req.Difficulty
	}
	if len(req.TestCases) > 0 {
		task.TestCases = req.TestCases
	}
	if req.TaskType != "" {
		task.TaskType = req.TaskType
	}
	if req.Input.Source != "" {
		task.Input.Source = req.Input.Source
	}
	if req.Input.Path != "" {
		task.Input.Path = req.Input.Path
	}
	if req.Input.Format != "" {
		task.Input.Format = req.Input.Format
	}
	if req.Gold.Type != "" {
		task.Gold.Type = req.Gold.Type
	}
	if req.Gold.Data != nil {
		task.Gold.Data = req.Gold.Data
	}
	if req.Scoring.PrimaryMetric != "" {
		task.Scoring.PrimaryMetric = req.Scoring.PrimaryMetric
	}
	if len(req.Scoring.SecondaryMetrics) > 0 {
		task.Scoring.SecondaryMetrics = req.Scoring.SecondaryMetrics
	}
	if req.Scoring.Threshold.Pass != 0 {
		task.Scoring.Threshold.Pass = req.Scoring.Threshold.Pass
	}
	if req.Scoring.Threshold.RegressionAlert != 0 {
		task.Scoring.Threshold.RegressionAlert = req.Scoring.Threshold.RegressionAlert
	}
	if req.Execution.Model != "" {
		task.Execution.Model = req.Execution.Model
	}
	if req.Execution.Temperature != 0 {
		task.Execution.Temperature = req.Execution.Temperature
	}
	if req.Execution.MaxTokens != 0 {
		task.Execution.MaxTokens = req.Execution.MaxTokens
	}
	if req.Execution.Seed != 0 {
		task.Execution.Seed = req.Execution.Seed
	}
	task.UpdatedAt = time.Now()

	return task, s.repo.UpdateTask(task)
}

func (s *TaskService) DeleteTask(id string) error {
	_, ok := s.repo.GetTask(id)
	if !ok {
		return ErrTaskNotFound
	}
	return s.repo.DeleteTask(id)
}

func isValidDifficulty(d string) bool {
	for _, valid := range validDifficulties {
		if d == valid {
			return true
		}
	}
	return false
}
