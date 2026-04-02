package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/config"
	"github.com/zealot/managing-up/apps/api/internal/engine"
	"github.com/zealot/managing-up/apps/api/internal/engine/agents"
	"github.com/zealot/managing-up/apps/api/internal/engine/tool/builtin"
	"github.com/zealot/managing-up/apps/api/internal/evaluator"
	"github.com/zealot/managing-up/apps/api/internal/llm"
	"github.com/zealot/managing-up/apps/api/internal/repository/postgres"
	"github.com/zealot/managing-up/apps/api/internal/server"
	"github.com/zealot/managing-up/apps/api/internal/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg := config.Load()
	server.SetLogger(slog.Default())

	if !cfg.Database.Enabled() {
		log.Fatal("DATABASE_URL and DB_DRIVER are required. PostgreSQL is the only supported database.")
	}

	repo, err := postgres.New(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}

	llmClient := createLLMClient()
	agent := createLLMAgent(llmClient)
	evalRunner := createEvaluationRunner(repo, agent, llmClient)
	taskRunnerAdapter := newTaskRunnerAdapter(evalRunner)
	experimentSvc := createExperimentService(repo, taskRunnerAdapter)

	srv := server.NewWithRepository(cfg, repo, repo.Close, experimentSvc)

	execEngine := engine.NewExecutionEngine(repo, engine.NewToolGateway())
	worker := engine.NewWorker(execEngine, repo, 2*time.Second)
	go worker.Start(context.Background())

	startServer(srv)
}

func startServer(srv *server.Server) {
	errCh := make(chan error, 1)

	go func() {
		errCh <- srv.Start()
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server exited with error: %v", err)
		}
	case <-stopCtx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
}

func createLLMClient() llm.Client {
	provider := llm.Provider(os.Getenv("LLM_PROVIDER"))
	if provider == "" {
		provider = llm.ProviderOllama
	}

	model := llm.Model(os.Getenv("LLM_MODEL"))
	if model == "" {
		model = "deepseek-r1-tool-calling:7b"
	}

	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	apiKey := os.Getenv("LLM_API_KEY")

	client, err := llm.NewClient(provider, model, apiKey)
	if err != nil {
		log.Fatalf("failed to create LLM client: %v", err)
	}
	return client
}

func createLLMAgent(client llm.Client) engine.Agent {
	toolRegistry := engine.NewToolRegistry()

	builtinRegistry := builtin.NewRegistry()
	for _, t := range builtinRegistry.List() {
		toolRegistry.Register(t)
	}

	return agents.NewLLMAgent(client, toolRegistry)
}

func createEvaluationRunner(repo server.Repository, agent engine.Agent, llmClient llm.Client) *evaluator.EvaluationRunner {
	registry := evaluator.NewEvaluatorRegistry()
	registry.Register(&evaluator.ExactMatchEvaluator{})
	registry.Register(evaluator.NewSemanticSimilarityEvaluator(0.8))

	embeddingClient, _ := llm.NewClient(llm.ProviderOpenAI, "text-embedding-3-small", os.Getenv("LLM_API_KEY"))
	registry.Register(evaluator.NewEmbeddingSimilarityEvaluator(embeddingClient, 0.85))

	evalRunner := evaluator.NewEvaluationRunner(
		repoToTaskRepoAdapter{repo: repo},
		repoToMetricRepoAdapter{repo: repo},
		repoToTaskExecutionRepoAdapter{repo: repo},
		repoToEvaluationRepoAdapter{repo: repo},
		evaluator.NewJudgeRouter(registry),
		nil,
		agent,
		llmClient,
	)

	return evalRunner
}

func newTaskRunnerAdapter(evalRunner *evaluator.EvaluationRunner) service.TaskRunner {
	return &taskRunnerAdapter{evalRunner: evalRunner}
}

type taskRunnerAdapter struct {
	evalRunner *evaluator.EvaluationRunner
}

func (a *taskRunnerAdapter) RunTask(ctx context.Context, taskID, agentID string, input map[string]any) (service.TaskExecution, error) {
	exec, err := a.evalRunner.RunTask(ctx, taskID, agentID, input)
	if err != nil {
		return service.TaskExecution{}, err
	}
	return service.TaskExecution{
		ID:         exec.ID,
		TaskID:     exec.TaskID,
		AgentID:    exec.AgentID,
		Status:     exec.Status,
		Input:      exec.Input,
		Output:     exec.Output,
		DurationMs: exec.DurationMs,
		CreatedAt:  exec.CreatedAt,
	}, nil
}

func (a *taskRunnerAdapter) EvaluateExecution(ctx context.Context, taskExecID, metricID string) (service.EvaluationResult, error) {
	eval, err := a.evalRunner.EvaluateExecution(ctx, taskExecID, metricID)
	if err != nil {
		return service.EvaluationResult{}, err
	}
	return service.EvaluationResult{
		ID:              eval.ID,
		TaskExecutionID: eval.TaskExecutionID,
		MetricID:        eval.MetricID,
		Score:           eval.Score,
		Details:         eval.Details,
		EvaluatedAt:     eval.EvaluatedAt,
	}, nil
}

func createExperimentService(repo server.Repository, runner service.TaskRunner) *service.ExperimentService {
	return service.NewExperimentService(
		repoToExperimentRepoAdapter{repo: repo},
		repoToExperimentRunRepoAdapter{repo: repo},
		repoToTaskRepoAdapter{repo: repo},
		repoToMetricRepoAdapter{repo: repo},
		runner,
	)
}

type repoToTaskRepoAdapter struct {
	repo server.Repository
}

func (a repoToTaskRepoAdapter) CreateTask(svcTask service.Task) (service.Task, error) {
	task, err := a.repo.CreateTask(toServerTask(svcTask))
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
	return a.repo.UpdateTask(toServerTask(svcTask))
}

func (a repoToTaskRepoAdapter) DeleteTask(id string) error {
	return a.repo.DeleteTask(id)
}

type repoToMetricRepoAdapter struct {
	repo server.Repository
}

func (a repoToMetricRepoAdapter) CreateMetric(metric service.Metric) (service.Metric, error) {
	serverMetric := server.Metric{
		ID:        metric.ID,
		Name:      metric.Name,
		Type:      metric.Type,
		Config:    metric.Config,
		CreatedAt: metric.CreatedAt,
	}
	created, err := a.repo.CreateMetric(serverMetric)
	if err != nil {
		return service.Metric{}, err
	}
	return toServiceMetric(created), nil
}

func (a repoToMetricRepoAdapter) GetMetric(id string) (service.Metric, bool) {
	metric, ok := a.repo.GetMetric(id)
	if !ok {
		return service.Metric{}, false
	}
	return toServiceMetric(metric), true
}

func (a repoToMetricRepoAdapter) ListMetrics() []service.Metric {
	metrics := a.repo.ListMetrics()
	result := make([]service.Metric, len(metrics))
	for i, m := range metrics {
		result[i] = toServiceMetric(m)
	}
	return result
}

type repoToTaskExecutionRepoAdapter struct {
	repo server.Repository
}

func (a repoToTaskExecutionRepoAdapter) CreateTaskExecution(ex evaluator.TaskExecution) (evaluator.TaskExecution, error) {
	serverExec := server.TaskExecution{
		ID:         ex.ID,
		TaskID:     ex.TaskID,
		AgentID:    ex.AgentID,
		Status:     ex.Status,
		Input:      ex.Input,
		Output:     ex.Output,
		DurationMs: ex.DurationMs,
		CreatedAt:  ex.CreatedAt,
	}
	created, err := a.repo.CreateTaskExecution(serverExec)
	if err != nil {
		return evaluator.TaskExecution{}, err
	}
	return toEvaluatorTaskExecution(created), nil
}

func (a repoToTaskExecutionRepoAdapter) GetTaskExecution(id string) (evaluator.TaskExecution, bool) {
	ex, ok := a.repo.GetTaskExecution(id)
	if !ok {
		return evaluator.TaskExecution{}, false
	}
	return toEvaluatorTaskExecution(ex), true
}

func (a repoToTaskExecutionRepoAdapter) UpdateTaskExecution(ex evaluator.TaskExecution) error {
	return a.repo.UpdateTaskExecution(toServerTaskExecution(ex))
}

type repoToEvaluationRepoAdapter struct {
	repo server.Repository
}

func (a repoToEvaluationRepoAdapter) CreateEvaluation(eval evaluator.EvaluationResult) error {
	serverEval := server.Evaluation{
		ID:              eval.ID,
		TaskExecutionID: eval.TaskExecutionID,
		MetricID:        eval.MetricID,
		Score:           eval.Score,
		Details:         eval.Details,
		EvaluatedAt:     eval.EvaluatedAt,
	}
	_, err := a.repo.CreateEvaluationResult(serverEval)
	return err
}

func (a repoToEvaluationRepoAdapter) GetEvaluation(id string) (evaluator.EvaluationResult, bool) {
	return evaluator.EvaluationResult{}, false
}

func (a repoToEvaluationRepoAdapter) ListEvaluations(taskExecutionID string) []evaluator.EvaluationResult {
	evals := a.repo.ListEvaluations(taskExecutionID)
	result := make([]evaluator.EvaluationResult, len(evals))
	for i, e := range evals {
		result[i] = toEvaluatorEvaluationResult(e)
	}
	return result
}

type repoToExperimentRepoAdapter struct {
	repo server.Repository
}

func (a repoToExperimentRepoAdapter) CreateExperiment(exp service.Experiment) (service.Experiment, error) {
	serverVariants := make([]server.Variant, len(exp.Variants))
	for i, v := range exp.Variants {
		serverVariants[i] = server.Variant{
			Name:        v.Name,
			Model:       v.Model,
			Temperature: v.Temperature,
			MaxTokens:   v.MaxTokens,
			Seed:        v.Seed,
			SkillConfig: v.SkillConfig,
		}
	}
	serverExp := server.Experiment{
		ID:          exp.ID,
		Name:        exp.Name,
		Description: exp.Description,
		TaskIDs:     exp.TaskIDs,
		AgentIDs:    exp.AgentIDs,
		Variants:    serverVariants,
		Status:      exp.Status,
		CreatedAt:   exp.CreatedAt,
		UpdatedAt:   exp.UpdatedAt,
	}
	created, err := a.repo.CreateExperiment(serverExp)
	if err != nil {
		return service.Experiment{}, err
	}
	return toServiceExperiment(created), nil
}

func (a repoToExperimentRepoAdapter) GetExperiment(id string) (service.Experiment, bool) {
	exp, ok := a.repo.GetExperiment(id)
	if !ok {
		return service.Experiment{}, false
	}
	return toServiceExperiment(exp), true
}

func (a repoToExperimentRepoAdapter) ListExperiments() []service.Experiment {
	experiments := a.repo.ListExperiments()
	result := make([]service.Experiment, len(experiments))
	for i, e := range experiments {
		result[i] = toServiceExperiment(e)
	}
	return result
}

func (a repoToExperimentRepoAdapter) UpdateExperimentStatus(id string, status string) error {
	exp, ok := a.repo.GetExperiment(id)
	if !ok {
		return errors.New("experiment not found")
	}
	exp.Status = status
	exp.UpdatedAt = time.Now()
	_, err := a.repo.CreateExperiment(exp)
	return err
}

type repoToExperimentRunRepoAdapter struct {
	repo server.Repository
}

func (a repoToExperimentRunRepoAdapter) CreateExperimentRun(run service.ExperimentRun) (service.ExperimentRun, error) {
	serverRun := server.ExperimentRun{
		ID:           run.ID,
		ExperimentID: run.ExperimentID,
		TaskID:       run.TaskID,
		AgentID:      run.AgentID,
		VariantID:    run.VariantID,
		MetricScores: run.MetricScores,
		OverallScore: run.OverallScore,
		DurationMs:   run.DurationMs,
		Status:       run.Status,
		CreatedAt:    run.CreatedAt,
	}
	created, err := a.repo.CreateExperimentRun(serverRun)
	if err != nil {
		return service.ExperimentRun{}, err
	}
	return service.ExperimentRun{
		ID:           created.ID,
		ExperimentID: created.ExperimentID,
		TaskID:       created.TaskID,
		AgentID:      created.AgentID,
		VariantID:    created.VariantID,
		MetricScores: created.MetricScores,
		OverallScore: created.OverallScore,
		DurationMs:   created.DurationMs,
		Status:       created.Status,
		CreatedAt:    created.CreatedAt,
	}, nil
}

func (a repoToExperimentRunRepoAdapter) GetExperimentRun(id string) (service.ExperimentRun, bool) {
	return service.ExperimentRun{}, false
}

func (a repoToExperimentRunRepoAdapter) ListExperimentRuns(experimentID string) []service.ExperimentRun {
	runs := a.repo.ListExperimentRuns(experimentID)
	result := make([]service.ExperimentRun, len(runs))
	for i, r := range runs {
		result[i] = toServiceExperimentRun(r)
	}
	return result
}

func (a repoToExperimentRunRepoAdapter) UpdateExperimentRun(run service.ExperimentRun) error {
	serverRun := server.ExperimentRun{
		ID:           run.ID,
		ExperimentID: run.ExperimentID,
		TaskID:       run.TaskID,
		AgentID:      run.AgentID,
		VariantID:    run.VariantID,
		MetricScores: run.MetricScores,
		OverallScore: run.OverallScore,
		DurationMs:   run.DurationMs,
		Status:       run.Status,
		CreatedAt:    run.CreatedAt,
	}
	return a.repo.UpdateExperimentRun(serverRun)
}

func toServerTask(svcTask service.Task) server.Task {
	testCases := make([]server.TestCase, len(svcTask.TestCases))
	for i, tc := range svcTask.TestCases {
		testCases[i] = server.TestCase{
			Input:    tc.Input,
			Expected: tc.Expected,
		}
	}
	return server.Task{
		ID:          svcTask.ID,
		Name:        svcTask.Name,
		Description: svcTask.Description,
		Tags:        svcTask.Tags,
		CreatedAt:   svcTask.CreatedAt,
		UpdatedAt:   svcTask.UpdatedAt,
		TaskType:    svcTask.TaskType,
		Input: server.TaskInput{
			Source: svcTask.Input.Source,
			Path:   svcTask.Input.Path,
			Format: svcTask.Input.Format,
		},
		Gold: server.GoldConfig{
			Type: svcTask.Gold.Type,
			Data: svcTask.Gold.Data,
		},
		Scoring: server.ScoringConfig{
			PrimaryMetric:    svcTask.Scoring.PrimaryMetric,
			SecondaryMetrics: svcTask.Scoring.SecondaryMetrics,
			Threshold: server.Threshold{
				Pass:            svcTask.Scoring.Threshold.Pass,
				RegressionAlert: svcTask.Scoring.Threshold.RegressionAlert,
			},
		},
		Execution: server.ExecutionConfig{
			Model:       svcTask.Execution.Model,
			Temperature: svcTask.Execution.Temperature,
			MaxTokens:   svcTask.Execution.MaxTokens,
			Seed:        svcTask.Execution.Seed,
		},
		SkillID:    svcTask.SkillID,
		Difficulty: svcTask.Difficulty,
		TestCases:  testCases,
	}
}

func toServiceTask(t server.Task) service.Task {
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

func toServiceMetric(m server.Metric) service.Metric {
	return service.Metric{
		ID:        m.ID,
		Name:      m.Name,
		Type:      m.Type,
		Config:    m.Config,
		CreatedAt: m.CreatedAt,
	}
}

func toEvaluatorTaskExecution(ex server.TaskExecution) evaluator.TaskExecution {
	return evaluator.TaskExecution{
		ID:         ex.ID,
		TaskID:     ex.TaskID,
		AgentID:    ex.AgentID,
		Status:     ex.Status,
		Input:      ex.Input,
		Output:     ex.Output,
		DurationMs: ex.DurationMs,
		CreatedAt:  ex.CreatedAt,
	}
}

func toServerTaskExecution(ex evaluator.TaskExecution) server.TaskExecution {
	return server.TaskExecution{
		ID:         ex.ID,
		TaskID:     ex.TaskID,
		AgentID:    ex.AgentID,
		Status:     ex.Status,
		Input:      ex.Input,
		Output:     ex.Output,
		DurationMs: ex.DurationMs,
		CreatedAt:  ex.CreatedAt,
	}
}

func toEvaluatorEvaluationResult(e server.Evaluation) evaluator.EvaluationResult {
	return evaluator.EvaluationResult{
		ID:              e.ID,
		TaskExecutionID: e.TaskExecutionID,
		MetricID:        e.MetricID,
		Score:           e.Score,
		Details:         e.Details,
		EvaluatedAt:     e.EvaluatedAt,
	}
}

func toServiceExperiment(e server.Experiment) service.Experiment {
	serviceVariants := make([]service.Variant, len(e.Variants))
	for i, v := range e.Variants {
		serviceVariants[i] = service.Variant{
			Name:        v.Name,
			Model:       v.Model,
			Temperature: v.Temperature,
			MaxTokens:   v.MaxTokens,
			Seed:        v.Seed,
			SkillConfig: v.SkillConfig,
		}
	}
	return service.Experiment{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		TaskIDs:     e.TaskIDs,
		AgentIDs:    e.AgentIDs,
		Variants:    serviceVariants,
		Status:      e.Status,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func toServiceExperimentRun(r server.ExperimentRun) service.ExperimentRun {
	return service.ExperimentRun{
		ID:           r.ID,
		ExperimentID: r.ExperimentID,
		TaskID:       r.TaskID,
		AgentID:      r.AgentID,
		VariantID:    r.VariantID,
		MetricScores: r.MetricScores,
		OverallScore: r.OverallScore,
		DurationMs:   r.DurationMs,
		Status:       r.Status,
		CreatedAt:    r.CreatedAt,
	}
}
