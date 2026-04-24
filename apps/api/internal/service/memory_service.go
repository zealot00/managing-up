package service

import (
	"context"
	"fmt"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/models"
)

type MemoryRepository interface {
	CreateMemoryCell(ctx context.Context, cell *models.MemoryCell) error
	GetMemoryCell(ctx context.Context, id string) (*models.MemoryCell, error)
	GetMemoryCellsBySession(ctx context.Context, sessionID string, limit int) ([]models.MemoryCell, error)
	GetMemoryCellsByAgent(ctx context.Context, agentID string, limit int) ([]models.MemoryCell, error)
	UpdateMemoryCell(ctx context.Context, cell *models.MemoryCell) error
	DeleteMemoryCell(ctx context.Context, id string) error
}

type MemoryHubService struct {
	repo MemoryRepository
}

func NewMemoryHubService(repo MemoryRepository) *MemoryHubService {
	return &MemoryHubService{repo: repo}
}

func (s *MemoryHubService) StoreMemory(ctx context.Context, scope, tenantID, agentID, sessionID, key string, value interface{}, tags []string) (*models.MemoryCell, error) {
	if tenantID == "" {
		tenantID = "default"
	}
	cell := &models.MemoryCell{
		ID:        generateID(),
		Scope:     scope,
		TenantID:  tenantID,
		AgentID:   agentID,
		SessionID: sessionID,
		Key:       key,
		Value:     map[string]interface{}{"data": value},
		ValueType: "json",
		Tags:      tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateMemoryCell(ctx, cell); err != nil {
		return nil, err
	}

	return cell, nil
}

func (s *MemoryHubService) GetMemory(ctx context.Context, id string) (*models.MemoryCell, error) {
	return s.repo.GetMemoryCell(ctx, id)
}

func (s *MemoryHubService) GetSessionMemory(ctx context.Context, sessionID string, limit int) ([]models.MemoryCell, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetMemoryCellsBySession(ctx, sessionID, limit)
}

func (s *MemoryHubService) GetAgentMemory(ctx context.Context, agentID string, limit int) ([]models.MemoryCell, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetMemoryCellsByAgent(ctx, agentID, limit)
}

func (s *MemoryHubService) UpdateMemory(ctx context.Context, cell *models.MemoryCell) error {
	cell.UpdatedAt = time.Now()
	return s.repo.UpdateMemoryCell(ctx, cell)
}

func (s *MemoryHubService) DeleteMemory(ctx context.Context, id string) error {
	return s.repo.DeleteMemoryCell(ctx, id)
}

func (s *MemoryHubService) BuildMemoryContext(ctx context.Context, sessionID, agentID string) (map[string]interface{}, error) {
	cells, err := s.GetSessionMemory(ctx, sessionID, 20)
	if err != nil {
		return nil, err
	}

	memoryMap := make(map[string]interface{})
	for _, cell := range cells {
		memoryMap[cell.Key] = cell.Value
	}

	return map[string]interface{}{
		"session_id": sessionID,
		"agent_id": agentID,
		"memory": memoryMap,
		"count": len(cells),
	}, nil
}

func generateID() string {
	return fmt.Sprintf("mem_%d", time.Now().UnixNano())
}