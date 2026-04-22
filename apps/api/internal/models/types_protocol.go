package models

import "time"

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

type PolicyRule struct {
	ID        string   `json:"id"`
	Condition string   `json:"condition"`
	Action    string   `json:"action"`
	Reason    string   `json:"reason"`
}

type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusCancelled SessionStatus = "cancelled"
)
