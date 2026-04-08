package service

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollector struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	MatchFailures   *prometheus.CounterVec
	registry        *prometheus.Registry
	mu              sync.Mutex
}

var (
	metricsInstance *MetricsCollector
	metricsOnce     sync.Once
)

func NewMetricsCollector() *MetricsCollector {
	metricsOnce.Do(func() {
		metricsInstance = &MetricsCollector{
			registry: prometheus.NewRegistry(),
		}

		metricsInstance.RequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mcp_router_requests_total",
				Help: "Total MCP router requests",
			},
			[]string{"agent", "task_type", "status"},
		)

		metricsInstance.RequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "mcp_router_request_duration_seconds",
				Help:    "MCP router request latency",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
			},
			[]string{"agent", "task_type"},
		)

		metricsInstance.MatchFailures = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mcp_router_match_failures_total",
				Help: "Route match failures",
			},
			[]string{"reason"},
		)

		metricsInstance.registry.MustRegister(
			metricsInstance.RequestsTotal,
			metricsInstance.RequestDuration,
			metricsInstance.MatchFailures,
		)
	})

	return metricsInstance
}

func (m *MetricsCollector) RecordRequest(agent, taskType, status string, duration float64) {
	m.RequestsTotal.WithLabelValues(agent, taskType, status).Inc()
	m.RequestDuration.WithLabelValues(agent, taskType).Observe(duration)
}

func (m *MetricsCollector) RecordMatchFailure(reason string) {
	m.MatchFailures.WithLabelValues(reason).Inc()
}
