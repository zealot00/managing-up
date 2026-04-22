package engine

import "time"

type GatewaySession struct {
	ID             string
	AgentID        string
	CorrelationID  string
	TaskIntent     TaskIntent
	RiskLevel      RiskLevel
	PolicyDecision *PolicyDecision
	Status         SessionStatus
	StartedAt      time.Time
}

type TaskIntent struct {
	TaskType       string
	Tags           []string
	RawDescription string
	Complexity     string
	Metadata       map[string]interface{}
}

type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

type PolicyDecision struct {
	Allowed          bool
	RequiredApprovals []string
	PolicyID         string
	PolicyVersion    string
	Reasons          []string
	DeterminedAt     time.Time
}

type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusCancelled SessionStatus = "cancelled"
)
