package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/config"
	"github.com/zealot/managing-up/apps/api/internal/service"
)

var logger *slog.Logger

func SetLogger(l *slog.Logger) {
	logger = l
}

type repoToSkillRepoAdapter struct {
	repo Repository
}

func (a repoToSkillRepoAdapter) ListSkills(status string) []service.Skill {
	skills := a.repo.ListSkills(status)
	result := make([]service.Skill, len(skills))
	for i, s := range skills {
		result[i] = service.Skill{
			ID:             s.ID,
			Name:           s.Name,
			OwnerTeam:      s.OwnerTeam,
			RiskLevel:      s.RiskLevel,
			Status:         s.Status,
			CurrentVersion: s.CurrentVersion,
			CreatedBy:      s.CreatedBy,
		}
	}
	return result
}

func (a repoToSkillRepoAdapter) GetSkill(id string) (service.Skill, bool) {
	skill, ok := a.repo.GetSkill(id)
	if !ok {
		return service.Skill{}, false
	}
	return service.Skill{
		ID:             skill.ID,
		Name:           skill.Name,
		OwnerTeam:      skill.OwnerTeam,
		RiskLevel:      skill.RiskLevel,
		Status:         skill.Status,
		CurrentVersion: skill.CurrentVersion,
		CreatedBy:      skill.CreatedBy,
	}, true
}

func (a repoToSkillRepoAdapter) CreateSkill(req service.CreateSkillRequest) service.Skill {
	skill := a.repo.CreateSkill(CreateSkillRequest(req))
	return service.Skill{
		ID:             skill.ID,
		Name:           skill.Name,
		OwnerTeam:      skill.OwnerTeam,
		RiskLevel:      skill.RiskLevel,
		Status:         skill.Status,
		CurrentVersion: skill.CurrentVersion,
		CreatedBy:      skill.CreatedBy,
	}
}

type repoToExecutionRepoAdapter struct {
	repo Repository
}

func (a repoToExecutionRepoAdapter) GetSkill(id string) (service.Skill, bool) {
	skill, ok := a.repo.GetSkill(id)
	if !ok {
		return service.Skill{}, false
	}
	return service.Skill{
		ID:             skill.ID,
		Name:           skill.Name,
		OwnerTeam:      skill.OwnerTeam,
		RiskLevel:      skill.RiskLevel,
		Status:         skill.Status,
		CurrentVersion: skill.CurrentVersion,
		CreatedBy:      skill.CreatedBy,
	}, true
}

func (a repoToExecutionRepoAdapter) CreateExecution(req service.CreateExecutionRequest) (service.Execution, bool) {
	exec, ok := a.repo.CreateExecution(CreateExecutionRequest(req))
	if !ok {
		return service.Execution{}, false
	}
	return service.Execution{
		ID:            exec.ID,
		SkillID:       exec.SkillID,
		SkillName:     exec.SkillName,
		Status:        exec.Status,
		TriggeredBy:   exec.TriggeredBy,
		CurrentStepID: exec.CurrentStepID,
		Input:         exec.Input,
	}, true
}

func (a repoToExecutionRepoAdapter) ApproveExecution(executionID string, req service.ApproveExecutionRequest) (service.Approval, bool) {
	approval, ok := a.repo.ApproveExecution(executionID, ApproveExecutionRequest(req))
	if !ok {
		return service.Approval{}, false
	}
	return service.Approval{
		ID:             approval.ID,
		ExecutionID:    approval.ExecutionID,
		SkillName:      approval.SkillName,
		StepID:         approval.StepID,
		Status:         approval.Status,
		ApproverGroup:  approval.ApproverGroup,
		ApprovedBy:     approval.ApprovedBy,
		ResolutionNote: approval.ResolutionNote,
	}, true
}

// Server wraps the HTTP server and route registration for the API service.
type Server struct {
	httpServer *http.Server
	repo       Repository
	skillSvc   *service.SkillService
	execSvc    *service.ExecutionService
	closeFn    func() error
}

// New creates a configured API server.
func New(cfg config.Config) *Server {
	return NewWithRepository(cfg, newStore(), nil)
}

// NewWithRepository creates a configured API server with an injected repository.
func NewWithRepository(cfg config.Config, repo Repository, closeFn func() error) *Server {
	mux := http.NewServeMux()
	srv := &Server{
		repo:     repo,
		closeFn:  closeFn,
		skillSvc: service.NewSkillService(repoToSkillRepoAdapter{repo}),
		execSvc:  service.NewExecutionService(repoToExecutionRepoAdapter{repo}),
	}

	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/api/v1/meta", handleMeta)
	mux.HandleFunc("/api/v1/dashboard", srv.handleDashboard)
	mux.HandleFunc("/api/v1/procedure-drafts", srv.handleProcedureDrafts)
	mux.HandleFunc("/api/v1/skills", srv.handleSkills)
	mux.HandleFunc("/api/v1/skills/", srv.handleSkillByID)
	mux.HandleFunc("/api/v1/skills/{id}/spec", srv.handleSkillSpec)
	mux.HandleFunc("/api/v1/skill-versions", srv.handleSkillVersions)
	mux.HandleFunc("/api/v1/approvals", srv.handleApprovals)
	mux.HandleFunc("/api/v1/executions", srv.handleExecutions)
	mux.HandleFunc("/api/v1/executions/", srv.handleExecutionByID)
	mux.HandleFunc("/api/v1/agents", srv.handleAgents)
	mux.HandleFunc("/api/v1/generate-skill", srv.handleGenerateSkill)

	mux.HandleFunc("/api/v1/tasks", srv.handleTasks)
	mux.HandleFunc("/api/v1/tasks/", srv.handleTaskByID)
	mux.HandleFunc("/api/v1/metrics", srv.handleMetrics)
	mux.HandleFunc("/api/v1/task-executions", srv.handleTaskExecutions)
	mux.HandleFunc("/api/v1/task-executions/", srv.handleTaskExecutionByID)
	mux.HandleFunc("/api/v1/experiments", srv.handleExperiments)
	mux.HandleFunc("/api/v1/experiments/", srv.handleExperimentByID)
	mux.HandleFunc("/api/v1/experiments/{id}/compare", srv.handleExperimentCompare)
	mux.HandleFunc("/api/v1/check-regression", srv.handleCheckRegression)
	mux.HandleFunc("/api/v1/replay-snapshots", srv.handleReplaySnapshots)
	mux.HandleFunc("/api/v1/replay-snapshots/", srv.handleReplaySnapshotByID)

	srv.httpServer = &http.Server{
		Addr:              cfg.Address(),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return srv
}

// Start runs the API server until it exits.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the API server.
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	if s.closeFn != nil {
		if err := s.closeFn(); err != nil {
			return err
		}
	}

	return nil
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func handleMeta(w http.ResponseWriter, _ *http.Request) {
	writeEnvelope(w, http.StatusOK, "req_meta", map[string]any{
		"service": "managing-up-api",
		"runtime": "go",
		"scope": []string{
			"registry",
			"execution",
			"approval",
		},
	})
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	writeEnvelope(w, http.StatusOK, "req_dashboard", s.repo.Dashboard())
}

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		limit, offset := parsePagination(r.URL.Query())
		items := s.repo.ListSkills(r.URL.Query().Get("status"))
		total := len(items)
		if offset > total {
			offset = total
		}
		end := offset + limit
		if end > total {
			end = total
		}
		paginatedItems := items[offset:end]
		pagination := &Pagination{Limit: limit, Offset: offset, Total: total}
		writeEnvelopeWithPagination(w, http.StatusOK, generateRequestID(), map[string]any{
			"items": paginatedItems,
		}, pagination)
	case http.MethodPost:
		if !isJSONRequest(r) {
			writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
			return
		}

		var req CreateSkillRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}

		skill, err := s.skillSvc.CreateSkill(service.CreateSkillRequest(req))
		if err != nil {
			switch {
			case errors.Is(err, service.ErrSkillNameRequired):
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "name is required.")
			case errors.Is(err, service.ErrOwnerTeamRequired):
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "owner_team is required.")
			case errors.Is(err, service.ErrInvalidRiskLevel):
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "risk_level must be one of low, medium, high.")
			case errors.Is(err, service.ErrDuplicateSkillName):
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "skill with this name already exists.")
			default:
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create skill.")
			}
			return
		}

		if logger != nil {
			logger.Info("skill created",
				slog.String("skill_id", skill.ID),
				slog.String("owner_team", skill.OwnerTeam),
			)
		}
		writeEnvelope(w, http.StatusCreated, generateRequestID(), skill)
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) handleSkillVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	writeEnvelope(w, http.StatusOK, "req_skill_versions", map[string]any{
		"items": s.repo.ListSkillVersions(r.URL.Query().Get("skill_id")),
	})
}

func (s *Server) handleSkillByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/skills/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Skill not found.")
		return
	}

	skill, ok := s.repo.GetSkill(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Skill not found.")
		return
	}

	writeEnvelope(w, http.StatusOK, "req_skill_detail", skill)
}

func (s *Server) handleSkillSpec(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/skills/")
	id := strings.TrimSuffix(path, "/spec")
	if id == "" || id == path {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Skill spec not found.")
		return
	}

	skillVersion, ok := s.repo.GetSkillVersionForExecution(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Skill spec not found.")
		return
	}

	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/yaml") || strings.Contains(accept, "application/x-yaml") {
		w.Header().Set("Content-Type", "application/x-yaml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(skillVersion.SpecYaml))
		return
	}

	writeEnvelope(w, http.StatusOK, "req_skill_spec", map[string]string{
		"spec_yaml": skillVersion.SpecYaml,
	})
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	if !isJSONRequest(r) {
		writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
		return
	}

	var req AgentRegistration
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if req.AgentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id is required.")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "name is required.")
		return
	}

	agent := Agent{
		AgentID:      req.AgentID,
		Name:         req.Name,
		Version:      req.Version,
		Capabilities: req.Capabilities,
		RegisteredAt: time.Now().UTC(),
	}

	writeEnvelope(w, http.StatusCreated, generateRequestID(), agent)
}

func (s *Server) handleProcedureDrafts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	writeEnvelope(w, http.StatusOK, "req_procedure_drafts", map[string]any{
		"items": s.repo.ListProcedureDrafts(r.URL.Query().Get("status")),
	})
}

func (s *Server) handleExecutions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		limit, offset := parsePagination(r.URL.Query())
		items := s.repo.ListExecutions(r.URL.Query().Get("status"))
		total := len(items)
		if offset > total {
			offset = total
		}
		end := offset + limit
		if end > total {
			end = total
		}
		paginatedItems := items[offset:end]
		pagination := &Pagination{Limit: limit, Offset: offset, Total: total}
		writeEnvelopeWithPagination(w, http.StatusOK, generateRequestID(), map[string]any{
			"items": paginatedItems,
		}, pagination)
	case http.MethodPost:
		if !isJSONRequest(r) {
			writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
			return
		}

		var req CreateExecutionRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}

		execution, err := s.execSvc.CreateExecution(service.CreateExecutionRequest(req))
		if err != nil {
			switch {
			case errors.Is(err, service.ErrSkillIDRequired):
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "skill_id is required.")
			case errors.Is(err, service.ErrTriggeredByRequired):
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "triggered_by is required.")
			case errors.Is(err, service.ErrSkillNotFound):
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Skill not found.")
			default:
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create execution.")
			}
			return
		}

		if logger != nil {
			logger.Info("execution created",
				slog.String("execution_id", execution.ID),
				slog.String("skill_id", execution.SkillID),
				slog.String("triggered_by", execution.TriggeredBy),
			)
		}
		writeEnvelope(w, http.StatusCreated, generateRequestID(), execution)
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) handleExecutionByID(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/approve") {
		s.handleExecutionApproval(w, r)
		return
	}

	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/executions/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Execution not found.")
		return
	}

	execution, ok := s.repo.GetExecution(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Execution not found.")
		return
	}

	writeEnvelope(w, http.StatusOK, "req_execution_detail", execution)
}

func (s *Server) handleApprovals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	limit, offset := parsePagination(r.URL.Query())
	items := s.repo.ListApprovals(r.URL.Query().Get("status"))
	total := len(items)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	paginatedItems := items[offset:end]
	pagination := &Pagination{Limit: limit, Offset: offset, Total: total}
	writeEnvelopeWithPagination(w, http.StatusOK, generateRequestID(), map[string]any{
		"items": paginatedItems,
	}, pagination)
}

func (s *Server) handleExecutionApproval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	if !isJSONRequest(r) {
		writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
		return
	}

	var req ApproveExecutionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/executions/"), "/approve")
	approval, err := s.execSvc.ApproveExecution(id, service.ApproveExecutionRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrApproverRequired):
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "approver is required.")
		case errors.Is(err, service.ErrInvalidDecision):
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "decision must be approved or rejected.")
		case errors.Is(err, service.ErrExecutionNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Approval not found.")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to process approval.")
		}
		return
	}

	if logger != nil {
		logger.Info("approval decision",
			slog.String("execution_id", id),
			slog.String("decision", req.Decision),
			slog.String("approver", req.Approver),
		)
	}
	writeEnvelope(w, http.StatusOK, generateRequestID(), approval)
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return fmt.Errorf("invalid json body: %w", err)
	}

	return nil
}

func isJSONRequest(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("Content-Type"), "application/json")
}

func validRiskLevel(level string) bool {
	switch level {
	case "low", "medium", "high":
		return true
	default:
		return false
	}
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func parsePagination(query map[string][]string) (limit int, offset int) {
	limit = 20
	offset = 0
	if l, ok := query["limit"]; ok && len(l) > 0 {
		if v, err := strconv.Atoi(l[0]); err == nil && v > 0 {
			limit = v
		}
	}
	if o, ok := query["offset"]; ok && len(o) > 0 {
		if v, err := strconv.Atoi(o[0]); err == nil && v >= 0 {
			offset = v
		}
	}
	return limit, offset
}

func writeEnvelope(w http.ResponseWriter, status int, requestID string, payload any) {
	writeJSON(w, status, Envelope{
		Data: payload,
		Meta: Meta{
			RequestID: requestID,
		},
	})
}

func writeEnvelopeWithPagination(w http.ResponseWriter, status int, requestID string, payload any, pagination *Pagination) {
	writeJSON(w, status, Envelope{
		Data: payload,
		Meta: Meta{
			RequestID:  requestID,
			Pagination: pagination,
		},
	})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, Envelope{
		Error: &APIError{
			Code:    code,
			Message: message,
		},
		Meta: Meta{
			RequestID: generateRequestID(),
		},
	})
}

func writeMethodNotAllowed(w http.ResponseWriter, method string) {
	writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", fmt.Sprintf("Method %s is not allowed.", method))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
	}
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		skillID := r.URL.Query().Get("skill_id")
		difficulty := r.URL.Query().Get("difficulty")
		items := s.repo.ListTasks(skillID, difficulty)
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"items": items})
	case http.MethodPost:
		if !isJSONRequest(r) {
			writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
			return
		}
		var req CreateTaskRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
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
		if task.Tags == nil {
			task.Tags = []string{}
		}
		if task.Difficulty == "" {
			task.Difficulty = "medium"
		}
		task, err := s.repo.CreateTask(task)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create task.")
			return
		}
		writeEnvelope(w, http.StatusCreated, generateRequestID(), task)
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	switch r.Method {
	case http.MethodGet:
		task, ok := s.repo.GetTask(id)
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "task not found.")
			return
		}
		writeEnvelope(w, http.StatusOK, generateRequestID(), Task(task))
	case http.MethodDelete:
		err := s.repo.DeleteTask(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete task.")
			return
		}
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]string{"status": "deleted"})
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items := s.repo.ListMetrics()
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"items": items})
	case http.MethodPost:
		if !isJSONRequest(r) {
			writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
			return
		}
		var req CreateMetricRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		metric := Metric{
			ID:        fmt.Sprintf("metric_%d", time.Now().UnixNano()),
			Name:      req.Name,
			Type:      req.Type,
			Config:    req.Config,
			CreatedAt: time.Now(),
		}
		if metric.Config == nil {
			metric.Config = map[string]any{}
		}
		metric, err := s.repo.CreateMetric(metric)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create metric.")
			return
		}
		writeEnvelope(w, http.StatusCreated, generateRequestID(), metric)
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) handleTaskExecutions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items := s.repo.ListTaskExecutions()
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"items": items})
	case http.MethodPost:
		if !isJSONRequest(r) {
			writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
			return
		}
		var req RunTaskRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		taskExec := TaskExecution{
			ID:        fmt.Sprintf("texec_%d", time.Now().UnixNano()),
			TaskID:    req.TaskID,
			AgentID:   req.AgentID,
			Status:    "completed",
			Input:     req.Input,
			Output:    map[string]any{"result": "simulated output"},
			CreatedAt: time.Now(),
		}
		taskExec, err := s.repo.CreateTaskExecution(taskExec)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create task execution.")
			return
		}
		writeEnvelope(w, http.StatusCreated, generateRequestID(), TaskExecution(taskExec))
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) handleTaskExecutionByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/task-executions/")
	path := r.URL.Path

	switch r.Method {
	case http.MethodGet:
		if strings.Contains(path, "/evaluate") {
			execID := strings.TrimSuffix(id, "/evaluate")
			items := s.repo.ListEvaluations(execID)
			writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"items": items})
			return
		}
		ex, ok := s.repo.GetTaskExecution(id)
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "task execution not found.")
			return
		}
		writeEnvelope(w, http.StatusOK, generateRequestID(), TaskExecution(ex))
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) handleExperiments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items := s.repo.ListExperiments()
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"items": items})
	case http.MethodPost:
		if !isJSONRequest(r) {
			writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
			return
		}
		var req CreateExperimentRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		now := time.Now()
		exp := Experiment{
			ID:          fmt.Sprintf("exp_%d", time.Now().UnixNano()),
			Name:        req.Name,
			Description: req.Description,
			TaskIDs:     req.TaskIDs,
			AgentIDs:    req.AgentIDs,
			Status:      "pending",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if exp.TaskIDs == nil {
			exp.TaskIDs = []string{}
		}
		if exp.AgentIDs == nil {
			exp.AgentIDs = []string{}
		}
		exp, err := s.repo.CreateExperiment(exp)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create experiment.")
			return
		}
		writeEnvelope(w, http.StatusCreated, generateRequestID(), Experiment(exp))
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) handleExperimentByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/experiments/")
	switch r.Method {
	case http.MethodGet:
		exp, ok := s.repo.GetExperiment(id)
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "experiment not found.")
			return
		}
		writeEnvelope(w, http.StatusOK, generateRequestID(), Experiment(exp))
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

// handleExperimentCompare handles GET /api/v1/experiments/{id}/compare?compare_with={other_id}
func (s *Server) handleExperimentCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "")
		return
	}

	ids := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/experiments/"), "/")
	if len(ids) < 2 || ids[1] == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "experiment id required")
		return
	}
	expID := ids[0]
	compareWithID := r.URL.Query().Get("compare_with")

	if compareWithID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_PARAM", "compare_with query param required")
		return
	}

	exp, ok := s.repo.GetExperiment(expID)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "experiment not found")
		return
	}
	other, ok := s.repo.GetExperiment(compareWithID)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "compare_with experiment not found")
		return
	}

	// Fetch experiment runs for both experiments
	expRuns := s.repo.ListExperimentRuns(expID)
	otherRuns := s.repo.ListExperimentRuns(compareWithID)

	// Compute average scores per task for each experiment
	expAverages := computeTaskAverages(expRuns)
	otherAverages := computeTaskAverages(otherRuns)

	// Get union of all task IDs
	allTasks := unionKeys(expAverages, otherAverages)

	// Compute deltas for each task
	var deltas []taskDelta
	regressionDetected := false

	for _, taskID := range allTasks {
		expScore := expAverages[taskID]
		otherScore := otherAverages[taskID]
		delta := expScore - otherScore

		deltas = append(deltas, taskDelta{
			TaskID:     taskID,
			ExpScore:   round2(expScore),
			OtherScore: round2(otherScore),
			Delta:      round2(delta),
		})

		if detectRegression(delta, 0.02) {
			regressionDetected = true
		}
	}

	writeJSON(w, http.StatusOK, Envelope{
		Data: map[string]any{
			"experiment":   exp.Name,
			"compare_with": other.Name,
			"deltas":       deltas,
			"regression":   regressionDetected,
		},
	})
}

type taskDelta struct {
	TaskID     string  `json:"task_id"`
	ExpScore   float64 `json:"exp_score"`
	OtherScore float64 `json:"other_score"`
	Delta      float64 `json:"delta"`
}

func round2(f float64) float64 {
	return math.Round(f*100) / 100
}

func computeTaskAverages(runs []ExperimentRun) map[string]float64 {
	sums := make(map[string]float64)
	counts := make(map[string]int)

	for _, run := range runs {
		if run.Status == "completed" {
			sums[run.TaskID] += run.OverallScore
			counts[run.TaskID]++
		}
	}

	averages := make(map[string]float64)
	for taskID, sum := range sums {
		if counts[taskID] > 0 {
			averages[taskID] = sum / float64(counts[taskID])
		}
	}
	return averages
}

func unionKeys(m1, m2 map[string]float64) []string {
	keys := make(map[string]bool)
	for k := range m1 {
		keys[k] = true
	}
	for k := range m2 {
		keys[k] = true
	}
	result := make([]string, 0, len(keys))
	for k := range keys {
		result = append(result, k)
	}
	return result
}

func detectRegression(delta, threshold float64) bool {
	return delta < -threshold
}

func (s *Server) handleCheckRegression(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "")
		return
	}

	var req struct {
		CurrentScore  float64 `json:"current_score"`
		BaselineScore float64 `json:"baseline_score"`
		Threshold     float64 `json:"threshold"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}

	if req.Threshold == 0 {
		req.Threshold = 0.02
	}

	delta := req.CurrentScore - req.BaselineScore
	regression := delta < -req.Threshold

	writeJSON(w, http.StatusOK, Envelope{
		Data: map[string]any{
			"current_score":  req.CurrentScore,
			"baseline_score": req.BaselineScore,
			"delta":          round2(delta),
			"threshold":      req.Threshold,
			"regression":     regression,
			"passed":         !regression,
		},
	})
}

func (s *Server) handleReplaySnapshots(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		executionID := r.URL.Query().Get("execution_id")
		items := s.repo.ListReplaySnapshots(executionID)
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"items": items})
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) handleReplaySnapshotByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/replay-snapshots/")
	switch r.Method {
	case http.MethodGet:
		snap, ok := s.repo.GetReplaySnapshot(id)
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "replay snapshot not found.")
			return
		}
		writeEnvelope(w, http.StatusOK, generateRequestID(), ReplaySnapshot(snap))
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}
