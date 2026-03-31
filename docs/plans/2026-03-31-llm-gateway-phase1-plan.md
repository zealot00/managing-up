# Phase 1: LLM Gateway 实现计划

## 概述

基于 `docs/plans/2026-03-31-llm-gateway-design.md` 设计，实现 LLM Gateway 核心功能。

---

## 任务列表

### 1. 设置 LLM Gateway 路由结构

- [ ] 创建 `apps/api/internal/gateway/` 目录
- [ ] 创建 `apps/api/internal/gateway/router.go` - 路由分发器
- [ ] 创建 `apps/api/internal/gateway/openai.go` - OpenAI 兼容端点
- [ ] 创建 `apps/api/internal/gateway/anthropic.go` - Anthropic 兼容端点
- [ ] 创建 `apps/api/internal/gateway/provider.go` - Provider 检测与路由

**文件位置:** `apps/api/internal/gateway/`

---

### 2. 实现 OpenAI /v1/chat/completions

- [ ] 创建 `POST /v1/chat/completions` 处理器
- [ ] 实现请求解析 (model, messages, temperature, max_tokens, stream)
- [ ] 实现 OpenAI 标准响应格式
- [ ] 实现 stream 模式 (可选)

**文件位置:** `apps/api/internal/gateway/openai_chat.go`

**验证:**
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <key>" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hello"}]}'
```

---

### 3. 实现 Anthropic /v1/messages

- [ ] 创建 `POST /v1/messages` 处理器
- [ ] 实现 Anthropic 风格请求解析
- [ ] 实现 Anthropic 响应格式
- [ ] 处理 `anthropic-version` header

**文件位置:** `apps/api/internal/gateway/anthropic_messages.go`

**验证:**
```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: <key>" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model":"claude-sonnet-4","messages":[{"role":"user","content":"Hello"}],"max_tokens":1024}'
```

---

### 4. 实现 Provider 检测与路由

- [ ] 实现 model 前缀解析 (openai:, anthropic:, etc.)
- [ ] 实现 API Key 透明传递
- [ ] 实现跨 Provider 错误映射
- [ ] 添加 provider 前缀支持: `openai:gpt-4o`, `anthropic:claude-sonnet-4`

**文件位置:** `apps/api/internal/gateway/provider.go`

**Provider 映射:**

| 前缀 | Provider |
|------|----------|
| `gpt-*`, `o1-*` | OpenAI |
| `claude-*` | Anthropic |
| `gemini-*` | Google |
| `deepseek-*` | DeepSeek |
| `glm-*` | ZhipuAI |
| `llama*`, `mistral*`, `qwen2*` | Ollama |

---

### 5. 添加 API Key 透明传递中间件

- [ ] 创建 auth 中间件 `apps/api/internal/gateway/auth.go`
- [ ] 支持 `Authorization: Bearer <key>` 头部
- [ ] 支持 `x-api-key` 头部 (Anthropic 风格)
- [ ] 实现可选认证接口白名单 (health, /v1/models)

**文件位置:** `apps/api/internal/gateway/auth.go`

---

### 6. 注册路由到 Server

- [ ] 在 `server.go` 中注册 `/v1/*` 路由
- [ ] 添加 `/health` 公开端点
- [ ] 确保现有 `/api/v1/*` 不受影响

**文件位置:** `apps/api/internal/server/server.go`

---

## 依赖关系

```
1. 设置路由结构
   ↓
2. OpenAI 端点 (依赖 1)
   ↓
3. Anthropic 端点 (依赖 1)
   ↓
4. Provider 路由 (依赖 2, 3)
   ↓
5. Auth 中间件 (依赖 4)
   ↓
6. 注册路由 (依赖 1, 5)
```

---

## 测试计划

### 单元测试

| 文件 | 测试内容 |
|------|----------|
| `gateway/provider_test.go` | Provider 检测逻辑 |
| `gateway/openai_test.go` | OpenAI 请求/响应解析 |
| `gateway/anthropic_test.go` | Anthropic 请求/响应解析 |
| `gateway/auth_test.go` | 认证中间件 |

### 集成测试

```bash
# OpenAI 兼容测试
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"Hi"}]}'

# Anthropic 兼容测试
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: sk-ant-xxx" \
  -d '{"model":"claude-haiku-3","messages":[{"role":"user","content":"Hi"}]}'
```

---

## 预计文件变更

```
apps/api/internal/gateway/
├── router.go           (NEW)
├── openai_chat.go     (NEW)
├── anthropic_messages.go (NEW)
├── provider.go        (NEW)
├── auth.go            (NEW)
├── router_test.go     (NEW)
├── openai_test.go     (NEW)
├── anthropic_test.go  (NEW)
├── provider_test.go   (NEW)
└── auth_test.go       (NEW)

apps/api/internal/server/
└── server.go          (MODIFIED - 添加路由注册)
```

---

## 风险与注意事项

1. **API Key 安全**: 确保 Key 不被日志记录
2. **错误处理**: 统一错误格式，避免暴露 Provider 内部错误
3. **向后兼容**: 确保现有 `/api/v1/*` 不受影响
