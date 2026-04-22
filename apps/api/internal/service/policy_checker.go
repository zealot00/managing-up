package service

import (
	"context"
	"fmt"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/engine"
)

type DefaultPolicyChecker struct {
	repo MCPRouterRepository
}

func NewDefaultPolicyChecker(repo MCPRouterRepository) *DefaultPolicyChecker {
	return &DefaultPolicyChecker{repo: repo}
}

func (c *DefaultPolicyChecker) CheckPolicy(ctx context.Context, intent engine.TaskIntent) (*engine.PolicyDecision, error) {
	decision := &engine.PolicyDecision{
		Allowed:     true,
		DeterminedAt: time.Now(),
	}

	highRiskTypes := map[string]bool{
		"delete":    true,
		"deploy":    true,
		"payment":   true,
		"user_data": true,
	}

	if highRiskTypes[intent.TaskType] {
		decision.Allowed = false
		decision.RequiredApprovals = []string{"risk_approval"}
		decision.Reasons = []string{fmt.Sprintf("Task type '%s' requires risk approval", intent.TaskType)}
	}

	return decision, nil
}
