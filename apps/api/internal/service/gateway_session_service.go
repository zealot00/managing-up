package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zealot/managing-up/apps/api/internal/models"
)

type GatewaySession struct {
	ID             string
	AgentID        string
	CorrelationID  string
	TaskIntent     map[string]interface{}
	RiskLevel      string
	PolicyDecision map[string]interface{}
	Status         string
	StartedAt      time.Time
	EndedAt        *time.Time
	Metadata       map[string]interface{}
}

type GatewaySessionRepository interface {
	CreateGatewaySession(ctx context.Context, session *GatewaySession) error
}

type GatewaySessionService struct {
	repo          GatewaySessionRepository
	routerService *MCPRouterService
}

func NewGatewaySessionService(repo GatewaySessionRepository, routerSvc *MCPRouterService) *GatewaySessionService {
	return &GatewaySessionService{repo: repo, routerService: routerSvc}
}

func (s *GatewaySessionService) CreateSession(ctx context.Context, agentID, correlationID string, intent models.TaskIntent) (*GatewaySession, error) {
	riskLevel := s.assessRiskLevel(intent)

	session := &GatewaySession{
		ID:            uuid.New().String(),
		AgentID:       agentID,
		CorrelationID: correlationID,
		RiskLevel:     string(riskLevel),
		Status:        string(models.SessionStatusActive),
		StartedAt:     time.Now(),
	}

	taskIntentMap := make(map[string]interface{})
	taskIntentMap["task_type"] = intent.TaskType
	taskIntentMap["tags"] = intent.Tags
	taskIntentMap["raw_description"] = intent.RawDescription
	taskIntentMap["complexity"] = intent.Complexity
	taskIntentMap["metadata"] = intent.Metadata
	session.TaskIntent = taskIntentMap

	if err := s.repo.CreateGatewaySession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *GatewaySessionService) assessRiskLevel(intent models.TaskIntent) models.RiskLevel {
	highRiskKeywords := []string{"delete", "deploy", "payment", "admin", "user_data"}
	for _, keyword := range highRiskKeywords {
		if strings.Contains(strings.ToLower(intent.TaskType), keyword) {
			return models.RiskLevelHigh
		}
	}
	return models.RiskLevelLow
}