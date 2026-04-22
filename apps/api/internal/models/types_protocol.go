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

type MemoryScope string

const (
	MemoryScopeSession   MemoryScope = "session"
	MemoryScopeExecution MemoryScope = "execution"
	MemoryScopeAgent     MemoryScope = "agent"
	MemoryScopeTenant    MemoryScope = "tenant"
)

type MemoryValueType string

const (
	MemoryValueTypeText   MemoryValueType = "text"
	MemoryValueTypeJSON   MemoryValueType = "json"
	MemoryValueTypeBinary MemoryValueType = "binary"
)

type MemoryCell struct {
	ID          string                 `json:"id"`
	Scope       string                 `json:"scope"`
	AgentID     string                 `json:"agent_id"`
	SessionID   string                 `json:"session_id,omitempty"`
	ExecutionID string                 `json:"execution_id,omitempty"`
	Key         string                 `json:"key"`
	Value       map[string]interface{} `json:"value"`
	ValueType   string                 `json:"value_type"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ExpiresAt   *time.Time            `json:"expires_at,omitempty"`
}
