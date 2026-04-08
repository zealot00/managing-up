package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/service"
)

type MCPRouterHandler struct {
	routerSvc *service.MCPRouterService
	metrics   *service.MetricsCollector
}

func NewMCPRouterHandler(routerSvc *service.MCPRouterService, metrics *service.MetricsCollector) *MCPRouterHandler {
	return &MCPRouterHandler{routerSvc: routerSvc, metrics: metrics}
}

type RouteRequest struct {
	AgentID       string    `json:"agent_id"`
	CorrelationID string    `json:"correlation_id"`
	Task          RouteTask `json:"task"`
}

type RouteTask struct {
	Structured RouteTaskStructured `json:"structured"`
}

type RouteTaskStructured struct {
	TaskType string   `json:"task_type"`
	Tags     []string `json:"tags"`
}

type RouteResponse struct {
	Matched       bool         `json:"matched"`
	Target        *RouteTarget `json:"target,omitempty"`
	MatchScore    float64      `json:"match_score,omitempty"`
	RoutingTimeMS int          `json:"routing_time_ms,omitempty"`
}

type RouteTarget struct {
	ServerID   string `json:"server_id"`
	ServerName string `json:"server_name"`
	Transport  string `json:"transport"`
	Endpoint   string `json:"endpoint"`
}

func (h *MCPRouterHandler) Route(w http.ResponseWriter, r *http.Request) {
	var req RouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	start := time.Now()

	taskTypes := []string{}
	tags := []string{}

	if req.Task.Structured.TaskType != "" {
		taskTypes = []string{req.Task.Structured.TaskType}
	}
	if len(req.Task.Structured.Tags) > 0 {
		tags = req.Task.Structured.Tags
	}

	result, err := h.routerSvc.MatchTask(ctx, taskTypes, tags)
	duration := time.Since(start).Seconds()

	if err != nil {
		h.metrics.RecordRequest(req.AgentID, req.Task.Structured.TaskType, "failure", duration)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	if !result.Matched {
		h.metrics.RecordMatchFailure("no_matching_server")
		h.metrics.RecordRequest(req.AgentID, req.Task.Structured.TaskType, "no_match", duration)
		writeEnvelopeMatch(w, http.StatusOK, false, nil, 0, 0, req.CorrelationID)
		return
	}

	h.metrics.RecordRequest(req.AgentID, req.Task.Structured.TaskType, "success", duration)

	writeEnvelopeMatch(w, http.StatusOK, true, result, time.Since(start).Milliseconds(), result.MatchScore, req.CorrelationID)
}

func (h *MCPRouterHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	entries, err := h.routerSvc.GetCatalog(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeEnvelope(w, http.StatusOK, "req_mcp_catalog", entries)
}

func (h *MCPRouterHandler) Match(w http.ResponseWriter, r *http.Request) {
	taskType := r.URL.Query().Get("task_type")
	tagsParam := r.URL.Query().Get("tags")

	tags := []string{}
	if tagsParam != "" {
		tags = splitTags(tagsParam)
	}

	ctx := r.Context()
	result, err := h.routerSvc.MatchTask(ctx, []string{taskType}, tags)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeEnvelope(w, http.StatusOK, "req_mcp_match", result)
}

func writeEnvelopeMatch(w http.ResponseWriter, status int, matched bool, result *service.MatchResult, routingTimeMs int64, matchScore float64, correlationID string) {
	resp := RouteResponse{
		Matched: matched,
	}
	if matched && result != nil {
		resp.Target = &RouteTarget{
			ServerID:   result.ServerID,
			ServerName: result.ServerName,
			Transport:  result.Transport,
			Endpoint:   result.Endpoint,
		}
		resp.MatchScore = matchScore
		resp.RoutingTimeMS = int(routingTimeMs)
	}
	writeEnvelope(w, status, correlationID, resp)
}

func splitTags(s string) []string {
	if s == "" {
		return nil
	}
	var tags []string
	for _, t := range splitString(s, ",") {
		tags = append(tags, trimSpace(t))
	}
	return tags
}

func splitString(s, sep string) []string {
	var result []string
	for i := 0; i < len(s); {
		idx := indexOf(s, sep, i)
		if idx == -1 {
			result = append(result, s[i:])
			break
		}
		result = append(result, s[i:idx])
		i = idx + len(sep)
	}
	return result
}

func indexOf(s, substr string, start int) int {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
