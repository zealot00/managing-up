# Managing-Up v2.0 Phase 0 实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 完成治理内核收敛的第一阶段，创建统一会话模型、策略钩子和能力快照

**Architecture:** 在现有 MCP Router 和 Execution Engine 基础上，增加 Gateway Session 层和 Pre-flight Policy Hook，形成"请求归一化 -> 风险/合规判定 -> 路由决策 -> Trace session 建立 -> 执行后评测回填"的闭环

**Tech Stack:** Go (PostgreSQL, JSONB, UUID)

---

## Task 1: 创建 mcp_gateway_sessions 数据模型

**Files:**
- Create: `apps/api/migrations/0017_add_gateway_sessions.up.sql`
- Modify: `apps/api/internal/server/types.go` (add GatewaySession type)
- Modify: `apps/api/internal/repository/postgres/repository.go` (add Session CRUD)

**Step 1: Write migration**

```sql
-- Gateway Sessions: 会话头，关联所有路由和执行事件
CREATE TABLE mcp_gateway_sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_type        VARCHAR(50) NOT NULL DEFAULT 'router',
    agent_id            VARCHAR(255) NOT NULL,
    correlation_id      VARCHAR(255) NOT NULL,
    task_intent         JSONB NOT NULL DEFAULT '{}',
    risk_level          VARCHAR(50) DEFAULT 'low',
    policy_decision     JSONB,
    status              VARCHAR(50) DEFAULT 'active',
    started_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at            TIMESTAMPTZ,
    metadata_           JSONB DEFAULT '{}'
);

CREATE INDEX idx_gateway_sessions_correlation ON mcp_gateway_sessions(correlation_id);
CREATE INDEX idx_gateway_sessions_agent ON mcp_gateway_sessions(agent_id);
CREATE INDEX idx_gateway_sessions_status ON mcp_gateway_sessions(status);
```

**Step 2: 创建 Go 类型**

在 `apps/api/internal/server/types.go` 添加:

```go
type GatewaySession struct {
    ID            string                 `json:"id"`
    SessionType   string                 `json:"session_type"`
    AgentID       string                 `json:"agent_id"`
    CorrelationID string                 `json:"correlation_id"`
    TaskIntent    map[string]interface{} `json:"task_intent"`
    RiskLevel     string                 `json:"risk_level"`
    PolicyDecision map[string]interface{} `json:"policy_decision"`
    Status        string                 `json:"status"`
    StartedAt     time.Time              `json:"started_at"`
    EndedAt       *time.Time             `json:"ended_at,omitempty"`
    Metadata      map[string]interface{} `json:"metadata,omitempty"`
}
```

**Step 3: 添加 Repository 方法**

在 `apps/api/internal/repository/postgres/repository.go` 添加:

```go
func (r *Repository) CreateGatewaySession(ctx context.Context, session *server.GatewaySession) error
func (r *Repository) GetGatewaySession(ctx context.Context, id string) (*server.GatewaySession, error)
func (r *Repository) UpdateGatewaySessionPolicyDecision(ctx context.Context, id string, decision map[string]interface{}) error
func (r *Repository) EndGatewaySession(ctx context.Context, id string) error
func (r *Repository) ListGatewaySessions(ctx context.Context, agentID string, limit int) ([]server.GatewaySession, error)
```

---

## Task 2: 创建 skill_capability_snapshots 数据模型

**Files:**
- Create: `apps/api/migrations/0018_add_capability_snapshots.up.sql`
- Modify: `apps/api/internal/server/types.go` (add SkillCapabilitySnapshot type)
- Modify: `apps/api/internal/repository/postgres/repository.go`

**Step 1: Write migration**

```sql
-- Skill Capability Snapshots: 记录技能版本的能力评测快照
CREATE TABLE skill_capability_snapshots (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id            TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    version             VARCHAR(50) NOT NULL,
    snapshot_type       VARCHAR(50) NOT NULL DEFAULT 'regression_gate',
    dataset_id          TEXT,
    run_id              TEXT,
    metrics             JSONB NOT NULL DEFAULT '{}',
    overall_score       DECIMAL(5,2),
    passed              BOOLEAN NOT NULL DEFAULT false,
    gate_policy_id      TEXT,
    evaluated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_snapshots_skill ON skill_capability_snapshots(skill_id, version DESC);
CREATE INDEX idx_snapshots_passed ON skill_capability_snapshots(passed);
CREATE INDEX idx_snapshots_evaluated ON skill_capability_snapshots(evaluated_at DESC);
```

**Step 2: 创建 Go 类型**

```go
type SkillCapabilitySnapshot struct {
    ID            string                  `json:"id"`
    SkillID      string                  `json:"skill_id"`
    Version      string                  `json:"version"`
    SnapshotType string                  `json:"snapshot_type"`
    DatasetID    string                  `json:"dataset_id,omitempty"`
    RunID        string                  `json:"run_id,omitempty"`
    Metrics      map[string]float64      `json:"metrics"`
    OverallScore float64                 `json:"overall_score"`
    Passed       bool                    `json:"passed"`
    GatePolicyID string                  `json:"gate_policy_id,omitempty"`
    EvaluatedAt  time.Time               `json:"evaluated_at"`
    CreatedAt    time.Time               `json:"created_at"`
}
```

**Step 3: 添加 Repository 方法**

```go
func (r *Repository) CreateCapabilitySnapshot(ctx context.Context, snap *server.SkillCapabilitySnapshot) error
func (r *Repository) GetLatestSnapshot(ctx context.Context, skillID, version string) (*server.SkillCapabilitySnapshot, error)
func (r *Repository) ListSnapshots(ctx context.Context, skillID string, limit int) ([]server.SkillCapabilitySnapshot, error)
func (r *Repository) GetLatestPassedSnapshot(ctx context.Context, skillID, version string) (*server.SkillCapabilitySnapshot, error)
```

---

## Task 3: 为 mcp_router_logs 增加 session_id

**Files:**
- Modify: `apps/api/migrations/0019_add_session_id_to_router_logs.up.sql`
- Modify: `apps/api/internal/server/types_mcp_router.go`
- Modify: `apps/api/internal/repository/postgres/repository.go`

**Step 1: Write migration**

```sql
ALTER TABLE mcp_router_logs ADD COLUMN session_id UUID REFERENCES mcp_gateway_sessions(id) ON DELETE SET NULL;
CREATE INDEX idx_mcp_router_logs_session ON mcp_router_logs(session_id);
```

**Step 2: 更新 Go 类型**

```go
type MCPRouterLog struct {
    // ... existing fields ...
    SessionID *string `json:"session_id,omitempty"`
}
```

---

## Task 4: 定义统一内部协议

**Files:**
- Create: `apps/api/internal/engine/protocol.go`

**Step 1: 创建协议文件**

```go
package engine

type GatewaySession struct {
    ID            string
    AgentID       string
    CorrelationID string
    TaskIntent    TaskIntent
    RiskLevel     RiskLevel
    PolicyDecision *PolicyDecision
    Status        SessionStatus
    StartedAt     time.Time
}

type TaskIntent struct {
    TaskType string
    Tags     []string
    RawDescription string
    Complexity  string
    Metadata    map[string]interface{}
}

type RiskLevel string

const (
    RiskLevelLow     RiskLevel = "low"
    RiskLevelMedium  RiskLevel = "medium"
    RiskLevelHigh    RiskLevel = "high"
    RiskLevelCritical RiskLevel = "critical"
)

type PolicyDecision struct {
    Allowed       bool
    RequiredApprovals []string
    PolicyID      string
    PolicyVersion string
    Reasons       []string
    DeterminedAt  time.Time
}

type SessionStatus string

const (
    SessionStatusActive    SessionStatus = "active"
    SessionStatusCompleted SessionStatus = "completed"
    SessionStatusCancelled SessionStatus = "cancelled"
)
```

---

## Task 5: 增加 Pre-flight Policy Hook 到 MCPRouterService

**Files:**
- Modify: `apps/api/internal/service/mcp_router_service.go`
- Modify: `apps/api/internal/server/handlers/mcp_router.go`

**Step 1: 扩展 MCPRouterService**

```go
type PolicyChecker interface {
    CheckPolicy(ctx context.Context, intent TaskIntent) (*PolicyDecision, error)
}

type MCPRouterService struct {
    repo             MCPRouterRepository
    metricsCollector *MetricsCollector
    policyChecker    PolicyChecker
}

func (s *MCPRouterService) MatchTaskWithPolicy(ctx context.Context, intent TaskIntent) (*MatchResult, *PolicyDecision, error) {
    decision, err := s.policyChecker.CheckPolicy(ctx, intent)
    if err != nil {
        return nil, nil, fmt.Errorf("policy check failed: %w", err)
    }

    if !decision.Allowed {
        return &MatchResult{Matched: false}, decision, nil
    }

    result, err := s.repo.FindMatchingServers(ctx, []string{intent.TaskType}, intent.Tags)
    // ... rest of matching logic
}
```

**Step 2: 基础 PolicyChecker 实现**

创建 `apps/api/internal/service/policy_checker.go`:

```go
type DefaultPolicyChecker struct {
    repo MCPRouterRepository
}

func (c *DefaultPolicyChecker) CheckPolicy(ctx context.Context, intent TaskIntent) (*PolicyDecision, error) {
    decision := &PolicyDecision{
        Allowed:     true,
        DeterminedAt: time.Now(),
    }

    // High risk task types require approval
    highRiskTypes := map[string]bool{
        "delete":     true,
        "deploy":     true,
        "payment":    true,
        "user_data":  true,
    }

    if highRiskTypes[intent.TaskType] {
        decision.Allowed = false
        decision.RequiredApprovals = []string{"risk_approval"}
        decision.Reasons = []string{fmt.Sprintf("Task type '%s' requires risk approval", intent.TaskType)}
    }

    return decision, nil
}
```

---

## Task 6: 创建 GatewaySessionService

**Files:**
- Create: `apps/api/internal/service/gateway_session_service.go`

**Step 1: 创建 Service**

```go
type GatewaySessionService struct {
    repo          Repository
    routerService *MCPRouterService
}

func NewGatewaySessionService(repo Repository, routerSvc *MCPRouterService) *GatewaySessionService {
    return &GatewaySessionService{repo: repo, routerService: routerSvc}
}

func (s *GatewaySessionService) CreateSession(ctx context.Context, agentID, correlationID string, intent engine.TaskIntent) (*engine.GatewaySession, error) {
    riskLevel := s.assessRiskLevel(intent)

    session := &engine.GatewaySession{
        ID:            uuid.New().String(),
        AgentID:       agentID,
        CorrelationID: correlationID,
        TaskIntent:    intent,
        RiskLevel:     riskLevel,
        Status:        engine.SessionStatusActive,
        StartedAt:     time.Now(),
    }

    if err := s.repo.CreateGatewaySession(ctx, session); err != nil {
        return nil, err
    }

    return session, nil
}

func (s *GatewaySessionService) assessRiskLevel(intent engine.TaskIntent) engine.RiskLevel {
    highRiskKeywords := []string{"delete", "deploy", "payment", "admin", "user_data"}
    for _, keyword := range highRiskKeywords {
        if strings.Contains(strings.ToLower(intent.TaskType), keyword) {
            return engine.RiskLevelHigh
        }
    }
    return engine.RiskLevelLow
}
```

---

## Task 7: 集成 Regression Gate 到 Skill Promote

**Files:**
- Modify: `apps/api/internal/service/skill_service.go` 或相关 promote handler
- Modify: `apps/seh/handlers.go` (SEH 已有的 gate 逻辑)

**Step 1: 添加 Promote 前检查**

在 skill promote API 或 handler 中添加:

```go
func (s *SkillService) PromoteSkillVersion(ctx context.Context, skillID, version string) error {
    // 1. Get latest capability snapshot
    snapshot, err := s.repo.GetLatestPassedSnapshot(ctx, skillID, version)
    if err != nil {
        return fmt.Errorf("failed to get snapshot: %w", err)
    }

    // 2. Check if snapshot exists and passed
    if snapshot == nil {
        return ErrNoPassedSnapshot
    }

    if !snapshot.Passed {
        return fmt.Errorf("skill version %s did not pass regression gate (score: %.2f)", version, snapshot.OverallScore)
    }

    // 3. Proceed with promote
    return s.repo.PromoteSkillVersion(ctx, skillID, version)
}
```

---

## Task 8: 更新 MCPRouterHandler 使用 Session

**Files:**
- Modify: `apps/api/internal/server/handlers/mcp_router.go`

**Step 1: 修改 Route 方法**

```go
func (h *MCPRouterHandler) Route(w http.ResponseWriter, r *http.Request) {
    var req RouteRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    // 1. Create or get session
    intent := engine.TaskIntent{
        TaskType:       req.Task.Structured.TaskType,
        Tags:          req.Task.Structured.Tags,
        RawDescription: req.Task.Structured.RawDescription,
    }

    session, err := h.sessionSvc.CreateSession(ctx, req.AgentID, req.CorrelationID, intent)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create session")
        return
    }

    // 2. Policy check
    decision, err := h.routerSvc.CheckPolicy(ctx, intent)
    if err != nil || !decision.Allowed {
        h.sessionSvc.RecordPolicyDecision(ctx, session.ID, decision)
        writeError(w, http.StatusForbidden, "POLICY_DENIED", "task not allowed by policy")
        return
    }

    // 3. Route with policy context
    result, err := h.routerSvc.MatchTask(ctx, []string{req.Task.Structured.TaskType}, req.Task.Structured.Tags)
    // ... rest of routing logic

    // 4. Record route in session
    h.routerSvc.LogRouteWithSession(ctx, result, session.ID)
}
```

---

## 执行选项

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**