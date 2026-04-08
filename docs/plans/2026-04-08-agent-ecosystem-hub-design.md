# Agent Ecosystem Hub 设计文档

**版本**：v1.0
**日期**：2026-04-08
**状态**：已评审

---

## 一、概述

### 1.1 背景

managing-up 项目定位为 AI Agent 生态系统的基础设施，目标是成为企业级 AI Agent 的"npm registry"。企业需要可靠、可追溯的 Skill、MCP 服务、Plugin 和远程记忆管理能力。

### 1.2 目标

| 模块 | 目标 |
|-----|------|
| **MCP Router** | 无状态接入层，基于任务类型+标签智能路由到最优 MCP 服务 |
| **Skill Repository** | 统一 Skill 平台，支持依赖管理、版本控制、市场发现、SOP 参照 |
| **Plugin Hub** | 企业级 Plugin 管理（规划中） |
| **Memory Hub** | 远程记忆管理（规划中） |

### 1.3 设计原则

1. **MCP Router 独立部署** — 无状态接入层，不影响现有 MCP Server Management
2. **Skill 统一平台** — 重构现有 Skill Registry，支持完整企业级功能
3. **松耦合** — 各 Hub 可独立演进，通过标准接口通信
4. **OpenClaw 兼容** — 接口能力兼容 OpenClaw agent 调用

---

## 二、架构

```
┌─────────────────────────────────────────────────────────────────┐
│                    Agent Ecosystem Hub                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              MCP Router (无状态接入层)                    │  │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │  │
│  │  │ Ingress │  │ Task    │  │ Match   │  │ Egress  │   │  │
│  │  │ Handler │──│ Parser  │──│ Engine  │──│ Proxy   │   │  │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │  │
│  │       │             │             │             │       │  │
│  │       └─────────────┴─────────────┴─────────────┘       │  │
│  │                      │              │                     │  │
│  │                      ▼              ▼                     │  │
│  │            ┌─────────────────┐ ┌──────────────┐         │  │
│  │            │  MCP Registry   │ │   Metrics    │         │  │
│  │            │  (分类+标签)     │ │  /metrics     │         │  │
│  │            └─────────────────┘ └──────────────┘         │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Unified Skill Repository                    │  │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │  │
│  │  │ Skill   │  │ Version │  │Depend   │  │Install  │   │  │
│  │  │ Registry │  │ Manager │  │Resolver │  │ Manager │   │  │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │  │
│  │       │             │             │             │       │  │
│  │       └─────────────┴─────────────┴─────────────┘       │  │
│  │                      │                                     │  │
│  │                      ▼                                     │  │
│  │            ┌─────────────────┐                           │  │
│  │            │  Skill Market   │ ◄── Discovery + Rating   │  │
│  │            └─────────────────┘                           │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 三、API 设计

### 3.1 MCP Router API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/router/mcp/route` | 路由请求 |
| GET | `/api/v1/router/mcp/catalog` | 可用 MCP 服务目录 |
| GET | `/api/v1/router/mcp/match` | 查询匹配结果（调试用） |
| GET | `/metrics` | Prometheus metrics |

**路由请求示例**：

```bash
POST /api/v1/router/mcp/route
X-Agent-ID: openclaw-v1
X-Correlation-ID: req-12345

{
  "task": {
    "description": "帮我审查这段 Python 代码的性能问题",
    "structured": {
      "task_type": "code_review",
      "language": "python",
      "complexity": "high"
    }
  }
}
```

**路由响应示例（OpenClaw envelope 兼容）**：

```json
{
  "data": {
    "matched": true,
    "target": {
      "server_id": "mcp-xxx",
      "server_name": "code-analysis-service",
      "transport": "stdio",
      "endpoint": "/path/to/mcp-server"
    },
    "match_score": 0.95,
    "routing_time_ms": 12
  },
  "meta": {
    "request_id": "req-12345",
    "agent_id": "openclaw-v1"
  }
}
```

### 3.2 Skill Repository API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/skills` | 列出 Skills（status=published） |
| GET | `/api/v1/skills/{id}` | 获取 Skill 元数据 |
| GET | `/api/v1/skills/{id}/spec` | 下载 Skill Spec YAML |
| POST | `/api/v1/executions` | 触发 Skill 执行 |
| GET | `/api/v1/executions/{execId}` | 获取执行状态 |
| POST | `/api/v1/agents` | 注册 Agent |
| GET | `/api/v1/skills/market` | 浏览市场 |
| GET | `/api/v1/skills/search` | 搜索 Skills |
| POST | `/api/v1/skills/{id}/rate` | 评分 |
| GET | `/api/v1/skills/{id}/dependencies` | 查看依赖 |
| POST | `/api/v1/skills/resolve` | 解析依赖树 |
| GET | `/api/v1/skills/{id}/dependents` | 查看被依赖 |
| GET | `/api/v1/skills/{id}/versions` | 列出版本 |
| POST | `/api/v1/skills/{id}/versions` | 创建版本 |
| POST | `/api/v1/skills/{id}/rollback/{ver}` | 回滚 |
| POST | `/api/v1/skills/{id}/promote/{ver}` | 升版 |

### 3.3 Skill Spec 定义

```yaml
name: code-review-gpt
version: 1.0.0
risk_level: medium
description: AI 代码审查技能
inputs:
  - name: code
    type: string
    required: true
steps:
  - id: review
    type: tool
    tool_ref: openai-gpt4
    with:
      model: gpt-4o
      task: code_review

# SOP 参照
sop_reference:
  sop_id: SOP-DEV-001
  sop_name: 代码审查标准操作规程
  sop_version: "2.1"
  sop_section: "4.3 自动化审查"
  compliance_required: true

# 企业级扩展
enterprise:
  category: code_analysis
  tags: [ai, review, security]
  dependencies:
    - skill_id: sandbox-runtime
      version_constraint: ">=1.0.0"
  security_level: medium
  trust_score: 0.95
  verified: true
```

---

## 四、数据模型

### 4.1 MCP Router 相关表

```sql
-- MCP 服务目录（路由池）
CREATE TABLE mcp_router_catalog (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       UUID NOT NULL UNIQUE,
    server_fk        FOREIGN KEY (server_id) REFERENCES mcp_servers(id) ON DELETE CASCADE,
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
    CONSTRAINT chk_transport_type CHECK (transport_type IN ('stdio', 'http')),
    CONSTRAINT chk_status CHECK (status IN ('active', 'disabled', 'maintenance')),
    CONSTRAINT chk_trust_score CHECK (trust_score >= 0 AND trust_score <= 1)
);

-- MCP 路由请求日志
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
    matched_server_fk   FOREIGN KEY (matched_server_id) REFERENCES mcp_router_catalog(id) ON DELETE SET NULL,
    match_score         DECIMAL(3,2),
    match_latency_ms    INTEGER,
    status              VARCHAR(50),
    error_code          VARCHAR(50),
    error_message       TEXT,
    duration_ms         INTEGER,
    created_at          TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_log_status CHECK (status IN ('success', 'failure', 'timeout'))
);

-- MCP 路由同步记录
CREATE TABLE mcp_router_sync_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       UUID NOT NULL,
    server_fk       FOREIGN KEY (server_id) REFERENCES mcp_servers(id) ON DELETE CASCADE,
    sync_type       VARCHAR(50) NOT NULL,
    old_status      VARCHAR(50),
    new_status      VARCHAR(50),
    approved_by     UUID,
    approved_by_fk FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL,
    approved_at     TIMESTAMP,
    note            TEXT,
    created_at      TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_sync_type CHECK (sync_type IN ('approved_sync', 'status_change', 'removal', 'catalog_update'))
);
```

### 4.2 Skill Repository 相关表

```sql
-- Skills 表扩展字段
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
ALTER TABLE skills ADD CONSTRAINT chk_trust_score CHECK (trust_score >= 0 AND trust_score <= 1);

-- Skill 依赖关系
CREATE TABLE skill_dependencies (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id            UUID NOT NULL,
    skill_fk            FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    dependency_skill_id UUID NOT NULL,
    dependency_fk        FOREIGN KEY (dependency_skill_id) REFERENCES skills(id) ON DELETE RESTRICT,
    version_constraint   VARCHAR(100) NOT NULL,
    created_at          TIMESTAMP DEFAULT NOW(),
    UNIQUE(skill_id, dependency_skill_id),
    CONSTRAINT chk_no_self_dependency CHECK (skill_id != dependency_skill_id)
);

-- Skill 版本历史扩展
ALTER TABLE skill_versions ADD COLUMN IF NOT EXISTS changelog TEXT;
ALTER TABLE skill_versions ADD COLUMN IF NOT EXISTS sop_version VARCHAR(50);
ALTER TABLE skill_versions ADD COLUMN IF NOT EXISTS approved_by UUID;
ALTER TABLE skill_versions ADD CONSTRAINT fk_skill_versions_approved_by FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL;

-- Skill 评分
CREATE TABLE skill_ratings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id    UUID NOT NULL,
    skill_fk    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL,
    user_fk     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    rating      INTEGER NOT NULL,
    comment     TEXT,
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW(),
    UNIQUE(skill_id, user_id),
    CONSTRAINT chk_rating CHECK (rating >= 1 AND rating <= 5)
);

-- Skill 安装记录
CREATE TABLE skill_installs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id    UUID NOT NULL,
    skill_fk    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    user_id     UUID,
    user_fk     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    version     VARCHAR(50) NOT NULL,
    environment VARCHAR(50) DEFAULT 'production',
    installed_at TIMESTAMP DEFAULT NOW(),
    skill_snapshot JSONB,
    CONSTRAINT chk_environment CHECK (environment IN ('production', 'staging', 'development'))
);

-- Skill 发布审批记录
CREATE TABLE skill_publish_approvals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id        UUID NOT NULL,
    skill_fk        FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    version         VARCHAR(50) NOT NULL,
    status          VARCHAR(50) DEFAULT 'pending',
    submitted_by    UUID NOT NULL,
    submitted_by_fk FOREIGN KEY (submitted_by) REFERENCES users(id) ON DELETE CASCADE,
    submitted_at    TIMESTAMP DEFAULT NOW(),
    reviewed_by     UUID,
    reviewed_by_fk  FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at     TIMESTAMP,
    review_note     TEXT,
    compliance_check_passed BOOLEAN DEFAULT false,
    compliance_check_note   TEXT,
    created_at      TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_publish_status CHECK (status IN ('pending', 'approved', 'rejected'))
);
```

---

## 五、工作流程

### 5.1 MCP Server 审批 → 自动同步到路由池

```
┌─────────────┐    approve    ┌──────────────┐    sync    ┌────────────────┐
│ mcp_servers │─────────────►│ sync_log     │───────────►│ mcp_router_    │
│ (status=    │              │ (记录同步)    │            │ catalog        │
│  approved)  │              │              │            │ (路由池)       │
└─────────────┘              └──────────────┘            └────────────────┘
```

**同步逻辑**：
1. MCP Server 审批通过
2. 写入 `mcp_router_sync_log` 记录
3. 从 `mcp_servers` 读取完整信息
4. 写入 `mcp_router_catalog`

### 5.2 MCP 路由请求流程

```
Agent 请求 → Ingress Handler → Task Parser → Match Engine → Egress Proxy → 响应
```

1. **Ingress Handler**: 解析 X-Agent-ID, X-Correlation-ID，提取 task 信息
2. **Task Parser**: 解析 structured 字段，降级时用 LLM 解析 raw description
3. **Match Engine**: 精确匹配 task_types + 标签过滤 + trust_score 排序
4. **Egress Proxy**: 根据 transport_type 转发到目标 MCP Server

### 5.3 Skill 生命周期

```
草稿来源: 用户创建 | 用户上传 | CLI 工具 | AI Agent
                ↓
           draft (草稿)
                ↓
    ┌───────────┼───────────┐
    ↓           ↓           ↓
  submit    discard     save as
    ↓        (delete)    version
pending_approval           ↓
    ↓                 revision
┌───┼───┐                │
↓   ↓   ↓                │
reject approve           │
    │    ↓                │
    │  published ◄────────┘
    │       (submit revision)
    └──► rejected
```

### 5.4 Skill 草稿来源

```go
const (
    SkillDraftSourceManual   = "manual"     // 用户手动创建
    SkillDraftSourceUpload  = "upload"     // 用户上传
    SkillDraftSourceCLI     = "cli"        // CLI 工具生成
    SkillDraftSourceAgent   = "agent"      // AI Agent 生成
)
```

---

## 六、前端改动

### 6.1 MCP Router 页面

```
/apps/web/app/
├── mcp-router/
│   ├── page.tsx           # MCP 路由概览仪表盘
│   ├── catalog/
│   │   └── page.tsx       # 路由目录管理
│   ├── metrics/
│   │   └── page.tsx       # Prometheus Metrics 监控
│   └── logs/
│       └── page.tsx        # 路由请求日志
```

### 6.2 Skill 仓库页面

```
/apps/web/app/
├── skills/
│   ├── page.tsx           # Skill 列表（保持现有）
│   ├── [id]/
│   │   └── page.tsx       # Skill 详情（保持现有）
│   ├── market/            # Skill 市场
│   │   └── page.tsx
│   ├── my-skills/         # 我的 Skills
│   │   └── page.tsx
│   ├── dependencies/       # 依赖管理
│   │   └── page.tsx
│   └── new/
│       └── page.tsx       # 创建 Skill
```

---

## 七、Prometheus Metrics

```bash
# HELP mcp_router_requests_total Total MCP router requests
# TYPE mcp_router_requests_total counter
mcp_router_requests_total{agent="openclaw-v1",task_type="code_review",status="success"} 1523

# HELP mcp_router_request_duration_seconds MCP router request latency
# TYPE mcp_router_request_duration_seconds histogram
mcp_router_request_duration_seconds_bucket{le="0.01"} 1200

# HELP mcp_router_match_failures_total Route match failures
# TYPE mcp_router_match_failures_total counter
mcp_router_match_failures_total{reason="no_matching_server"} 23
```

---

## 八、ER 关系图

```
┌─────────────────┐       ┌─────────────────────────┐
│     users       │       │      mcp_servers        │
├─────────────────┤       ├─────────────────────────┤
│ id (PK)         │       │ id (PK)                 │
│ username        │       │ name                    │
└────────┬────────┘       │ status                  │
         │               └───────────┬─────────────┘
         │                           │
         │               ┌───────────▼─────────────┐
         │               │   mcp_router_catalog     │
         │               ├─────────────────────────┤
         │               │ id (PK)                 │
         └───────────────│ server_id (FK, UNIQUE) │
         │               │ task_types[]            │
         │               │ tags[]                  │
         │               │ status                  │
         │               │ trust_score             │
         │               └───────────┬─────────────┘
         │                           │
         │               ┌───────────▼─────────────┐
         │               │    mcp_router_logs       │
         │               ├───────────────────────────│
         │               │ id (PK)                  │
         │               │ matched_server_id (FK)   │
         └───────────────│ agent_id                 │
                         │ status                   │
                         └─────────────────────────┘

┌─────────────────┐       ┌─────────────────────────┐
│     skills      │       │   skill_dependencies    │
├─────────────────┤       ├───────────────────────────│
│ id (PK)         │◄──────│ skill_id (FK)           │
│ name            │       │ dependency_skill_id(FK) │
│ version         │       └───────────┬──────────────┘
│ category        │                   │
│ trust_score     │       ┌───────────▼──────────────┐
│ verified        │       │   skill_ratings          │
│ sop_reference   │       ├──────────────────────────┤
│ published_by(FK)│       │ id (PK)                 │
└────────┬────────┘       │ skill_id (FK)           │
         │               │ user_id (FK)             │
         │               │ rating                   │
         │               └─────────────────────────┘
         │
         │               ┌─────────────────────────┐
         │               │   skill_installs        │
         │               ├─────────────────────────┤
         └───────────────│ id (PK)                 │
                         │ skill_id (FK)           │
                         │ user_id (FK)            │
                         │ environment             │
                         └─────────────────────────┘
```

---

## 九、待定项

| 模块 | 状态 |
|-----|------|
| Plugin Hub | 规划中 |
| Memory Hub | 规划中 |
| MCP 多版本同类型共存 | 不支持（企业应用追求可靠性和明确性） |

---

**文档信息**：
- 创建日期：2026-04-08
- 版本：v1.0
- 状态：已评审
