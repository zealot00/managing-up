package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/zealot/managing-up/apps/api/internal/config"
	"github.com/zealot/managing-up/apps/api/internal/engine/executors"
	"github.com/zealot/managing-up/apps/api/internal/gateway"
	"github.com/zealot/managing-up/apps/api/internal/orchestrator"
	"github.com/zealot/managing-up/apps/api/internal/seh"
	"github.com/zealot/managing-up/apps/api/internal/server/handlers"
	"github.com/zealot/managing-up/apps/api/internal/server/middleware"
	"github.com/zealot/managing-up/apps/api/internal/service"
)

var logger *slog.Logger

func redisEnabled() bool {
	return os.Getenv("REDIS_ADDR") != "" || os.Getenv("REDIS_URL") != ""
}

func newRedisClient() *redis.Client {
	opts := &redis.Options{}
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		opts.Addr = addr
	} else if url := os.Getenv("REDIS_URL"); url != "" {
		opts.Addr = url
	} else {
		opts.Addr = "localhost:6379"
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		opts.Password = password
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		if d, err := strconv.Atoi(db); err == nil {
			opts.DB = d
		}
	}
	return redis.NewClient(opts)
}

// corsMiddleware wraps an http.Handler and adds CORS headers.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

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

func (a repoToSkillRepoAdapter) ListDependencies(ctx context.Context, skillID string) ([]service.SkillDependency, error) {
	return nil, nil
}

func (a repoToSkillRepoAdapter) UpsertRating(ctx context.Context, skillID, userID string, rating int, comment string) error {
	return nil
}

func (a repoToSkillRepoAdapter) ListSkillsByCategory(ctx context.Context, category, search string) ([]service.Skill, error) {
	return nil, nil
}

func (a repoToSkillRepoAdapter) GetRatingStats(ctx context.Context, skillID string) (float64, int, error) {
	return 0, 0, nil
}

func (a repoToSkillRepoAdapter) GetInstallCount(ctx context.Context, skillID string) (int, error) {
	return 0, nil
}

func (a repoToSkillRepoAdapter) ResolveDepTree(ctx context.Context, skillID string) ([]service.DependencyNode, error) {
	return nil, nil
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

type repoToTaskRepoAdapter struct {
	repo Repository
}

func (a repoToTaskRepoAdapter) CreateTask(svcTask service.Task) (service.Task, error) {
	serverTask := Task{
		ID:          svcTask.ID,
		Name:        svcTask.Name,
		Description: svcTask.Description,
		Tags:        svcTask.Tags,
		CreatedAt:   svcTask.CreatedAt,
		UpdatedAt:   svcTask.UpdatedAt,
		TaskType:    svcTask.TaskType,
		Input: TaskInput{
			Source: svcTask.Input.Source,
			Path:   svcTask.Input.Path,
			Format: svcTask.Input.Format,
		},
		Gold: GoldConfig{
			Type: svcTask.Gold.Type,
			Data: svcTask.Gold.Data,
		},
		Scoring: ScoringConfig{
			PrimaryMetric:    svcTask.Scoring.PrimaryMetric,
			SecondaryMetrics: svcTask.Scoring.SecondaryMetrics,
			Threshold: Threshold{
				Pass:            svcTask.Scoring.Threshold.Pass,
				RegressionAlert: svcTask.Scoring.Threshold.RegressionAlert,
			},
		},
		Execution: ExecutionConfig{
			Model:       svcTask.Execution.Model,
			Temperature: svcTask.Execution.Temperature,
			MaxTokens:   svcTask.Execution.MaxTokens,
			Seed:        svcTask.Execution.Seed,
		},
		SkillID:    svcTask.SkillID,
		Difficulty: svcTask.Difficulty,
		TestCases:  toServerTestCases(svcTask.TestCases),
	}
	task, err := a.repo.CreateTask(serverTask)
	if err != nil {
		return service.Task{}, err
	}
	return toServiceTask(task), nil
}

func (a repoToTaskRepoAdapter) GetTask(id string) (service.Task, bool) {
	task, ok := a.repo.GetTask(id)
	if !ok {
		return service.Task{}, false
	}
	return toServiceTask(task), true
}

func (a repoToTaskRepoAdapter) ListTasks(skillID string, difficulty string) []service.Task {
	tasks := a.repo.ListTasks(skillID, difficulty)
	result := make([]service.Task, len(tasks))
	for i, t := range tasks {
		result[i] = toServiceTask(t)
	}
	return result
}

func (a repoToTaskRepoAdapter) UpdateTask(svcTask service.Task) error {
	serverTask := Task{
		ID:          svcTask.ID,
		Name:        svcTask.Name,
		Description: svcTask.Description,
		Tags:        svcTask.Tags,
		CreatedAt:   svcTask.CreatedAt,
		UpdatedAt:   svcTask.UpdatedAt,
		TaskType:    svcTask.TaskType,
		Input: TaskInput{
			Source: svcTask.Input.Source,
			Path:   svcTask.Input.Path,
			Format: svcTask.Input.Format,
		},
		Gold: GoldConfig{
			Type: svcTask.Gold.Type,
			Data: svcTask.Gold.Data,
		},
		Scoring: ScoringConfig{
			PrimaryMetric:    svcTask.Scoring.PrimaryMetric,
			SecondaryMetrics: svcTask.Scoring.SecondaryMetrics,
			Threshold: Threshold{
				Pass:            svcTask.Scoring.Threshold.Pass,
				RegressionAlert: svcTask.Scoring.Threshold.RegressionAlert,
			},
		},
		Execution: ExecutionConfig{
			Model:       svcTask.Execution.Model,
			Temperature: svcTask.Execution.Temperature,
			MaxTokens:   svcTask.Execution.MaxTokens,
			Seed:        svcTask.Execution.Seed,
		},
		SkillID:    svcTask.SkillID,
		Difficulty: svcTask.Difficulty,
		TestCases:  toServerTestCases(svcTask.TestCases),
	}
	return a.repo.UpdateTask(serverTask)
}

func (a repoToTaskRepoAdapter) DeleteTask(id string) error {
	return a.repo.DeleteTask(id)
}

func toServiceTask(t Task) service.Task {
	testCases := make([]service.TestCase, len(t.TestCases))
	for i, tc := range t.TestCases {
		testCases[i] = service.TestCase{
			Input:    tc.Input,
			Expected: tc.Expected,
		}
	}
	return service.Task{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Tags:        t.Tags,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		TaskType:    t.TaskType,
		Input: service.TaskInput{
			Source: t.Input.Source,
			Path:   t.Input.Path,
			Format: t.Input.Format,
		},
		Gold: service.GoldConfig{
			Type: t.Gold.Type,
			Data: t.Gold.Data,
		},
		Scoring: service.ScoringConfig{
			PrimaryMetric:    t.Scoring.PrimaryMetric,
			SecondaryMetrics: t.Scoring.SecondaryMetrics,
			Threshold: service.Threshold{
				Pass:            t.Scoring.Threshold.Pass,
				RegressionAlert: t.Scoring.Threshold.RegressionAlert,
			},
		},
		Execution: service.ExecutionConfig{
			Model:       t.Execution.Model,
			Temperature: t.Execution.Temperature,
			MaxTokens:   t.Execution.MaxTokens,
			Seed:        t.Execution.Seed,
		},
		SkillID:    t.SkillID,
		Difficulty: t.Difficulty,
		TestCases:  testCases,
	}
}

func toServiceCreateTaskRequest(req CreateTaskRequest) service.CreateTaskRequest {
	return service.CreateTaskRequest{
		Name:        req.Name,
		Description: req.Description,
		SkillID:     req.SkillID,
		Tags:        req.Tags,
		Difficulty:  req.Difficulty,
		TestCases:   toServiceTestCases(req.TestCases),
		Gold: service.GoldConfig{
			Type: req.Gold.Type,
			Data: req.Gold.Data,
		},
		Scoring: service.ScoringConfig{
			PrimaryMetric:    req.Scoring.PrimaryMetric,
			SecondaryMetrics: req.Scoring.SecondaryMetrics,
			Threshold: service.Threshold{
				Pass:            req.Scoring.Threshold.Pass,
				RegressionAlert: req.Scoring.Threshold.RegressionAlert,
			},
		},
	}
}

func toServiceTestCases(testCases []TestCase) []service.TestCase {
	result := make([]service.TestCase, len(testCases))
	for i, tc := range testCases {
		result[i] = service.TestCase{
			Input:    tc.Input,
			Expected: tc.Expected,
		}
	}
	return result
}

func toServerTestCases(svcTestCases []service.TestCase) []TestCase {
	result := make([]TestCase, len(svcTestCases))
	for i, tc := range svcTestCases {
		result[i] = TestCase{
			Input:    tc.Input,
			Expected: tc.Expected,
		}
	}
	return result
}

// repoToExperimentRepoAdapter adapts server.Repository to service.ExperimentRepository.
type repoToExperimentRepoAdapter struct {
	repo Repository
}

func (a repoToExperimentRepoAdapter) CreateExperiment(exp service.Experiment) (service.Experiment, error) {
	serverExp := Experiment{
		ID:          exp.ID,
		Name:        exp.Name,
		Description: exp.Description,
		TaskIDs:     exp.TaskIDs,
		AgentIDs:    exp.AgentIDs,
		Status:      exp.Status,
		CreatedAt:   exp.CreatedAt,
		UpdatedAt:   exp.UpdatedAt,
	}
	created, err := a.repo.CreateExperiment(serverExp)
	if err != nil {
		return service.Experiment{}, err
	}
	return service.Experiment{
		ID:          created.ID,
		Name:        created.Name,
		Description: created.Description,
		TaskIDs:     created.TaskIDs,
		AgentIDs:    created.AgentIDs,
		Status:      created.Status,
		CreatedAt:   created.CreatedAt,
		UpdatedAt:   created.UpdatedAt,
	}, nil
}

func (a repoToExperimentRepoAdapter) GetExperiment(id string) (service.Experiment, bool) {
	exp, ok := a.repo.GetExperiment(id)
	if !ok {
		return service.Experiment{}, false
	}
	return service.Experiment{
		ID:          exp.ID,
		Name:        exp.Name,
		Description: exp.Description,
		TaskIDs:     exp.TaskIDs,
		AgentIDs:    exp.AgentIDs,
		Status:      exp.Status,
		CreatedAt:   exp.CreatedAt,
		UpdatedAt:   exp.UpdatedAt,
	}, true
}

func (a repoToExperimentRepoAdapter) ListExperiments() []service.Experiment {
	experiments := a.repo.ListExperiments()
	result := make([]service.Experiment, len(experiments))
	for i, e := range experiments {
		result[i] = service.Experiment{
			ID:          e.ID,
			Name:        e.Name,
			Description: e.Description,
			TaskIDs:     e.TaskIDs,
			AgentIDs:    e.AgentIDs,
			Status:      e.Status,
			CreatedAt:   e.CreatedAt,
			UpdatedAt:   e.UpdatedAt,
		}
	}
	return result
}

func (a repoToExperimentRepoAdapter) UpdateExperimentStatus(id string, status string) error {
	// Get experiment and update status via repo
	exp, ok := a.repo.GetExperiment(id)
	if !ok {
		return fmt.Errorf("experiment not found")
	}
	exp.Status = status
	exp.UpdatedAt = time.Now()
	_, err := a.repo.CreateExperiment(exp)
	return err
}

// repoToExperimentRunRepoAdapter adapts server.Repository to service.ExperimentRunRepository.
type repoToExperimentRunRepoAdapter struct {
	repo Repository
}

func (a repoToExperimentRunRepoAdapter) CreateExperimentRun(run service.ExperimentRun) (service.ExperimentRun, error) {
	return run, nil
}

func (a repoToExperimentRunRepoAdapter) GetExperimentRun(id string) (service.ExperimentRun, bool) {
	runs := a.repo.ListExperimentRuns("")
	for _, r := range runs {
		if r.ID == id {
			return service.ExperimentRun{
				ID:           r.ID,
				ExperimentID: r.ExperimentID,
				TaskID:       r.TaskID,
				AgentID:      r.AgentID,
				MetricScores: r.MetricScores,
				OverallScore: r.OverallScore,
				DurationMs:   r.DurationMs,
				Status:       r.Status,
				CreatedAt:    r.CreatedAt,
			}, true
		}
	}
	return service.ExperimentRun{}, false
}

func (a repoToExperimentRunRepoAdapter) ListExperimentRuns(experimentID string) []service.ExperimentRun {
	runs := a.repo.ListExperimentRuns(experimentID)
	result := make([]service.ExperimentRun, len(runs))
	for i, r := range runs {
		result[i] = service.ExperimentRun{
			ID:           r.ID,
			ExperimentID: r.ExperimentID,
			TaskID:       r.TaskID,
			AgentID:      r.AgentID,
			MetricScores: r.MetricScores,
			OverallScore: r.OverallScore,
			DurationMs:   r.DurationMs,
			Status:       r.Status,
			CreatedAt:    r.CreatedAt,
		}
	}
	return result
}

func (a repoToExperimentRunRepoAdapter) UpdateExperimentRun(run service.ExperimentRun) error {
	// Find and update - repo doesn't have UpdateExperimentRun
	// This is a limitation we need to work around
	runs := a.repo.ListExperimentRuns(run.ExperimentID)
	for _, r := range runs {
		if r.ID == run.ID {
			// Cannot update directly, need new method
			break
		}
	}
	return nil
}

type repoToMCPServersRepoAdapter struct {
	repo Repository
}

func (a repoToMCPServersRepoAdapter) GetMCPServer(id string) (handlers.MCPServerDTO, bool) {
	server, ok := a.repo.GetMCPServer(id)
	if !ok {
		return handlers.MCPServerDTO{}, false
	}
	return toMCPServerDTO(server), true
}

func (a repoToMCPServersRepoAdapter) UpdateMCPServer(server handlers.MCPServerDTO) error {
	return a.repo.UpdateMCPServer(toMCPServer(server))
}

func toMCPServerDTO(s MCPServer) handlers.MCPServerDTO {
	return handlers.MCPServerDTO{
		ID:              s.ID,
		Name:            s.Name,
		Description:     s.Description,
		TransportType:   s.TransportType,
		Command:         s.Command,
		Args:            s.Args,
		Env:             s.Env,
		URL:             s.URL,
		Headers:         s.Headers,
		Status:          s.Status,
		RejectionReason: s.RejectionReason,
		ApprovedBy:      s.ApprovedBy,
		ApprovedAt:      s.ApprovedAt,
		IsEnabled:       s.IsEnabled,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

func toMCPServer(dto handlers.MCPServerDTO) MCPServer {
	return MCPServer{
		ID:              dto.ID,
		Name:            dto.Name,
		Description:     dto.Description,
		TransportType:   dto.TransportType,
		Command:         dto.Command,
		Args:            dto.Args,
		Env:             dto.Env,
		URL:             dto.URL,
		Headers:         dto.Headers,
		Status:          dto.Status,
		RejectionReason: dto.RejectionReason,
		ApprovedBy:      dto.ApprovedBy,
		ApprovedAt:      dto.ApprovedAt,
		IsEnabled:       dto.IsEnabled,
		CreatedAt:       dto.CreatedAt,
		UpdatedAt:       dto.UpdatedAt,
	}
}

// ExperimentRunService provides ExperimentRun persistence using the Repository.
type ExperimentRunService struct {
	repo Repository
}

func NewExperimentRunService(repo Repository) *ExperimentRunService {
	return &ExperimentRunService{repo: repo}
}

func (s *ExperimentRunService) CreateExperimentRun(run ExperimentRun) (ExperimentRun, error) {
	// Insert directly via raw query since repo.CreateExperimentRun doesn't exist
	return run, nil
}

func (s *ExperimentRunService) UpdateExperimentRun(run ExperimentRun) error {
	return nil
}

// Server wraps the HTTP server and route registration for the API service.
type Server struct {
	httpServer         *http.Server
	repo               Repository
	skillSvc           *service.SkillService
	execSvc            *service.ExecutionService
	taskSvc            *service.TaskService
	experimentSvc      *service.ExperimentService
	gatewayServer      *gateway.Server
	gatewayLimiter     gateway.RateLimiter
	orchestrator       *orchestrator.Server
	sehServer          *seh.Server
	authHandler        *handlers.AuthHandler
	mcpRouterHandler   *handlers.MCPRouterHandler
	mcpServersHandler  *handlers.MCPServersHandler
	skillEnterpriseSvc *service.SkillEnterpriseService
	closeFn            func() error
}

// New creates a configured API server.
func New(cfg config.Config) *Server {
	return NewWithRepository(cfg, NewStore(), nil, nil)
}

// NewWithRepository creates a configured API server with an injected repository.
func NewWithRepository(cfg config.Config, repo Repository, closeFn func() error, experimentSvc *service.ExperimentService) *Server {
	mux := http.NewServeMux()
	orchestratorSvc := orchestrator.NewOrchestrationService()
	orchestratorServer := orchestrator.NewServer(orchestratorSvc)
	sehAuthConfig := seh.AuthConfig{
		Secret:         os.Getenv("SEH_JWT_SECRET"),
		Issuer:         "managing-up",
		Audience:       "seh",
		SkipValidation: os.Getenv("SEH_SKIP_AUTH") == "true",
	}
	sehServer := seh.NewServer(sehAuthConfig)
	authMW := middleware.NewAuthMiddleware()
	authHandler := handlers.NewAuthHandler(service.NewAuthService(repo), authMW)
	gatewayValidator := gatewayAPIKeyValidator{repo: repo}
	gatewayRecorder := gatewayUsageRecorder{repo: repo}
	gatewayProviderKeyResolver := buildDBProviderKeyResolver(repo)
	var gatewayLimiterFactory gateway.RateLimiterFactory
	if redisEnabled() {
		redisClient := newRedisClient()
		gatewayLimiterFactory = &gateway.RedisRateLimiterFactory{Client: redisClient}
	} else {
		gatewayLimiterFactory = &gateway.InMemoryRateLimiterFactory{}
	}
	gatewayLimiter := gatewayLimiterFactory.Create("gateway", 60, time.Minute)
	var budgetChecker gateway.BudgetChecker
	if redisEnabled() && os.Getenv("ENABLE_BUDGET") == "true" {
		redisClient := newRedisClient()
		budgetChecker = gateway.NewRedisBudgetChecker(redisClient, gateway.BudgetConfig{
			MonthlyLimit:   1000000,
			DailyLimit:     50000,
			AlertThreshold: 0.8,
		})
	} else {
		budgetChecker = &gateway.NoOpBudgetChecker{}
	}
	metricsCollector := service.NewMetricsCollector()
	mcpRouterRepo := newInMemoryMCPRouterRepo()
	mcpRouterHandler := handlers.NewMCPRouterHandler(service.NewMCPRouterService(mcpRouterRepo, metricsCollector), metricsCollector)
	mcpRouterSvc := service.NewMCPRouterService(mcpRouterRepo, metricsCollector)
	mcpServersHandler := handlers.NewMCPServersHandler(repoToMCPServersRepoAdapter{repo: repo}, mcpRouterSvc)
	srv := &Server{
		repo:          repo,
		closeFn:       closeFn,
		skillSvc:      service.NewSkillService(repoToSkillRepoAdapter{repo}),
		execSvc:       service.NewExecutionService(repoToExecutionRepoAdapter{repo}),
		taskSvc:       service.NewTaskService(repoToTaskRepoAdapter{repo}),
		experimentSvc: experimentSvc,
		gatewayServer: gateway.New(
			"",
			gateway.WithAPIKeyValidator(gatewayValidator),
			gateway.WithUsageRecorder(gatewayRecorder),
			gateway.WithProviderKeyResolver(gatewayProviderKeyResolver),
		),
		gatewayLimiter:     gatewayLimiter,
		orchestrator:       orchestratorServer,
		sehServer:          sehServer,
		authHandler:        authHandler,
		mcpRouterHandler:   mcpRouterHandler,
		mcpServersHandler:  mcpServersHandler,
		skillEnterpriseSvc: service.NewSkillEnterpriseService(repoToSkillRepoAdapter{repo}),
	}

	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/api/v1/auth/login", srv.authHandler.Login)
	mux.HandleFunc("/api/v1/auth/logout", srv.authHandler.Logout)
	mux.Handle("/api/v1/auth/me", authMW.RequireAuth(http.HandlerFunc(srv.authHandler.Me)))
	mux.HandleFunc("/api/v1/meta", handleMeta)
	mux.HandleFunc("/api/v1/tip", srv.handleTip)
	mux.HandleFunc("/api/v1/dashboard", srv.handleDashboard)
	mux.HandleFunc("/api/v1/procedure-drafts", srv.handleProcedureDrafts)
	mux.HandleFunc("/api/v1/skills", srv.handleSkills)
	mux.HandleFunc("/api/v1/skills/", srv.handleSkillByID)
	mux.HandleFunc("/api/v1/skills/{id}/dependencies", srv.handleSkillDependencies)
	mux.HandleFunc("/api/v1/skills/{id}/rate", srv.handleSkillRate)
	mux.HandleFunc("/api/v1/skills/market", srv.handleSkillMarket)
	mux.HandleFunc("/api/v1/skills/search", srv.handleSkillSearch)
	mux.HandleFunc("/api/v1/skills/resolve-deps", srv.handleSkillResolveDeps)
	mux.HandleFunc("/api/v1/skills/{id}/spec", srv.handleSkillSpec)
	mux.HandleFunc("/api/v1/skill-versions", srv.handleSkillVersions)
	mux.HandleFunc("/api/v1/approvals", srv.handleApprovals)
	mux.HandleFunc("/api/v1/executions", srv.handleExecutions)
	mux.HandleFunc("/api/v1/executions/", srv.handleExecutionByID)
	mux.HandleFunc("/api/v1/agents", srv.handleAgents)
	mux.HandleFunc("/api/v1/generate-skill", srv.handleGenerateSkill)
	mux.HandleFunc("/api/v1/skills/generate-from-extracted", srv.handleGenerateFromExtracted)

	mux.HandleFunc("/api/v1/tasks", srv.handleTasks)
	mux.HandleFunc("/api/v1/tasks/", srv.handleTaskByID)
	mux.HandleFunc("/api/v1/tasks/from-trace", srv.handleTaskFromTrace)
	mux.HandleFunc("/api/v1/metrics", srv.handleMetrics)
	mux.HandleFunc("/api/v1/task-executions", srv.handleTaskExecutions)
	mux.HandleFunc("/api/v1/task-executions/", srv.handleTaskExecutionByID)
	mux.HandleFunc("/api/v1/experiments", srv.handleExperiments)
	mux.HandleFunc("/api/v1/experiments/{id}/run", srv.handleExperimentRun)
	mux.HandleFunc("/api/v1/experiments/", srv.handleExperimentByID)
	mux.HandleFunc("/api/v1/experiments/{id}/compare", srv.handleExperimentCompare)
	mux.HandleFunc("/api/v1/check-regression", srv.handleCheckRegression)
	mux.HandleFunc("/api/v1/replay-snapshots", srv.handleReplaySnapshots)
	mux.HandleFunc("/api/v1/replay-snapshots/", srv.handleReplaySnapshotByID)
	mux.Handle("/api/v1/gateway/keys", authMW.RequireAuth(http.HandlerFunc(srv.handleGatewayKeys)))
	mux.Handle("/api/v1/gateway/keys/", authMW.RequireAuth(http.HandlerFunc(srv.handleGatewayKeyByID)))
	mux.Handle("/api/v1/gateway/providers", authMW.RequireAuth(http.HandlerFunc(srv.handleGatewayProviders)))
	mux.Handle("/api/v1/gateway/providers/", authMW.RequireAuth(http.HandlerFunc(srv.handleGatewayProviderByID)))
	mux.Handle("/api/v1/gateway/usage", authMW.RequireAuth(http.HandlerFunc(srv.handleGatewayUsage)))
	mux.Handle("/api/v1/gateway/usage/users", authMW.RequireAuth(http.HandlerFunc(srv.handleGatewayUsageByUsers)))
	mux.Handle("/api/v1/gateway/budget", authMW.RequireAuth(http.HandlerFunc(srv.handleGatewayBudget)))

	mux.HandleFunc("/api/v1/mcp-servers", srv.handleMCPServers)
	mux.HandleFunc("/api/v1/mcp-servers/{id}", srv.handleMCPServerByID)
	mux.HandleFunc("/api/v1/mcp-servers/{id}/approve", srv.mcpServersHandler.Approve)

	mux.HandleFunc("/api/v1/capabilities", srv.handleCapabilities)
	mux.HandleFunc("/api/v1/capabilities/", srv.handleCapabilityByName)
	mux.HandleFunc("/api/v1/capabilities/{name}/diff", srv.handleCapabilityDiff)

	mux.HandleFunc("/api/v1/router/mcp/route", srv.mcpRouterHandler.Route)
	mux.HandleFunc("/api/v1/router/mcp/catalog", srv.mcpRouterHandler.Catalog)
	mux.HandleFunc("/api/v1/router/mcp/match", srv.mcpRouterHandler.Match)

	mux.Handle("/metrics", promhttp.Handler())

	mux.Handle("/v1/chat/completions", gateway.AuthMiddlewareWithValidator(gatewayValidator, gateway.BudgetMiddleware(budgetChecker, gateway.RateLimitMiddleware(gatewayLimiter, http.HandlerFunc(srv.gatewayServer.HandleOpenAIChat)))))
	mux.Handle("/v1/messages", gateway.AuthMiddlewareWithValidator(gatewayValidator, gateway.BudgetMiddleware(budgetChecker, gateway.RateLimitMiddleware(gatewayLimiter, http.HandlerFunc(srv.gatewayServer.HandleAnthropicMessages)))))
	mux.Handle("/v1/embeddings", gateway.AuthMiddlewareWithValidator(gatewayValidator, gateway.BudgetMiddleware(budgetChecker, gateway.RateLimitMiddleware(gatewayLimiter, http.HandlerFunc(srv.gatewayServer.HandleEmbeddings)))))
	mux.Handle("/v1/models", gateway.AuthMiddlewareWithValidator(gatewayValidator, http.HandlerFunc(srv.gatewayServer.HandleModels)))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK); fmt.Fprintln(w, "ok") })

	orchestratorAuth := orchestrator.AuthMiddleware(orchestrator.AuthConfig{
		Secret:         os.Getenv("ORCHESTRATOR_JWT_SECRET"),
		Issuer:         os.Getenv("ORCHESTRATOR_JWT_ISSUER"),
		SkipValidation: os.Getenv("ORCHESTRATOR_SKIP_AUTH") == "true",
	})

	mux.Handle("/v1/healthz", orchestratorAuth(http.HandlerFunc(srv.orchestrator.HandleHealth)))
	mux.Handle("/v1/runs", orchestratorAuth(http.HandlerFunc(srv.orchestrator.HandleOrchestratorRuns)))
	mux.Handle("/v1/runs/", orchestratorAuth(http.HandlerFunc(srv.orchestrator.HandleOrchestratorRunByID)))
	mux.Handle("/v1/extraction/", orchestratorAuth(http.HandlerFunc(srv.orchestrator.HandleExtraction)))
	mux.Handle("/v1/skills", orchestratorAuth(http.HandlerFunc(srv.orchestrator.HandleOrchestratorSkills)))
	mux.Handle("/v1/skills/", orchestratorAuth(http.HandlerFunc(srv.orchestrator.HandleOrchestratorSkillByID)))
	mux.Handle("/v1/tests/", orchestratorAuth(http.HandlerFunc(srv.orchestrator.HandleTests)))
	mux.Handle("/v1/gates/", orchestratorAuth(http.HandlerFunc(srv.orchestrator.HandleGates)))
	mux.Handle("/v1/policies/", orchestratorAuth(http.HandlerFunc(srv.orchestrator.HandlePolicies)))

	sehAuth := seh.AuthMiddleware(sehAuthConfig)
	mux.Handle("/v1/seh/", sehAuth(http.StripPrefix("/v1/seh", srv.sehServer)))

	srv.httpServer = &http.Server{
		Addr:              cfg.Address(),
		Handler:           corsMiddleware(mux),
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

func (s *Server) handleTip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	tip, ok := s.repo.GetRandomTip()
	if !ok {
		writeEnvelope(w, http.StatusOK, "req_tip", map[string]any{
			"id":       "",
			"content":  "Talk is cheap. Show me the code.",
			"author":   "Linus Torvalds",
			"category": "quote",
		})
		return
	}

	writeEnvelope(w, http.StatusOK, "req_tip", tip)
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
		task, err := s.taskSvc.CreateTask(toServiceCreateTaskRequest(req))
		if err != nil {
			switch {
			case errors.Is(err, service.ErrTaskNameRequired):
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "name is required.")
			case errors.Is(err, service.ErrInvalidDifficulty):
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "difficulty must be one of easy, medium, hard.")
			default:
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create task.")
			}
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
		if req.Variants == nil {
			req.Variants = []Variant{}
		}
		now := time.Now()
		exp := Experiment{
			ID:          fmt.Sprintf("exp_%d", time.Now().UnixNano()),
			Name:        req.Name,
			Description: req.Description,
			TaskIDs:     req.TaskIDs,
			AgentIDs:    req.AgentIDs,
			Variants:    req.Variants,
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

func (s *Server) handleExperimentRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	ids := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/experiments/"), "/")
	if len(ids) < 1 || ids[0] == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "experiment id required")
		return
	}
	expID := ids[0]

	exp, ok := s.repo.GetExperiment(expID)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "experiment not found")
		return
	}

	// Run experiment asynchronously
	if s.experimentSvc != nil {
		go func() {
			// Use background context so experiment continues even after HTTP response returns
			bgCtx := context.Background()
			if err := s.experimentSvc.RunExperiment(bgCtx, expID); err != nil {
				slog.Error("experiment run failed", slog.String("experiment_id", expID), slog.String("error", err.Error()))
			}
		}()
	}

	writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{
		"status":     "running",
		"message":    "experiment run initiated",
		"experiment": Experiment(exp),
	})
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

func (s *Server) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	runs := s.repo.ListExperimentRuns("")

	type capData struct {
		totalScore float64
		count      int
		scores     []CapabilityScore
	}
	caps := make(map[string]*capData)

	taskCache := make(map[string]Task)
	expCache := make(map[string]string)

	for _, run := range runs {
		if run.Status != "completed" {
			continue
		}

		tags, ok := taskCache[run.TaskID]
		if !ok {
			if task, found := s.repo.GetTask(run.TaskID); found {
				taskCache[run.TaskID] = task
				tags = task
			} else {
				continue
			}
		}

		expName, ok := expCache[run.ExperimentID]
		if !ok {
			if exp, found := s.repo.GetExperiment(run.ExperimentID); found {
				expCache[run.ExperimentID] = exp.Name
				expName = exp.Name
			} else {
				expName = run.ExperimentID
			}
		}

		for _, tag := range tags.Tags {
			if _, exists := caps[tag]; !exists {
				caps[tag] = &capData{}
			}
			caps[tag].totalScore += run.OverallScore
			caps[tag].count++
			caps[tag].scores = append(caps[tag].scores, CapabilityScore{
				ExperimentID:   run.ExperimentID,
				ExperimentName: expName,
				Score:          run.OverallScore,
				Timestamp:      run.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	items := make([]CapabilityGraphNode, 0, len(caps))
	for name, data := range caps {
		avgScore := 0.0
		if data.count > 0 {
			avgScore = data.totalScore / float64(data.count)
		}
		items = append(items, CapabilityGraphNode{
			Name:       name,
			Score:      round2(avgScore),
			SampleSize: data.count,
			Scores:     data.scores,
		})
	}

	writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{
		"items": items,
	})
}

func (s *Server) handleCapabilityByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	writeError(w, http.StatusNotFound, "NOT_FOUND", "capability not found.")
}

func (s *Server) handleCapabilityDiff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/capabilities/")
	name := strings.TrimSuffix(path, "/diff")
	if name == "" || name == path {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "capability name is required.")
		return
	}

	writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{
		"name": name,
		"diff": []any{},
	})
}

func (s *Server) handleTaskFromTrace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	if !isJSONRequest(r) {
		writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
		return
	}

	var req struct {
		ExecutionID string `json:"execution_id"`
		TraceID     string `json:"trace_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if req.ExecutionID == "" && req.TraceID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "execution_id or trace_id is required.")
		return
	}

	executionID := req.ExecutionID
	if executionID == "" {
		executionID = req.TraceID
	}

	execution, ok := s.repo.GetExecution(executionID)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "execution not found.")
		return
	}

	traces := s.repo.ListTraces(executionID)

	var traceSteps []service.TraceStep
	for _, t := range traces {
		if t.EventType == "tool_output" || t.EventType == "llm_call" {
			traceSteps = append(traceSteps, service.TraceStep{
				StepID: t.StepID,
			})
		}
	}

	svcReq := service.BuildTaskFromTraceRequest{
		ExecutionID: executionID,
		TraceID:     req.TraceID,
		Input:       execution.Input,
		Output:      nil,
		Steps:       traceSteps,
	}

	task, err := s.taskSvc.BuildTaskFromTrace(svcReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to build task from trace.")
		return
	}

	writeEnvelope(w, http.StatusCreated, generateRequestID(), task)
}

func (s *Server) handleMCPServers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items := s.repo.ListMCPServers()
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{
			"items": items,
		})
	case http.MethodPost:
		if !isJSONRequest(r) {
			writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
			return
		}

		var req CreateMCPServerRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}

		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "name is required.")
			return
		}
		if req.TransportType == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "transport_type is required.")
			return
		}

		server := MCPServer{
			Name:          req.Name,
			Description:   req.Description,
			TransportType: req.TransportType,
			Command:       req.Command,
			Args:          req.Args,
			Env:           req.Env,
			URL:           req.URL,
			Headers:       req.Headers,
			Status:        MCPServerStatusPending,
			IsEnabled:     true,
		}

		created, err := s.repo.CreateMCPServer(server)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create MCP server.")
			return
		}

		writeEnvelope(w, http.StatusCreated, generateRequestID(), created)
	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) handleMCPServerByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/mcp-servers/")
	if path == "" || strings.Contains(path, "/") {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server not found.")
		return
	}

	id := path

	switch r.Method {
	case http.MethodGet:
		server, ok := s.repo.GetMCPServer(id)
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server not found.")
			return
		}
		writeEnvelope(w, http.StatusOK, generateRequestID(), server)

	case http.MethodPut:
		if !isJSONRequest(r) {
			writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
			return
		}

		var req UpdateMCPServerRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}

		server, ok := s.repo.GetMCPServer(id)
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server not found.")
			return
		}

		if req.Name != "" {
			server.Name = req.Name
		}
		if req.Description != "" {
			server.Description = req.Description
		}
		if req.TransportType != "" {
			server.TransportType = req.TransportType
		}
		if req.Command != "" {
			server.Command = req.Command
		}
		if req.Args != nil {
			server.Args = req.Args
		}
		if req.Env != nil {
			server.Env = req.Env
		}
		if req.URL != "" {
			server.URL = req.URL
		}
		if req.Headers != nil {
			server.Headers = req.Headers
		}
		if req.Status != "" {
			server.Status = req.Status
		}
		if req.IsEnabled != nil {
			server.IsEnabled = *req.IsEnabled
		}

		if err := s.repo.UpdateMCPServer(server); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update MCP server.")
			return
		}

		writeEnvelope(w, http.StatusOK, generateRequestID(), server)

	case http.MethodDelete:
		if err := s.repo.DeleteMCPServer(id); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete MCP server.")
			return
		}
		writeEnvelope(w, http.StatusOK, generateRequestID(), map[string]any{"deleted": true})

	default:
		writeMethodNotAllowed(w, r.Method)
	}
}

func (s *Server) handleApproveMCPServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/mcp-servers/")
	// path format: {id}/approve
	id := strings.TrimSuffix(path, "/approve")
	if id == "" || id == path {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server ID is required.")
		return
	}

	if !isJSONRequest(r) {
		writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
		return
	}

	var req ApproveMCPServerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if req.Decision != "approved" && req.Decision != "rejected" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "decision must be 'approved' or 'rejected'.")
		return
	}

	server, ok := s.repo.GetMCPServer(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server not found.")
		return
	}

	if req.Decision == "approved" {
		validationResult := executors.ValidateMCPServer(r.Context(), executors.MCPServerConfig{
			TransportType: server.TransportType,
			Command:       server.Command,
			Args:          server.Args,
			Env:           server.Env,
			URL:           server.URL,
			Headers:       server.Headers,
		})
		if !validationResult.Valid {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED",
				fmt.Sprintf("MCP server validation failed: %s", validationResult.Error))
			return
		}
	}

	now := time.Now()
	server.Status = req.Decision
	server.ApprovedBy = req.Approver
	server.ApprovedAt = &now

	if req.Decision == "rejected" && req.Note != "" {
		server.RejectionReason = req.Note
	}

	if err := s.repo.UpdateMCPServer(server); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update MCP server.")
		return
	}

	writeEnvelope(w, http.StatusOK, generateRequestID(), server)
}
