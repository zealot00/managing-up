package server

import (
	"github.com/zealot/managing-up/apps/api/internal/server/handlers"
	"github.com/zealot/managing-up/apps/api/internal/service"
)

type repoToMCPInvokeRepoAdapter struct {
	repo Repository
}

func (a *repoToMCPInvokeRepoAdapter) CheckMCPPermission(mcpServerID, userID, apiKeyID, skillID string) (bool, error) {
	return a.repo.CheckMCPPermission(mcpServerID, userID, apiKeyID, skillID)
}

func (a *repoToMCPInvokeRepoAdapter) IncrementMCPRouterCatalogUseCount(serverID string) error {
	return a.repo.IncrementMCPRouterCatalogUseCount(serverID)
}

func (a *repoToMCPInvokeRepoAdapter) GetMCPServer(id string) (handlers.MCPServerDTO, bool) {
	server, ok := a.repo.GetMCPServer(id)
	if !ok {
		return handlers.MCPServerDTO{}, false
	}
	return toMCPServerDTO(server), true
}

type repoToMCPGrantRepoAdapter struct {
	repo Repository
}

func (a repoToMCPGrantRepoAdapter) CreateMCPServerPermission(p handlers.MCPServerPermission) (handlers.MCPServerPermission, error) {
	serverPerm := toServerMCPServerPermission(p)
	result, err := a.repo.CreateMCPServerPermission(serverPerm)
	if err != nil {
		return handlers.MCPServerPermission{}, err
	}
	return toHandlersMCPServerPermission(result), nil
}

func (a repoToMCPGrantRepoAdapter) ListMCPServerPermissions(mcpServerID string) ([]handlers.MCPServerPermission, error) {
	perms, err := a.repo.ListMCPServerPermissions(mcpServerID)
	if err != nil {
		return nil, err
	}
	result := make([]handlers.MCPServerPermission, len(perms))
	for i, p := range perms {
		result[i] = toHandlersMCPServerPermission(p)
	}
	return result, nil
}

func (a repoToMCPGrantRepoAdapter) GetMCPServer(id string) (handlers.MCPServerDTO, bool) {
	server, ok := a.repo.GetMCPServer(id)
	if !ok {
		return handlers.MCPServerDTO{}, false
	}
	return toMCPServerDTO(server), true
}

func toServerMCPServerPermission(p handlers.MCPServerPermission) MCPServerPermission {
	return MCPServerPermission{
		ID:              p.ID,
		MCPServerID:     p.MCPServerID,
		UserID:          p.UserID,
		APIKeyID:        p.APIKeyID,
		SkillID:         p.SkillID,
		PermissionType:  p.PermissionType,
		IsGranted:       p.IsGranted,
		GrantedBy:       p.GrantedBy,
		GrantedAt:       p.GrantedAt,
		ExpiresAt:       p.ExpiresAt,
	}
}

func toHandlersMCPServerPermission(p MCPServerPermission) handlers.MCPServerPermission {
	return handlers.MCPServerPermission{
		ID:              p.ID,
		MCPServerID:     p.MCPServerID,
		UserID:          p.UserID,
		APIKeyID:        p.APIKeyID,
		SkillID:         p.SkillID,
		PermissionType:  p.PermissionType,
		IsGranted:       p.IsGranted,
		GrantedBy:       p.GrantedBy,
		GrantedAt:       p.GrantedAt,
		ExpiresAt:       p.ExpiresAt,
	}
}

type repoToSweepRepoAdapter struct {
	repo Repository
}

func (a *repoToSweepRepoAdapter) CreateSweepConfig(cfg handlers.SweepConfig) (handlers.SweepConfig, error) {
	result, err := a.repo.CreateSweepConfig(SweepConfig{
		ID:          cfg.ID,
		Name:        cfg.Name,
		Description: cfg.Description,
		TaskID:      cfg.TaskID,
		Parameters:  toServerSweepParameters(cfg.Parameters),
		Status:      cfg.Status,
		TotalRuns:   cfg.TotalRuns,
		Completed:   cfg.Completed,
		CreatedBy:   cfg.CreatedBy,
		CreatedAt:   cfg.CreatedAt,
		UpdatedAt:   cfg.UpdatedAt,
	})
	if err != nil {
		return handlers.SweepConfig{}, err
	}
	return toHandlersSweepConfig(result), nil
}

func (a *repoToSweepRepoAdapter) GetSweepConfig(id string) (handlers.SweepConfig, bool) {
	cfg, ok := a.repo.GetSweepConfig(id)
	if !ok {
		return handlers.SweepConfig{}, false
	}
	return toHandlersSweepConfig(cfg), true
}

func (a *repoToSweepRepoAdapter) ListSweepConfigs() ([]handlers.SweepConfig, error) {
	configs, err := a.repo.ListSweepConfigs()
	if err != nil {
		return nil, err
	}
	result := make([]handlers.SweepConfig, len(configs))
	for i, cfg := range configs {
		result[i] = toHandlersSweepConfig(cfg)
	}
	return result, nil
}

func (a *repoToSweepRepoAdapter) UpdateSweepConfig(cfg handlers.SweepConfig) error {
	return a.repo.UpdateSweepConfig(SweepConfig{
		ID:          cfg.ID,
		Name:        cfg.Name,
		Description: cfg.Description,
		TaskID:      cfg.TaskID,
		Parameters:  toServerSweepParameters(cfg.Parameters),
		Status:      cfg.Status,
		TotalRuns:   cfg.TotalRuns,
		Completed:   cfg.Completed,
		CreatedBy:   cfg.CreatedBy,
		CreatedAt:   cfg.CreatedAt,
		UpdatedAt:   cfg.UpdatedAt,
	})
}

func (a *repoToSweepRepoAdapter) DeleteSweepConfig(id string) error {
	return a.repo.DeleteSweepConfig(id)
}

func (a *repoToSweepRepoAdapter) CreateSweepRuns(runs []handlers.SweepRun) error {
	sweepRuns := make([]SweepRun, len(runs))
	for i, r := range runs {
		sweepRuns[i] = SweepRun{
			ID:              r.ID,
			SweepConfigID:   r.SweepConfigID,
			VariantIndex:    r.VariantIndex,
			Model:           r.Model,
			Temperature:     r.Temperature,
			MaxTokens:       r.MaxTokens,
			PromptID:        r.PromptID,
			PromptLabel:     r.PromptLabel,
			Status:          r.Status,
			TaskExecutionID: r.TaskExecutionID,
			Score:           r.Score,
			DurationMs:      r.DurationMs,
			Error:           r.Error,
			CreatedAt:       r.CreatedAt,
			CompletedAt:     r.CompletedAt,
		}
	}
	return a.repo.CreateSweepRuns(sweepRuns)
}

func (a *repoToSweepRepoAdapter) GetSweepRunsByConfigID(configID string) ([]handlers.SweepRun, error) {
	runs, err := a.repo.GetSweepRunsByConfigID(configID)
	if err != nil {
		return nil, err
	}
	result := make([]handlers.SweepRun, len(runs))
	for i, r := range runs {
		result[i] = toHandlersSweepRun(r)
	}
	return result, nil
}

func (a *repoToSweepRepoAdapter) UpdateSweepRun(run handlers.SweepRun) error {
	return a.repo.UpdateSweepRun(SweepRun{
		ID:              run.ID,
		SweepConfigID:   run.SweepConfigID,
		VariantIndex:    run.VariantIndex,
		Model:           run.Model,
		Temperature:    run.Temperature,
		MaxTokens:      run.MaxTokens,
		PromptID:       run.PromptID,
		PromptLabel:    run.PromptLabel,
		Status:         run.Status,
		TaskExecutionID: run.TaskExecutionID,
		Score:          run.Score,
		DurationMs:     run.DurationMs,
		Error:          run.Error,
		CreatedAt:      run.CreatedAt,
		CompletedAt:    run.CompletedAt,
	})
}

func (a *repoToSweepRepoAdapter) GetTask(id string) (service.Task, bool) {
	task, ok := a.repo.GetTask(id)
	if !ok {
		return service.Task{}, false
	}
	return service.Task{
		ID:          task.ID,
		Name:        task.Name,
		Description: task.Description,
		Tags:        task.Tags,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		TaskType:    task.TaskType,
		Input:       service.TaskInput{Source: task.Input.Source, Path: task.Input.Path, Format: task.Input.Format},
		Gold:        service.GoldConfig{Type: task.Gold.Type, Data: task.Gold.Data},
		Scoring:     service.ScoringConfig{
			PrimaryMetric:    task.Scoring.PrimaryMetric,
			SecondaryMetrics: task.Scoring.SecondaryMetrics,
			Threshold: service.Threshold{
				Pass:            task.Scoring.Threshold.Pass,
				RegressionAlert: task.Scoring.Threshold.RegressionAlert,
			},
		},
		Execution: service.ExecutionConfig{
			Model:       task.Execution.Model,
			Temperature: task.Execution.Temperature,
			MaxTokens:   task.Execution.MaxTokens,
			Seed:        task.Execution.Seed,
		},
		SkillID:    task.SkillID,
		Difficulty: task.Difficulty,
		TestCases:  convertTestCases(task.TestCases),
	}, true
}

func convertTestCases(cases []TestCase) []service.TestCase {
	result := make([]service.TestCase, len(cases))
	for i, c := range cases {
		result[i] = service.TestCase{Input: c.Input, Expected: c.Expected}
	}
	return result
}

func toHandlersSweepConfig(cfg SweepConfig) handlers.SweepConfig {
	return handlers.SweepConfig{
		ID:          cfg.ID,
		Name:        cfg.Name,
		Description: cfg.Description,
		TaskID:      cfg.TaskID,
		Parameters:  toHandlersSweepParameters(cfg.Parameters),
		Status:      cfg.Status,
		TotalRuns:   cfg.TotalRuns,
		Completed:   cfg.Completed,
		CreatedBy:   cfg.CreatedBy,
		CreatedAt:   cfg.CreatedAt,
		UpdatedAt:   cfg.UpdatedAt,
	}
}

func toHandlersSweepParameters(params SweepParameters) handlers.SweepParameters {
	prompts := make([]handlers.SweepPromptVariant, len(params.Prompts))
	for i, p := range params.Prompts {
		prompts[i] = handlers.SweepPromptVariant{
			ID:      p.ID,
			Label:   p.Label,
			Content: p.Content,
		}
	}
	return handlers.SweepParameters{
		Models:       params.Models,
		Temperatures: params.Temperatures,
		MaxTokens:    params.MaxTokens,
		Prompts:      prompts,
	}
}

func toServerSweepParameters(params handlers.SweepParameters) SweepParameters {
	prompts := make([]SweepPromptVariant, len(params.Prompts))
	for i, p := range params.Prompts {
		prompts[i] = SweepPromptVariant{
			ID:      p.ID,
			Label:   p.Label,
			Content: p.Content,
		}
	}
	return SweepParameters{
		Models:       params.Models,
		Temperatures: params.Temperatures,
		MaxTokens:    params.MaxTokens,
		Prompts:      prompts,
	}
}

func toServerSweepConfig(cfg handlers.SweepConfig) SweepConfig {
	return SweepConfig{
		ID:          cfg.ID,
		Name:        cfg.Name,
		Description: cfg.Description,
		TaskID:      cfg.TaskID,
		Parameters:  toServerSweepParameters(cfg.Parameters),
		Status:      cfg.Status,
		TotalRuns:   cfg.TotalRuns,
		Completed:   cfg.Completed,
		CreatedBy:   cfg.CreatedBy,
		CreatedAt:   cfg.CreatedAt,
		UpdatedAt:   cfg.UpdatedAt,
	}
}

func toHandlersSweepRun(run SweepRun) handlers.SweepRun {
	return handlers.SweepRun{
		ID:              run.ID,
		SweepConfigID:   run.SweepConfigID,
		VariantIndex:    run.VariantIndex,
		Model:           run.Model,
		Temperature:     run.Temperature,
		MaxTokens:       run.MaxTokens,
		PromptID:        run.PromptID,
		PromptLabel:     run.PromptLabel,
		Status:          run.Status,
		TaskExecutionID: run.TaskExecutionID,
		Score:           run.Score,
		DurationMs:      run.DurationMs,
		Error:           run.Error,
		CreatedAt:       run.CreatedAt,
		CompletedAt:     run.CompletedAt,
	}
}