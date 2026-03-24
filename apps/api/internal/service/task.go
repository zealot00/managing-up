package service

import (
	"fmt"
	"time"
)

type Task struct {
	ID          string
	Name        string
	Description string
	SkillID     string
	Tags        []string
	Difficulty  string
	TestCases   []TestCase
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TestCase struct {
	Input    map[string]any
	Expected any
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
}

func (s *TaskService) CreateTask(req CreateTaskRequest) (Task, error) {
	if req.Name == "" {
		return Task{}, ErrTaskNameRequired
	}

	if req.Difficulty == "" {
		req.Difficulty = "medium"
	}

	if !isValidDifficulty(req.Difficulty) {
		return Task{}, ErrInvalidDifficulty
	}

	if req.Tags == nil {
		req.Tags = []string{}
	}

	if req.TestCases == nil {
		req.TestCases = []TestCase{}
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
