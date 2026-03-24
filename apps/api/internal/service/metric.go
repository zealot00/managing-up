package service

import (
	"fmt"
	"time"
)

type Metric struct {
	ID        string
	Name      string
	Type      string
	Config    map[string]any
	CreatedAt time.Time
}

type MetricRepository interface {
	CreateMetric(metric Metric) (Metric, error)
	GetMetric(id string) (Metric, bool)
	ListMetrics() []Metric
}

type MetricService struct {
	repo MetricRepository
}

func NewMetricService(repo MetricRepository) *MetricService {
	return &MetricService{repo: repo}
}

type CreateMetricRequest struct {
	Name   string
	Type   string
	Config map[string]any
}

var validMetricTypes = []string{"exact_match", "semantic_similarity", "judge_model"}

func (s *MetricService) CreateMetric(req CreateMetricRequest) (Metric, error) {
	if req.Name == "" {
		return Metric{}, ErrMetricNameRequired
	}

	if req.Type == "" {
		return Metric{}, ErrInvalidMetricType
	}

	if !isValidMetricType(req.Type) {
		return Metric{}, ErrInvalidMetricType
	}

	if req.Config == nil {
		req.Config = map[string]any{}
	}

	metric := Metric{
		ID:        fmt.Sprintf("metric_%d", time.Now().UnixNano()),
		Name:      req.Name,
		Type:      req.Type,
		Config:    req.Config,
		CreatedAt: time.Now(),
	}

	return s.repo.CreateMetric(metric)
}

func (s *MetricService) GetMetric(id string) (Metric, error) {
	metric, ok := s.repo.GetMetric(id)
	if !ok {
		return Metric{}, ErrMetricNotFound
	}
	return metric, nil
}

func (s *MetricService) ListMetrics() []Metric {
	return s.repo.ListMetrics()
}

func isValidMetricType(t string) bool {
	for _, valid := range validMetricTypes {
		if t == valid {
			return true
		}
	}
	return false
}
