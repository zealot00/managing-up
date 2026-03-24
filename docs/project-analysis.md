# Skill Hub EE - 项目分析报告

> 生成时间: 2026-03-20

---

## 一、项目概述

**项目名称**: managing-up (Enterprise Operating System for Intelligence)

**官方定位**: 企业运营知识作为可执行技能的管理平台

**实际定位**（推断）: AI Agent 的 Skill 分发平台（类似 "App Store"）

---

## 二、技术栈

| 层级 | 技术 | 完成度 |
|------|------|--------|
| 前端 | Next.js + TypeScript | ~4/10 |
| 后端 | Go HTTP API | ~10% |
| 数据库 | PostgreSQL (有迁移文件) | 骨架 |
| 缓存 | Redis (设计中有) | 未实现 |
| 文档 | 完整架构设计文档 | 完善 |

---

## 三、当前代码结构

### 3.1 前端 (apps/web)

| 页面 | 功能 | 完成度 |
|------|------|--------|
| `/` | Dashboard - 指标展示 + 最近执行 | 只读 |
| `/skills` | 技能列表 + 版本历史 | 只读 |
| `/executions` | 执行记录列表 | 只读 |
| `/approvals` | 审批列表 + 程序草稿 | 只读 |

**UI/UX 问题**:
- 无 SVG 图标
- 无 hover 交互反馈
- 无 loading skeleton
- 缺少 viewport meta / font preload
- 无 keyboard focus 样式
- 所有操作按钮均未实现（approve/reject 等）

### 3.2 后端 (apps/api)

| 模块 | 功能 | 完成度 |
|------|------|--------|
| `cmd/server` | HTTP 路由 + 处理器 | 基础 CRUD |
| `cmd/migrate` | 数据库迁移 | 完整 |
| `cmd/seed` | 种子数据 | 基础 |
| `internal/persistence/postgres` | PostgreSQL 持久化 | 骨架 |
| `internal/server/store.go` | 内存存储 | 完整 |

**API 端点**:
- `GET /api/v1/skills` - 技能列表
- `GET /api/v1/skill-versions` - 版本列表
- `GET /api/v1/executions` - 执行列表
- `GET /api/v1/approvals` - 审批列表
- `GET /api/v1/dashboard` - 仪表盘
- `POST /api/v1/executions/{id}/approve` - 审批

---

## 四、设计架构（6 个核心服务）

| 服务 | 设计功能 | 实现状态 |
|------|----------|----------|
| `ops-console` | 前端控制台 | ⚠️ 只读展示层 |
| `ingestion-service` | SOP 文档解析 + LLM 提取 | ❌ 未实现 |
| `composer-service` | 草稿 → 技能转换 | ❌ 未实现 |
| `registry-service` | 技能注册 + 版本管理 | ⚠️ 只有 CRUD |
| `runtime-service` | 执行引擎 + 步骤编排 | ❌ 未实现 |
| `tool-gateway` | 工具访问控制 | ❌ 未实现 |

---

## 五、推断的产品形态

### 5.1 核心价值

**AI Agent 的 "App Store"** - 企业内部 AI Agent 获取可信技能的市场

```
AI Agent                    Skill Hub                    人类
   │                            │                          │
   │─── 查询可用 Skills ────────▶│                          │
   │◀─── 返回 Skill 列表 ────────│                          │
   │                            │                          │
   │─── 下载 Skill 定义 ────────▶│                          │
   │◀─── 返回 Skill Spec ───────│                          │
   │                            │                          │
   │─── 执行 Skill ──────────────▶│                          │
   │                            │                          │
   │                            │─── 高风险操作 ────────────▶│
   │                            │◀─── 审批决定 ─────────────│
   │                            │                          │
   │◀─── 执行结果 ───────────────│                          │
```

### 5.2 设计的核心流程

```
SOP 文档 / 手动创建
       │
       ▼
  [生产] Skill 批量创建/导入
       │
       ▼
  [检测] 沙箱执行测试可用性
       │
       ▼
  [评分] 质量/风险/稳定性评级
       │
       ▼
  [分发] Agent 发现 + 订阅 + 下载
       │
       ▼
  [审计] 执行记录 + 人工审批
```

### 5.3 Meta Skill 概念

平台需要有一个 root/meta skill，让 Agent 知道如何发现和使用其他 skill：

```
Agent: "我想用技能"
   │
   ▼
Meta Skill 返回: { registry_url, how_to_use, available_skills }
   │
   ▼
Agent 根据描述调用具体 skill
```

---

## 六、缺失的关键功能

### 6.1 执行引擎 (Runtime Service)

- 状态机实现
- 步骤编排
- 超时/重试机制
- Checkpoint 恢复
- 执行日志

### 6.2 工具网关 (Tool Gateway)

- Shell 沙箱
- SQL 操作防护
- 秘钥管理
- 操作审计

### 6.3 SOP 解析 (Ingestion Service)

- PDF/Word/Markdown 解析
- LLM 辅助提取
- 结构化草案生成

### 6.4 前端交互

- Skill 创建/编辑表单
- SOP 上传界面
- 审批 approve/reject 按钮
- 执行详情页
- 实时状态刷新
- Toast/Notification 错误提示

### 6.5 Agent 集成

- OpenAPI/Skill 发现协议
- SDK 或 Client 库
- Webhook 通知机制

---

## 七、工作量估算

让 1 个真实 Skill 跑通端到端：

| 模块 | 估算 |
|------|------|
| 执行引擎 | 2-3 周 |
| SOP 解析 + LLM | 2-3 周 |
| 前端交互 | 2 周 |
| 审批流 | 1-2 周 |
| 联调 + 测试 | 2 周 |
| **总计** | **9-13 周** |

---

## 八、可行性结论

| 维度 | 评估 |
|------|------|
| **技术可行性** | ✅ 可行，无硬伤 |
| **产品价值** | ✅ 企业 AI Agent 管控是真实需求 |
| **资源需求** | ⚠️ 需要 LLMOps 能力 + 安全沙箱经验 |
| **当前状态** | ❌ 架构设计完善，代码只有 5-10% |

---

## 九、建议

1. **明确产品方向** - 确认是 "AI Agent App Store" 还是 "企业 SOP 执行平台"，两者技术路径不同
2. **聚焦执行引擎** - 先让一个硬编码 Skill 能跑通，再逐步完善
3. **补充前端交互** - 审批按钮、执行详情、错误处理
4. **考虑 Agent 集成** - 定义 Skill 发现/调用协议

---

## 十、原始分析来源

- `README.md`
- `docs/mvp-architecture.md`
- `docs/system-topology-and-service-boundaries.md`
- 代码审查: `apps/web/app/` + `apps/api/`
