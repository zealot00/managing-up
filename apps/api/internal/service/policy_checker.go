package service

import (
	"context"
	"strings"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/models"
)

var DefaultPolicyRules = []models.PolicyRule{
	{ID: "high_risk_types", Version: "v1", Condition: "task_type in [delete, deploy, payment, user_data]", Action: "require_approval", Reason: "High-risk task type", Priority: 10, IsActive: true},
	{ID: "high_risk_level", Version: "v1", Condition: "risk_level == high", Action: "require_approval", Reason: "High-risk operation", Priority: 10, IsActive: true},
	{ID: "compliance_required", Version: "v1", Condition: "compliance_required == true", Action: "require_compliance_check", Reason: "Compliance check required for this skill", Priority: 20, IsActive: true},
}

type PolicyRepository interface {
	ListPolicyVersions() ([]models.PolicyVersion, error)
	GetPolicyVersion(name, version string) (*models.PolicyVersion, error)
	GetDefaultPolicyVersion() (*models.PolicyVersion, error)
	CreatePolicyVersion(pv *models.PolicyVersion) error
	UpdatePolicyVersion(pv *models.PolicyVersion) error
}

type DefaultPolicyChecker struct {
	repo  MCPRouterRepository
	rules []models.PolicyRule
}

func NewDefaultPolicyChecker(repo MCPRouterRepository, rules []models.PolicyRule) *DefaultPolicyChecker {
	if rules == nil {
		rules = DefaultPolicyRules
	}
	return &DefaultPolicyChecker{repo: repo, rules: rules}
}

func (c *DefaultPolicyChecker) CheckPolicy(ctx context.Context, intent models.TaskIntent) (*models.PolicyDecision, error) {
	decision := c.evaluateRules(intent)
	return decision, nil
}

func (c *DefaultPolicyChecker) evaluateRules(intent models.TaskIntent) *models.PolicyDecision {
	for _, rule := range c.rules {
		if !rule.IsActive {
			continue
		}
		if c.conditionMatches(rule.Condition, intent) {
			decision := &models.PolicyDecision{
				Allowed:          rule.Action == "allow",
				RequiredApprovals: []string{},
				Reasons:          []string{rule.Reason},
				DeterminedAt:     time.Now(),
				PolicyID:         rule.ID,
				PolicyVersion:    rule.Version,
			}

			switch rule.Action {
			case "require_approval":
				decision.RequiredApprovals = []string{"risk_approval"}
			case "require_compliance_check":
				decision.ComplianceRequired = true
				decision.RequiredApprovals = []string{"compliance_check"}
				if intent.Metadata != nil {
					if sop, ok := intent.Metadata["sop_reference"].(models.SOPReference); ok {
						decision.SOPReference = sop
					}
				}
			}

			return decision
		}
	}
	return &models.PolicyDecision{Allowed: true, DeterminedAt: time.Now()}
}

func (c *DefaultPolicyChecker) conditionMatches(condition string, intent models.TaskIntent) bool {
	condition = strings.TrimSpace(condition)

	if strings.HasPrefix(condition, "task_type") {
		return c.matchTaskType(condition, intent.TaskType)
	}

	if strings.HasPrefix(condition, "risk_level") {
		return c.matchRiskLevel(condition, intent)
	}

	if strings.HasPrefix(condition, "compliance_required") {
		return c.matchComplianceRequired(condition, intent)
	}

	if strings.Contains(condition, "contains") {
		parts := strings.Split(condition, "contains")
		if len(parts) == 2 {
			field := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if field == "task_type" {
				return strings.Contains(intent.TaskType, value)
			}
		}
	}

	return false
}

func (c *DefaultPolicyChecker) matchTaskType(condition, taskType string) bool {
	if strings.Contains(condition, " in [") {
		parts := strings.Split(condition, " in [")
		if len(parts) == 2 {
			valuesStr := strings.TrimSuffix(parts[1], "]")
			values := strings.Split(valuesStr, ", ")
			for _, v := range values {
				if strings.TrimSpace(v) == taskType {
					return true
				}
			}
		}
	}

	if strings.Contains(condition, "==") {
		parts := strings.Split(condition, "==")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1]) == taskType
		}
	}

	return false
}

func (c *DefaultPolicyChecker) matchRiskLevel(condition string, intent models.TaskIntent) bool {
	if len(intent.Tags) == 0 {
		return false
	}

	var riskLevel string
	for _, tag := range intent.Tags {
		if strings.HasPrefix(tag, "risk_level:") {
			riskLevel = strings.TrimPrefix(tag, "risk_level:")
			break
		}
	}

	if riskLevel == "" {
		return false
	}

	if strings.Contains(condition, "==") {
		parts := strings.Split(condition, "==")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1]) == riskLevel
		}
	}

	return false
}

func (c *DefaultPolicyChecker) matchComplianceRequired(condition string, intent models.TaskIntent) bool {
	if intent.Metadata == nil {
		return false
	}

	if val, ok := intent.Metadata["compliance_required"]; ok {
		if b, ok := val.(bool); ok && b {
			return true
		}
	}

	return false
}