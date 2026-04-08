# Agent Ecosystem Hub Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现 MCP Router 无状态接入层 + Skill Repository 企业级扩展

**Architecture:** 
- MCP Router 是无状态接入层，基于任务类型+标签智能路由
- MCP Server 审批通过后自动同步到路由池
- Skill Repository 重构现有 Registry，支持依赖、评分、SOP 参照
- OpenClaw 兼容接口

**Tech Stack:** Go (Gin), Next.js, PostgreSQL, Prometheus

---

## Phase 1: Database Migrations

### Task 1: Create Migration 0015 - MCP Router Tables

**Files:**
- Create: `apps/api/migrations/0015_add_mcp_router_tables.up.sql`
- Create: `apps/api/migrations/0015_add_mcp_router_tables.down.sql`

**Step 1: Write migration SQL**

```sql
-- MCP Router Catalog
CREATE TABLE mcp_router_catalog (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       UUID NOT NULL UNIQUE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    transport_type  VARCHAR(50) NOT NULL,
    command         VARCHAR(500),
    args            TEXT[],
    url             VARCHAR(500),
    task_types      TEXT[] NOT NULL,
    tags            TEXT[],
    capabilities    JSONB NOT NULL DEFAULT '{}',
    routing_config  JSONB DEFAULT '{}',
    status          VARCHAR(50) DEFAULT 'active',
    trust_score     DECIMAL(3,2) DEFAULT 0.5,
    use_count       BIGINT DEFAULT 0,
    error_count     BIGINT DEFAULT 0,
    last_used_at    TIMESTAMP,
    synced_at       TIMESTAMP DEFAULT NOW(),
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_mcp_catalog_transport CHECK (transport_type IN ('stdio', 'http')),
    CONSTRAINT chk_mcp_catalog_status CHECK (status IN ('active', 'disabled', 'maintenance')),
    CONSTRAINT chk_mcp_catalog_trust CHECK (trust_score >= 0 AND trust_score <= 1)
);

CREATE INDEX idx_mcp_catalog_task_types ON mcp_router_catalog USING GIN(task_types);
CREATE INDEX idx_mcp_catalog_tags ON mcp_router_catalog USING GIN(tags);
CREATE INDEX idx_mcp_catalog_status ON mcp_router_catalog(status);
CREATE INDEX idx_mcp_catalog_trust ON mcp_router_catalog(trust_score DESC);

-- MCP Router Logs
CREATE TABLE mcp_router_logs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    correlation_id      VARCHAR(255) NOT NULL,
    agent_id            VARCHAR(255),
    task_type           VARCHAR(100),
    task_tags           TEXT[],
    task_complexity     VARCHAR(50),
    raw_description     TEXT,
    matched             BOOLEAN NOT NULL,
    matched_server_id   UUID,
    match_score         DECIMAL(3,2),
    match_latency_ms    INTEGER,
    status              VARCHAR(50),
    error_code          VARCHAR(50),
    error_message       TEXT,
    duration_ms         INTEGER,
    created_at          TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_mcp_log_status CHECK (status IN ('success', 'failure', 'timeout'))
);

CREATE INDEX idx_mcp_router_logs_created ON mcp_router_logs(created_at);
CREATE INDEX idx_mcp_router_logs_agent ON mcp_router_logs(agent_id);
CREATE INDEX idx_mcp_router_logs_correlation ON mcp_router_logs(correlation_id);
CREATE INDEX idx_mcp_router_logs_server ON mcp_router_logs(matched_server_id);

-- MCP Router Sync Log
CREATE TABLE mcp_router_sync_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       UUID NOT NULL,
    sync_type       VARCHAR(50) NOT NULL,
    old_status      VARCHAR(50),
    new_status      VARCHAR(50),
    approved_by     UUID,
    approved_at     TIMESTAMP,
    note            TEXT,
    created_at      TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_mcp_sync_type CHECK (sync_type IN ('approved_sync', 'status_change', 'removal', 'catalog_update'))
);

CREATE INDEX idx_mcp_sync_log_server ON mcp_router_sync_log(server_id);
```

**Step 2: Write down migration**

```sql
DROP TABLE IF EXISTS mcp_router_sync_log;
DROP TABLE IF EXISTS mcp_router_logs;
DROP TABLE IF EXISTS mcp_router_catalog;
```

**Step 3: Run migration test**

```bash
cd apps/api && go run cmd/migrate/main.go
```

Expected: Migration runs successfully

---

### Task 2: Create Migration 0016 - Skill Repository Extensions

**Files:**
- Create: `apps/api/migrations/0016_extend_skills_for_enterprise.up.sql`
- Create: `apps/api/migrations/0016_extend_skills_for_enterprise.down.sql`

**Step 1: Write migration SQL**

```sql
-- Extend skills table
ALTER TABLE skills ADD COLUMN IF NOT EXISTS sop_id VARCHAR(100);
ALTER TABLE skills ADD COLUMN IF NOT EXISTS sop_name VARCHAR(255);
ALTER TABLE skills ADD COLUMN IF NOT EXISTS sop_version VARCHAR(50);
ALTER TABLE skills ADD COLUMN IF NOT EXISTS sop_section VARCHAR(255);
ALTER TABLE skills ADD COLUMN IF NOT EXISTS compliance_required BOOLEAN DEFAULT false;
ALTER TABLE skills ADD COLUMN IF NOT EXISTS category VARCHAR(100);
ALTER TABLE skills ADD COLUMN IF NOT EXISTS tags TEXT[];
ALTER TABLE skills ADD COLUMN IF NOT EXISTS trust_score DECIMAL(3,2) DEFAULT 0.5;
ALTER TABLE skills ADD COLUMN IF NOT EXISTS verified BOOLEAN DEFAULT false;
ALTER TABLE skills ADD COLUMN IF NOT EXISTS published_at TIMESTAMP;
ALTER TABLE skills ADD COLUMN IF NOT EXISTS published_by UUID;
ALTER TABLE skills ADD COLUMN IF NOT EXISTS draft_source VARCHAR(50) DEFAULT 'manual';
ALTER TABLE skills ADD COLUMN IF NOT EXISTS draft_source_meta JSONB DEFAULT '{}';
ALTER TABLE skills ADD COLUMN IF NOT EXISTS created_by UUID;

ALTER TABLE skills ADD CONSTRAINT fk_skills_published_by FOREIGN KEY (published_by) REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE skills ADD CONSTRAINT fk_skills_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE skills ADD CONSTRAINT chk_skills_trust CHECK (trust_score >= 0 AND trust_score <= 1);

CREATE INDEX IF NOT EXISTS idx_skills_category ON skills(category);
CREATE INDEX IF NOT EXISTS idx_skills_tags ON skills USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_skills_trust ON skills(trust_score DESC);
CREATE INDEX IF NOT EXISTS idx_skills_sop ON skills(sop_id);

-- Extend skill_versions table
ALTER TABLE skill_versions ADD COLUMN IF NOT EXISTS changelog TEXT;
ALTER TABLE skill_versions ADD COLUMN IF NOT EXISTS sop_version VARCHAR(50);
ALTER TABLE skill_versions ADD COLUMN IF NOT EXISTS approved_by UUID;
ALTER TABLE skill_versions ADD CONSTRAINT fk_skill_versions_approved_by FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_skill_versions_skill ON skill_versions(skill_id);

-- Skill Dependencies
CREATE TABLE skill_dependencies (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id            UUID NOT NULL,
    dependency_skill_id UUID NOT NULL,
    version_constraint   VARCHAR(100) NOT NULL,
    created_at          TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_skill_deps_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    CONSTRAINT fk_skill_deps_dep FOREIGN KEY (dependency_skill_id) REFERENCES skills(id) ON DELETE RESTRICT,
    UNIQUE(skill_id, dependency_skill_id),
    CONSTRAINT chk_skill_deps_no_self CHECK (skill_id != dependency_skill_id)
);

CREATE INDEX idx_skill_deps_skill ON skill_dependencies(skill_id);
CREATE INDEX idx_skill_deps_dep ON skill_dependencies(dependency_skill_id);

-- Skill Ratings
CREATE TABLE skill_ratings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id    UUID NOT NULL,
    user_id     UUID NOT NULL,
    rating      INTEGER NOT NULL,
    comment     TEXT,
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_skill_ratings_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    CONSTRAINT fk_skill_ratings_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(skill_id, user_id),
    CONSTRAINT chk_skill_ratings CHECK (rating >= 1 AND rating <= 5)
);

CREATE INDEX idx_skill_ratings_skill ON skill_ratings(skill_id);
CREATE INDEX idx_skill_ratings_user ON skill_ratings(user_id);

-- Skill Installs
CREATE TABLE skill_installs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id        UUID NOT NULL,
    user_id         UUID,
    version         VARCHAR(50) NOT NULL,
    environment     VARCHAR(50) DEFAULT 'production',
    installed_at    TIMESTAMP DEFAULT NOW(),
    skill_snapshot  JSONB,
    CONSTRAINT fk_skill_installs_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    CONSTRAINT fk_skill_installs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_skill_installs_env CHECK (environment IN ('production', 'staging', 'development'))
);

CREATE INDEX idx_skill_installs_skill ON skill_installs(skill_id);
CREATE INDEX idx_skill_installs_user ON skill_installs(user_id);
CREATE INDEX idx_skill_installs_env ON skill_installs(environment);

-- Skill Publish Approvals
CREATE TABLE skill_publish_approvals (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id                UUID NOT NULL,
    version                 VARCHAR(50) NOT NULL,
    status                  VARCHAR(50) DEFAULT 'pending',
    submitted_by            UUID NOT NULL,
    submitted_at            TIMESTAMP DEFAULT NOW(),
    reviewed_by             UUID,
    reviewed_at             TIMESTAMP,
    review_note             TEXT,
    compliance_check_passed  BOOLEAN DEFAULT false,
    compliance_check_note   TEXT,
    created_at              TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_skill_pub_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    CONSTRAINT fk_skill_pub_submitted FOREIGN KEY (submitted_by) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_skill_pub_reviewed FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_skill_pub_status CHECK (status IN ('pending', 'approved', 'rejected'))
);

CREATE INDEX idx_skill_pub_skill ON skill_publish_approvals(skill_id);
CREATE INDEX idx_skill_pub_status ON skill_publish_approvals(status);
CREATE INDEX idx_skill_pub_submitted ON skill_publish_approvals(submitted_by);
```

**Step 2: Write down migration**

```sql
ALTER TABLE skill_publish_approvals DROP CONSTRAINT IF EXISTS fk_skill_pub_reviewed;
ALTER TABLE skill_publish_approvals DROP CONSTRAINT IF EXISTS fk_skill_pub_submitted;
ALTER TABLE skill_publish_approvals DROP CONSTRAINT IF EXISTS fk_skill_pub_skill;
ALTER TABLE skill_publish_approvals DROP CONSTRAINT IF EXISTS idx_skill_pub_submitted;
ALTER TABLE skill_publish_approvals DROP CONSTRAINT IF EXISTS idx_skill_pub_status;
ALTER TABLE skill_publish_approvals DROP CONSTRAINT IF EXISTS idx_skill_pub_skill;
DROP TABLE IF EXISTS skill_publish_approvals;

ALTER TABLE skill_installs DROP CONSTRAINT IF EXISTS fk_skill_installs_user;
ALTER TABLE skill_installs DROP CONSTRAINT IF EXISTS fk_skill_installs_skill;
DROP INDEX IF EXISTS idx_skill_installs_env;
DROP INDEX IF EXISTS idx_skill_installs_user;
DROP INDEX IF EXISTS idx_skill_installs_skill;
DROP TABLE IF EXISTS skill_installs;

ALTER TABLE skill_ratings DROP CONSTRAINT IF EXISTS fk_skill_ratings_user;
ALTER TABLE skill_ratings DROP CONSTRAINT IF EXISTS fk_skill_ratings_skill;
DROP INDEX IF EXISTS idx_skill_ratings_user;
DROP INDEX IF EXISTS idx_skill_ratings_skill;
DROP TABLE IF EXISTS skill_ratings;

ALTER TABLE skill_dependencies DROP CONSTRAINT IF EXISTS fk_skill_deps_dep;
ALTER TABLE skill_dependencies DROP CONSTRAINT IF EXISTS fk_skill_deps_skill;
DROP INDEX IF EXISTS idx_skill_deps_dep;
DROP INDEX IF EXISTS idx_skill_deps_skill;
DROP TABLE IF EXISTS skill_dependencies;

ALTER TABLE skill_versions DROP CONSTRAINT IF EXISTS fk_skill_versions_approved_by;
ALTER TABLE skill_versions DROP COLUMN IF EXISTS approved_by;
ALTER TABLE skill_versions DROP COLUMN IF EXISTS sop_version;
ALTER TABLE skill_versions DROP COLUMN IF EXISTS changelog;

ALTER TABLE skills DROP CONSTRAINT IF EXISTS fk_skills_created_by;
ALTER TABLE skills DROP CONSTRAINT IF EXISTS fk_skills_published_by;
ALTER TABLE skills DROP COLUMN IF EXISTS created_by;
ALTER TABLE skills DROP COLUMN IF EXISTS draft_source_meta;
ALTER TABLE skills DROP COLUMN IF EXISTS draft_source;
ALTER TABLE skills DROP COLUMN IF EXISTS published_by;
ALTER TABLE skills DROP COLUMN IF EXISTS published_at;
ALTER TABLE skills DROP COLUMN IF EXISTS verified;
ALTER TABLE skills DROP COLUMN IF EXISTS trust_score;
ALTER TABLE skills DROP COLUMN IF EXISTS tags;
ALTER TABLE skills DROP COLUMN IF EXISTS category;
ALTER TABLE skills DROP COLUMN IF EXISTS compliance_required;
ALTER TABLE skills DROP COLUMN IF EXISTS sop_section;
ALTER TABLE skills DROP COLUMN IF EXISTS sop_version;
ALTER TABLE skills DROP COLUMN IF EXISTS sop_name;
ALTER TABLE skills DROP COLUMN IF EXISTS sop_id;
```

**Step 3: Run migration test**

```bash
cd apps/api && go run cmd/migrate/main.go
```

Expected: Migration runs successfully

---

## Phase 2: Backend - MCP Router

### Task 3: Create MCP Router Types

**Files:**
- Create: `apps/api/internal/server/types_mcp_router.go`

**Step 1: Define types**

```go
package server

// MCPRouterCatalog represents an MCP server in the routing pool
type MCPRouterCatalog struct {
    ID             string    `json:"id"`
    ServerID       string    `json:"server_id"`
    Name           string    `json:"name"`
    Description    string    `json:"description,omitempty"`
    TransportType  string    `json:"transport_type"`
    Command        string    `json:"command,omitempty"`
    Args           []string  `json:"args,omitempty"`
    URL            string    `json:"url,omitempty"`
    TaskTypes      []string  `json:"task_types"`
    Tags           []string  `json:"tags,omitempty"`
    Capabilities   JSONMap   `json:"capabilities"`
    RoutingConfig  JSONMap   `json:"routing_config,omitempty"`
    Status         string    `json:"status"`
    TrustScore     float64   `json:"trust_score"`
    UseCount       int64     `json:"use_count"`
    ErrorCount     int64     `json:"error_count"`
    LastUsedAt     *string   `json:"last_used_at,omitempty"`
    SyncedAt       string    `json:"synced_at"`
    CreatedAt      string    `json:"created_at"`
}

// MCPRouterLog represents a routing request log
type MCPRouterLog struct {
    ID                string  `json:"id"`
    CorrelationID     string  `json:"correlation_id"`
    AgentID           string  `json:"agent_id,omitempty"`
    TaskType          string  `json:"task_type,omitempty"`
    TaskTags          []string `json:"task_tags,omitempty"`
    TaskComplexity    string  `json:"task_complexity,omitempty"`
    RawDescription    string  `json:"raw_description,omitempty"`
    Matched           bool    `json:"matched"`
    MatchedServerID   string  `json:"matched_server_id,omitempty"`
    MatchScore        float64 `json:"match_score,omitempty"`
    MatchLatencyMS    int     `json:"match_latency_ms,omitempty"`
    Status            string  `json:"status,omitempty"`
    ErrorCode         string  `json:"error_code,omitempty"`
    ErrorMessage      string  `json:"error_message,omitempty"`
    DurationMS        int     `json:"duration_ms,omitempty"`
    CreatedAt         string  `json:"created_at"`
}

// RouteRequest represents an MCP routing request
type RouteRequest struct {
    Task            RouteTask `json:"task"`
    AgentID         string    `json:"agent_id,omitempty"`
    CorrelationID   string    `json:"correlation_id,omitempty"`
}

type RouteTask struct {
    Description   string          `json:"description,omitempty"`
    Structured    TaskStructured `json:"structured,omitempty"`
}

type TaskStructured struct {
    TaskType     string `json:"task_type,omitempty"`
    Language     string `json:"language,omitempty"`
    Complexity   string `json:"complexity,omitempty"`
    Tags         []string `json:"tags,omitempty"`
}

// RouteResponse represents an MCP routing response
type RouteResponse struct {
    Matched      bool           `json:"matched"`
    Target       *RouteTarget   `json:"target,omitempty"`
    MatchScore   float64        `json:"match_score,omitempty"`
    RoutingTimeMS int           `json:"routing_time_ms,omitempty"`
}

type RouteTarget struct {
    ServerID     string `json:"server_id"`
    ServerName   string `json:"server_name"`
    Transport    string `json:"transport"`
    Endpoint     string `json:"endpoint,omitempty"`
}
```

**Step 2: Run build verification**

```bash
cd apps/api && go build ./...
```

Expected: Build succeeds

---

### Task 4: Create MCP Router Service

**Files:**
- Create: `apps/api/internal/service/mcp_router_service.go`
- Create: `apps/api/internal/service/mcp_router_metrics.go`

**Step 1: Write service**

```go
package service

import (
    "context"
    "fmt"
    "time"

    "github.com/zealot/managing-up/apps/api/internal/repository"
)

type MCPRouterService struct {
    repo             repository.MCPRouterRepository
    metricsCollector *MetricsCollector
}

func NewMCPRouterService(repo repository.MCPRouterRepository, mc *MetricsCollector) *MCPRouterService {
    return &MCPRouterService{repo: repo, metricsCollector: mc}
}

// MatchTask matches a task to the best MCP server
func (s *MCPRouterService) MatchTask(ctx context.Context, taskTypes []string, tags []string) (*MatchResult, error) {
    start := time.Now()
    
    // Query matching servers
    servers, err := s.repo.FindMatchingServers(ctx, taskTypes, tags)
    if err != nil {
        return nil, fmt.Errorf("failed to find matching servers: %w", err)
    }
    
    if len(servers) == 0 {
        return &MatchResult{Matched: false}, nil
    }
    
    // Sort by trust_score DESC, use_count DESC
    best := servers[0]
    
    // Increment use count
    s.repo.IncrementUseCount(ctx, best.ID)
    
    return &MatchResult{
        Matched:    true,
        ServerID:   best.ServerID,
        ServerName: best.Name,
        Transport:  best.TransportType,
        Endpoint:   best.URL,
        MatchScore: best.TrustScore,
    }, nil
}

type MatchResult struct {
    Matched    bool
    ServerID   string
    ServerName string
    Transport  string
    Endpoint   string
    MatchScore float64
}

// SyncFromMCPServer syncs an approved MCP server to the router catalog
func (s *MCPRouterService) SyncFromMCPServer(ctx context.Context, serverID string, approvedBy string) error {
    return s.repo.SyncServer(ctx, serverID, approvedBy)
}

// GetCatalog returns all servers in the routing catalog
func (s *MCPRouterService) GetCatalog(ctx context.Context) ([]RouterCatalogEntry, error) {
    return s.repo.ListCatalog(ctx)
}
```

**Step 2: Write metrics collector**

```go
package service

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricsCollector struct {
    RequestsTotal    *prometheus.CounterVec
    RequestDuration  *prometheus.HistogramVec
    MatchFailures    *prometheus.CounterVec
}

func NewMetricsCollector() *MetricsCollector {
    return &MetricsCollector{
        RequestsTotal: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mcp_router_requests_total",
                Help: "Total MCP router requests",
            },
            []string{"agent", "task_type", "status"},
        ),
        RequestDuration: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "mcp_router_request_duration_seconds",
                Help:    "MCP router request latency",
                Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
            },
            []string{"agent", "task_type"},
        ),
        MatchFailures: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mcp_router_match_failures_total",
                Help: "Route match failures",
            },
            []string{"reason"},
        ),
    }
}

func (m *MetricsCollector) RecordRequest(agent, taskType, status string, duration float64) {
    m.RequestsTotal.WithLabelValues(agent, taskType, status).Inc()
    m.RequestDuration.WithLabelValues(agent, taskType).Observe(duration)
}

func (m *MetricsCollector) RecordMatchFailure(reason string) {
    m.MatchFailures.WithLabelValues(reason).Inc()
}
```

**Step 3: Run build verification**

```bash
cd apps/api && go build ./...
```

Expected: Build succeeds

---

### Task 5: Create MCP Router Repository

**Files:**
- Create: `apps/api/internal/repository/mcp_router_repo.go`

**Step 1: Write repository interface and implementation**

```go
package repository

import (
    "context"
    "database/sql"
    "fmt"
    "time"
)

type MCPRouterRepository interface {
    FindMatchingServers(ctx context.Context, taskTypes []string, tags []string) ([]RouterCatalogEntry, error)
    ListCatalog(ctx context.Context) ([]RouterCatalogEntry, error)
    IncrementUseCount(ctx context.Context, id string) error
    SyncServer(ctx context.Context, serverID string, approvedBy string) error
    GetServerByID(ctx context.Context, serverID string) (*MCPServer, error)
    LogRoute(ctx context.Context, log *RouteLogEntry) error
}

type RouterCatalogEntry struct {
    ID            string
    ServerID      string
    Name          string
    TransportType string
    Command       string
    Args          []string
    URL           string
    TaskTypes     []string
    Tags          []string
    TrustScore    float64
    UseCount      int64
}

type RouteLogEntry struct {
    CorrelationID   string
    AgentID         string
    TaskType        string
    TaskTags        []string
    RawDescription  string
    Matched         bool
    MatchedServerID string
    MatchScore      float64
    MatchLatencyMS  int
    Status          string
    ErrorCode       string
    ErrorMessage    string
    DurationMS      int
}

type PostgresMCPRouterRepository struct {
    db *sql.DB
}

func NewPostgresMCPRouterRepository(db *sql.DB) *PostgresMCPRouterRepository {
    return &PostgresMCPRouterRepository{db: db}
}

func (r *PostgresMCPRouterRepository) FindMatchingServers(ctx context.Context, taskTypes []string, tags []string) ([]RouterCatalogEntry, error) {
    query := `
        SELECT id, server_id, name, transport_type, command, args, url, task_types, tags, trust_score, use_count
        FROM mcp_router_catalog
        WHERE status = 'active'
        AND task_types @> $1
        ORDER BY trust_score DESC, use_count DESC
        LIMIT 1
    `
    
    var entry RouterCatalogEntry
    var argsJSON []byte
    err := r.db.QueryRowContext(ctx, query, taskTypes).Scan(
        &entry.ID, &entry.ServerID, &entry.Name, &entry.TransportType,
        &entry.Command, &argsJSON, &entry.URL, &entry.TaskTypes, &entry.Tags,
        &entry.TrustScore, &entry.UseCount,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    
    return []RouterCatalogEntry{entry}, nil
}

func (r *PostgresMCPRouterRepository) IncrementUseCount(ctx context.Context, id string) error {
    query := `UPDATE mcp_router_catalog SET use_count = use_count + 1, last_used_at = $1 WHERE id = $2`
    _, err := r.db.ExecContext(ctx, query, time.Now(), id)
    return err
}

func (r *PostgresMCPRouterRepository) SyncServer(ctx context.Context, serverID string, approvedBy string) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Get MCP server details
    var server MCPServer
    err = tx.QueryRowContext(ctx, `
        SELECT id, name, description, transport_type, command, args, env, url, headers
        FROM mcp_servers WHERE id = $1
    `, serverID).Scan(&server.ID, &server.Name, &server.Description, &server.TransportType,
        &server.Command, &server.Args, &server.Env, &server.URL, &server.Headers)
    if err != nil {
        return fmt.Errorf("failed to get server: %w", err)
    }
    
    // Insert or update catalog
    _, err = tx.ExecContext(ctx, `
        INSERT INTO mcp_router_catalog (server_id, name, description, transport_type, command, args, url, task_types, tags, capabilities, status)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, '{}', 'active')
        ON CONFLICT (server_id) DO UPDATE SET
            name = EXCLUDED.name,
            description = EXCLUDED.description,
            transport_type = EXCLUDED.transport_type,
            command = EXCLUDED.command,
            args = EXCLUDED.args,
            url = EXCLUDED.url,
            synced_at = NOW()
    `, server.ID, server.Name, server.Description, server.TransportType, server.Command, server.Args, server.URL, server.Tags)
    if err != nil {
        return fmt.Errorf("failed to sync catalog: %w", err)
    }
    
    // Log sync
    _, err = tx.ExecContext(ctx, `
        INSERT INTO mcp_router_sync_log (server_id, sync_type, approved_by, approved_at)
        VALUES ($1, 'approved_sync', $2, NOW())
    `, serverID, approvedBy)
    if err != nil {
        return fmt.Errorf("failed to log sync: %w", err)
    }
    
    return tx.Commit()
}
```

**Step 2: Run build verification**

```bash
cd apps/api && go build ./...
```

Expected: Build succeeds

---

### Task 6: Create MCP Router Handlers and Routes

**Files:**
- Modify: `apps/api/internal/server/server.go` (add routes)
- Create: `apps/api/internal/server/handlers/mcp_router.go`

**Step 1: Write handler**

```go
package handlers

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/zealot/managing-up/apps/api/internal/service"
)

type MCPRouterHandler struct {
    routerSvc *service.MCPRouterService
    metrics   *service.MetricsCollector
}

func NewMCPRouterHandler(routerSvc *service.MCPRouterService, metrics *service.MetricsCollector) *MCPRouterHandler {
    return &MCPRouterHandler{routerSvc: routerSvc, metrics: metrics}
}

func (h *MCPRouterHandler) Route(w http.ResponseWriter, r *http.Request) {
    var req RouteRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    start := time.Now()
    
    // Extract task info
    taskTypes := []string{}
    tags := []string{}
    
    if req.Task.Structured.TaskType != "" {
        taskTypes = []string{req.Task.Structured.TaskType}
    }
    if len(req.Task.Structured.Tags) > 0 {
        tags = req.Task.Structured.Tags
    }
    
    // Match
    result, err := h.routerSvc.MatchTask(ctx, taskTypes, tags)
    duration := time.Since(start).Seconds()
    
    if err != nil {
        h.metrics.RecordRequest(req.AgentID, req.Task.Structured.TaskType, "failure", duration)
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    if !result.Matched {
        h.metrics.RecordMatchFailure("no_matching_server")
        h.metrics.RecordRequest(req.AgentID, req.Task.Structured.TaskType, "no_match", duration)
        writeEnvelope(w, RouteResponse{Matched: false}, req.CorrelationID)
        return
    }
    
    h.metrics.RecordRequest(req.AgentID, req.Task.Structured.TaskType, "success", duration)
    
    writeEnvelope(w, RouteResponse{
        Matched: true,
        Target: &RouteTarget{
            ServerID:   result.ServerID,
            ServerName: result.ServerName,
            Transport:  result.Transport,
            Endpoint:   result.Endpoint,
        },
        MatchScore:    result.MatchScore,
        RoutingTimeMS: int(time.Since(start).Milliseconds()),
    }, req.CorrelationID)
}

func (h *MCPRouterHandler) Catalog(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    entries, err := h.routerSvc.GetCatalog(ctx)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    writeEnvelope(w, entries, "")
}

func (h *MCPRouterHandler) Match(w http.ResponseWriter, r *http.Request) {
    taskType := r.URL.Query().Get("task_type")
    tagsParam := r.URL.Query().Get("tags")
    
    tags := []string{}
    if tagsParam != "" {
        tags = splitTags(tagsParam)
    }
    
    ctx := r.Context()
    result, err := h.routerSvc.MatchTask(ctx, []string{taskType}, tags)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    writeEnvelope(w, result, "")
}

func splitTags(s string) []string {
    if s == "" {
        return nil
    }
    var tags []string
    for _, t := range splitString(s, ",") {
        tags = append(tags, trimSpace(t))
    }
    return tags
}

func splitString(s, sep string) []string {
    var result []string
    for i := 0; i < len(s); {
        idx := indexOf(s, sep, i)
        if idx == -1 {
            result = append(result, s[i:])
            break
        }
        result = append(result, s[i:idx])
        i = idx + len(sep)
    }
    return result
}

func indexOf(s, substr string, start int) int {
    for i := start; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return i
        }
    }
    return -1
}

func trimSpace(s string) string {
    start, end := 0, len(s)
    for start < end && (s[start] == ' ' || s[start] == '\t') {
        start++
    }
    for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
        end--
    }
    return s[start:end]
}
```

**Step 2: Add routes to server.go**

In `NewWithRepository()` function, add:

```go
// MCP Router routes
mux.HandleFunc("/api/v1/router/mcp/route", srv.mcpRouterHandler.Route)
mux.HandleFunc("/api/v1/router/mcp/catalog", srv.mcpRouterHandler.Catalog)
mux.HandleFunc("/api/v1/router/mcp/match", srv.mcpRouterHandler.Match)

// Prometheus metrics
mux.Handle("/metrics", promhttp.Handler())
```

**Step 3: Run build verification**

```bash
cd apps/api && go build ./...
```

Expected: Build succeeds

---

### Task 7: Integrate MCP Server Approval → Router Sync

**Files:**
- Modify: `apps/api/internal/server/handlers/mcp_servers.go` (in approval handler)

**Step 1: Modify approval handler**

In the approval handler, after `UpdateMCPServerStatus`, add:

```go
// Sync to router catalog if approved
if decision == "approved" {
    approvedBy := getUserID(r) // get from auth context
    if err := h.mcpRouterSvc.SyncFromMCPServer(ctx, serverID, approvedBy); err != nil {
        // Log error but don't fail the approval
        log.Printf("failed to sync to router catalog: %v", err)
    }
}
```

**Step 2: Run build verification**

```bash
cd apps/api && go build ./...
```

Expected: Build succeeds

---

## Phase 3: Backend - Skill Repository Extensions

### Task 8: Extend Skill Models

**Files:**
- Modify: `apps/api/internal/server/types.go`

**Step 1: Add new fields to Skill type**

```go
// Add to existing Skill struct
type Skill struct {
    // ... existing fields ...
    
    // Enterprise extensions
    SOPID              string    `json:"sop_id,omitempty"`
    SOPName            string    `json:"sop_name,omitempty"`
    SOPVersion         string    `json:"sop_version,omitempty"`
    SOPSection         string    `json:"sop_section,omitempty"`
    ComplianceRequired bool      `json:"compliance_required"`
    Category           string    `json:"category,omitempty"`
    Tags               []string `json:"tags,omitempty"`
    TrustScore         float64  `json:"trust_score"`
    Verified           bool      `json:"verified"`
    PublishedAt        string    `json:"published_at,omitempty"`
    PublishedBy        string    `json:"published_by,omitempty"`
    DraftSource        string    `json:"draft_source"`
    DraftSourceMeta    JSONMap   `json:"draft_source_meta,omitempty"`
    CreatedBy          string    `json:"created_by,omitempty"`
}

// New types for skill dependencies
type SkillDependency struct {
    ID                string `json:"id"`
    SkillID           string `json:"skill_id"`
    DependencySkillID string `json:"dependency_skill_id"`
    VersionConstraint string `json:"version_constraint"`
    CreatedAt         string `json:"created_at"`
}

type SkillRating struct {
    ID        string `json:"id"`
    SkillID   string `json:"skill_id"`
    UserID    string `json:"user_id"`
    Rating    int    `json:"rating"`
    Comment   string `json:"comment,omitempty"`
    CreatedAt string `json:"created_at"`
}

type SkillInstall struct {
    ID             string   `json:"id"`
    SkillID        string   `json:"skill_id"`
    UserID         string   `json:"user_id,omitempty"`
    Version        string   `json:"version"`
    Environment    string   `json:"environment"`
    InstalledAt    string   `json:"installed_at"`
    SkillSnapshot  JSONMap  `json:"skill_snapshot,omitempty"`
}
```

**Step 2: Run build verification**

```bash
cd apps/api && go build ./...
```

Expected: Build succeeds

---

### Task 9: Create Skill Repository Service

**Files:**
- Create: `apps/api/internal/service/skill_enterprise_service.go`

**Step 1: Write service**

```go
package service

import (
    "context"
    "fmt"
)

type SkillEnterpriseService struct {
    repo SkillRepository
}

func NewSkillEnterpriseService(repo SkillRepository) *SkillEnterpriseService {
    return &SkillEnterpriseService{repo: repo}
}

func (s *SkillEnterpriseService) GetSkillWithDeps(ctx context.Context, skillID string) (*SkillWithDeps, error) {
    skill, err := s.repo.GetSkill(skillID)
    if err != nil {
        return nil, err
    }
    
    deps, err := s.repo.ListDependencies(ctx, skillID)
    if err != nil {
        return nil, fmt.Errorf("failed to get dependencies: %w", err)
    }
    
    return &SkillWithDeps{
        Skill:       skill,
        Dependencies: deps,
    }, nil
}

func (s *SkillEnterpriseService) RateSkill(ctx context.Context, skillID, userID string, rating int, comment string) error {
    if rating < 1 || rating > 5 {
        return ErrInvalidRating
    }
    return s.repo.UpsertRating(ctx, skillID, userID, rating, comment)
}

func (s *SkillEnterpriseService) GetSkillMarket(ctx context.Context, category, search string) ([]SkillMarketEntry, error) {
    skills, err := s.repo.ListSkillsByCategory(ctx, category, search)
    if err != nil {
        return nil, err
    }
    
    var entries []SkillMarketEntry
    for _, skill := range skills {
        avgRating, count, _ := s.repo.GetRatingStats(ctx, skill.ID)
        installCount, _ := s.repo.GetInstallCount(ctx, skill.ID)
        
        entries = append(entries, SkillMarketEntry{
            Skill:        skill,
            AvgRating:    avgRating,
            RatingCount:  count,
            InstallCount: installCount,
        })
    }
    
    return entries, nil
}

func (s *SkillEnterpriseService) ResolveDependencies(ctx context.Context, skillID string) ([]DependencyNode, error) {
    return s.repo.ResolveDepTree(ctx, skillID)
}

type SkillWithDeps struct {
    Skill       Skill
    Dependencies []SkillDependency
}

type SkillMarketEntry struct {
    Skill
    AvgRating    float64 `json:"avg_rating"`
    RatingCount  int     `json:"rating_count"`
    InstallCount int     `json:"install_count"`
}

type DependencyNode struct {
    SkillID   string           `json:"skill_id"`
    Name      string           `json:"name"`
    Version   string           `json:"version"`
    Children  []DependencyNode  `json:"children,omitempty"`
}
```

**Step 2: Run build verification**

```bash
cd apps/api && go build ./...
```

Expected: Build succeeds

---

### Task 10: Add Skill Market API Handlers

**Files:**
- Modify: `apps/api/internal/server/server.go` (add routes)
- Create: `apps/api/internal/server/handlers/skill_market.go`

**Step 1: Write handlers**

```go
func (srv *Server) handleSkillMarket(w http.ResponseWriter, r *http.Request) {
    category := r.URL.Query().Get("category")
    search := r.URL.Query().Get("search")
    
    entries, err := srv.skillEnterpriseSvc.GetSkillMarket(r.Context(), category, search)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    writeEnvelope(w, entries, "")
}

func (srv *Server) handleSkillDependencies(w http.ResponseWriter, r *http.Request) {
    id := trimPrefix(r.URL.Path, "/api/v1/skills/")
    id = trimSuffix(id, "/dependencies")
    
    deps, err := srv.skillEnterpriseSvc.GetSkillWithDeps(r.Context(), id)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    writeEnvelope(w, deps, "")
}

func (srv *Server) handleSkillRate(w http.ResponseWriter, r *http.Request) {
    id := trimPrefix(r.URL.Path, "/api/v1/skills/")
    id = trimSuffix(id, "/rate")
    
    var req RateSkillRequest
    if err := decodeJSON(r, &req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request")
        return
    }
    
    userID := getUserID(r)
    if err := srv.skillEnterpriseSvc.RateSkill(r.Context(), id, userID, req.Rating, req.Comment); err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }
    
    writeEnvelope(w, map[string]bool{"success": true}, "")
}

func (srv *Server) handleSkillResolveDeps(w http.ResponseWriter, r *http.Request) {
    var req struct {
        SkillID string `json:"skill_id"`
    }
    if err := decodeJSON(r, &req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request")
        return
    }
    
    deps, err := srv.skillEnterpriseSvc.ResolveDependencies(r.Context(), req.SkillID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    writeEnvelope(w, deps, "")
}

func trimPrefix(s, prefix string) string {
    if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
        return s[len(prefix):]
    }
    return s
}

func trimSuffix(s, suffix string) string {
    if len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
        return s[:len(s)-len(suffix)]
    }
    return s
}
```

**Step 2: Add routes to server.go**

```go
// Skill market routes
mux.HandleFunc("/api/v1/skills/market", srv.handleSkillMarket)
mux.HandleFunc("/api/v1/skills/search", srv.handleSkillSearch)
mux.HandleFunc("/api/v1/skills/", srv.handleSkillByID)
```

**Step 3: Run build verification**

```bash
cd apps/api && go build ./...
```

Expected: Build succeeds

---

## Phase 4: Frontend - MCP Router Pages

### Task 11: Create MCP Router API Client

**Files:**
- Modify: `apps/web/lib/api.ts`

**Step 1: Add MCP Router API functions**

```typescript
interface MCPRouterCatalogEntry {
  id: string;
  server_id: string;
  name: string;
  description?: string;
  transport_type: string;
  task_types: string[];
  tags?: string[];
  trust_score: number;
  use_count: number;
  status: string;
}

interface RouteRequest {
  task: {
    description?: string;
    structured?: {
      task_type?: string;
      language?: string;
      complexity?: string;
      tags?: string[];
    };
  };
  agent_id?: string;
  correlation_id?: string;
}

interface RouteResponse {
  matched: boolean;
  target?: {
    server_id: string;
    server_name: string;
    transport: string;
    endpoint?: string;
  };
  match_score?: number;
  routing_time_ms?: number;
}

async function getMCPRouterCatalog(): Promise<MCPRouterCatalogEntry[]> {
  const response = await fetch(`${API_BASE_URL}/api/v1/router/mcp/catalog`);
  const body = await response.json();
  return body.data;
}

async function routeMCP(request: RouteRequest): Promise<RouteResponse> {
  const response = await fetch(`${API_BASE_URL}/api/v1/router/mcp/route`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(request),
  });
  const body = await response.json();
  return body.data;
}

async function matchMCPRouter(taskType: string, tags?: string[]): Promise<RouteResponse | null> {
  const params = new URLSearchParams({ task_type: taskType });
  if (tags?.length) {
    params.set("tags", tags.join(","));
  }
  const response = await fetch(`${API_BASE_URL}/api/v1/router/mcp/match?${params}`);
  const body = await response.json();
  return body.data;
}
```

**Step 2: Run TypeScript check**

```bash
cd apps/web && npx tsc --noEmit
```

Expected: No errors

---

### Task 12: Create MCP Router Dashboard Page

**Files:**
- Create: `apps/web/app/mcp-router/page.tsx`
- Create: `apps/web/app/mcp-router/MCPRouterDashboardClient.tsx`

**Step 1: Create server page**

```tsx
import { Suspense } from "react";
import { MCPRouterDashboardClient } from "./MCPRouterDashboardClient";
import { Skeleton } from "app/components/ui/Skeleton";

export default function MCPRouterPage() {
  return (
    <Suspense fallback={<Skeleton className="h-[400px]" />}>
      <MCPRouterDashboardClient />
    </Suspense>
  );
}
```

**Step 2: Create client component**

```tsx
"use client";

import { useQuery } from "@tanstack/react-query";
import { getMCPRouterCatalog } from "app/lib/api";

export function MCPRouterDashboardClient() {
  const { data: catalog, isLoading } = useQuery({
    queryKey: ["mcp-router-catalog"],
    queryFn: getMCPRouterCatalog,
  });

  if (isLoading) {
    return <div>Loading...</div>;
  }

  const stats = {
    total: catalog?.length ?? 0,
    active: catalog?.filter(s => s.status === "active").length ?? 0,
    avgTrust: catalog?.reduce((sum, s) => sum + s.trust_score, 0) / (catalog?.length || 1) ?? 0,
  };

  return (
    <div>
      <h1>MCP Router Dashboard</h1>
      <div className="grid grid-cols-3 gap-4">
        <StatCard label="Total Servers" value={stats.total} />
        <StatCard label="Active" value={stats.active} />
        <StatCard label="Avg Trust Score" value={stats.avgTrust.toFixed(2)} />
      </div>
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="card">
      <div className="text-sm text-muted">{label}</div>
      <div className="text-2xl font-bold">{value}</div>
    </div>
  );
}
```

**Step 3: Run build verification**

```bash
cd apps/web && npm run build 2>&1 | tail -20
```

Expected: Build succeeds

---

### Task 13: Create MCP Router Metrics Page

**Files:**
- Create: `apps/web/app/mcp-router/metrics/page.tsx`

**Step 1: Create metrics page**

```tsx
export default function MCPRouterMetricsPage() {
  return (
    <div>
      <h1>MCP Router Metrics</h1>
      <p className="text-muted">
        Prometheus metrics available at <code>/metrics</code> endpoint
      </p>
      <MetricsChart />
    </div>
  );
}

function MetricsChart() {
  // Placeholder for actual metrics visualization
  // Could use a library like recharts or integrate with Grafana
  return (
    <div className="card">
      <h3>Request Rate</h3>
      <p>Integration with Prometheus/Grafana coming soon</p>
    </div>
  );
}
```

**Step 2: Run build verification**

```bash
cd apps/web && npm run build 2>&1 | tail -20
```

Expected: Build succeeds

---

## Phase 5: Frontend - Skill Repository Pages

### Task 14: Create Skill Market Page

**Files:**
- Create: `apps/web/app/skills/market/page.tsx`
- Create: `apps/web/app/skills/market/SkillMarketClient.tsx`

**Step 1: Create server page**

```tsx
import { Suspense } from "react";
import { SkillMarketClient } from "./SkillMarketClient";
import { Skeleton } from "app/components/ui/Skeleton";

export default function SkillMarketPage() {
  return (
    <Suspense fallback={<Skeleton className="h-[400px]" />}>
      <SkillMarketClient />
    </Suspense>
  );
}
```

**Step 2: Create client component**

```tsx
"use client";

import { useQuery } from "@tanstack/react-query";
import { getSkillMarket } from "app/lib/api";
import { DataToolbar } from "app/components/ui/DataToolbar";
import { Card, CardGrid } from "app/components/ui/Card";

export function SkillMarketClient() {
  const { data: skills, isLoading } = useQuery({
    queryKey: ["skill-market"],
    queryFn: () => getSkillMarket({ category: "", search: "" }),
  });

  return (
    <div>
      <h1>Skill Market</h1>
      <DataToolbar
        onSearch={(q) => {}}
        onFilter={(f) => {}}
        filters={["category", "rating", "verified"]}
      />
      <CardGrid>
        {skills?.map((skill) => (
          <Card key={skill.id}>
            <h3>{skill.name}</h3>
            <p>{skill.description}</p>
            <div className="flex gap-2">
              {skill.tags?.map((tag) => (
                <span key={tag} className="badge">{tag}</span>
              ))}
            </div>
            <div className="flex justify-between items-center mt-4">
              <span>Trust: {skill.trust_score.toFixed(2)}</span>
              {skill.verified && <span className="badge badge-success">Verified</span>}
            </div>
          </Card>
        ))}
      </CardGrid>
    </div>
  );
}
```

**Step 3: Run build verification**

```bash
cd apps/web && npm run build 2>&1 | tail -20
```

Expected: Build succeeds

---

### Task 15: Create My Skills Page

**Files:**
- Create: `apps/web/app/skills/my-skills/page.tsx`
- Create: `apps/web/app/skills/my-skills/MySkillsClient.tsx`

**Step 1: Create client component**

```tsx
"use client";

import { useQuery } from "@tanstack/react-query";
import { getMySkills } from "app/lib/api";
import { Badge } from "app/components/ui/Badge";
import { Card, CardGrid } from "app/components/ui/Card";

export function MySkillsClient() {
  const { data: skills, isLoading } = useQuery({
    queryKey: ["my-skills"],
    queryFn: () => getMySkills(),
  });

  const drafts = skills?.filter(s => s.status === "draft") ?? [];
  const published = skills?.filter(s => s.status === "published") ?? [];
  const installed = skills?.filter(s => s.status === "installed") ?? [];

  return (
    <div>
      <h1>My Skills</h1>
      
      <section>
        <h2>Drafts ({drafts.length})</h2>
        <CardGrid>
          {drafts.map((skill) => (
            <Card key={skill.id}>
              <h3>{skill.name}</h3>
              <p>{skill.description}</p>
              <Badge variant="warning">{skill.draft_source}</Badge>
            </Card>
          ))}
        </CardGrid>
      </section>

      <section>
        <h2>Published ({published.length})</h2>
        <CardGrid>
          {published.map((skill) => (
            <Card key={skill.id}>
              <h3>{skill.name}</h3>
              <p>{skill.description}</p>
              <div className="mt-2">
                <span className="text-sm">SOP: {skill.sop_name}</span>
              </div>
            </Card>
          ))}
        </CardGrid>
      </section>
    </div>
  );
}
```

**Step 2: Run build verification**

```bash
cd apps/web && npm run build 2>&1 | tail -20
```

Expected: Build succeeds

---

## Phase 6: Final Integration & Testing

### Task 16: Run Full Test Suite

**Step 1: Run Go tests**

```bash
cd apps/api && go test ./... 2>&1 | tail -30
```

**Step 2: Run frontend build**

```bash
cd apps/web && npm run build 2>&1 | tail -30
```

---

## Summary

| Phase | Tasks | Files Created/Modified |
|-------|-------|------------------------|
| 1. Migrations | 2 | 4 migration files |
| 2. MCP Router | 5 | 6 files |
| 3. Skill Extensions | 3 | 4 files |
| 4. MCP Frontend | 3 | 6 files |
| 5. Skill Frontend | 2 | 4 files |
| 6. Testing | 1 | - |

**Total: 16 tasks across 6 phases**

---

**Plan complete and saved to `docs/plans/2026-04-08-agent-ecosystem-hub-implementation-plan.md`**

**Two execution options:**

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**
