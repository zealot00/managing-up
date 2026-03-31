# managing-up API Server 重构设计方案

## 1. 目标

将 managing-up 打造成 **LLM Gateway + Skill 执行引擎**，作为多应用的统一入口。

### 核心能力

| 模块 | 优先级 | 说明 |
|------|--------|------|
| LLM Gateway | 🔴 最高 | OpenAI/Anthropic 接口兼容 |
| Auth | 🔴 最高 | API Key + OAuth 2.0 |
| Skill 管理与执行 | 🟡 后续 | 等 sop-to-skill 重构后统一 |

---

## 2. 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                    managing-up API Server                     │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              LLM Gateway (核心)                     │   │
│  │  - OpenAI 兼容 (/v1/chat/completions)               │   │
│  │  - Anthropic 兼容 (/v1/messages)                   │   │
│  │  - 多 provider 路由                                  │   │
│  │  - 调用方 API Key 透明传递                           │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Auth 模块                              │   │
│  │  - API Key 认证 (Bearer Token)                     │   │
│  │  - OAuth 2.0 (授权码流程)                           │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │          Skill 引擎 (后续)                          │   │
│  │  - 执行引擎                                          │   │
│  │  - 动态决策 (LLM 同步调用)                          │   │
│  │  - Session 后台运行机制                              │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. LLM Gateway 接口设计

### 3.1 OpenAI 兼容

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/chat/completions` | 聊天补全 |
| POST | `/v1/completions` | 文本补全 |
| POST | `/v1/embeddings` | 向量嵌入 |

**Chat Completions 请求：**
```json
{
  "model": "gpt-4o",
  "messages": [
    {"role": "system", "content": "..."},
    {"role": "user", "content": "..."}
  ],
  "temperature": 0.7,
  "max_tokens": 1000,
  "stream": false
}
```

**响应格式：** OpenAI 标准格式

### 3.2 Anthropic 兼容

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/messages` | 消息创建 |

**Messages 请求：**
```json
{
  "model": "claude-sonnet-4",
  "messages": [
    {"role": "user", "content": "..."}
  ],
  "max_tokens": 1024
}
```

### 3.3 Provider 识别

**方式：** 根据 model 前缀自动路由

| 前缀 | Provider |
|------|----------|
| `gpt-*`, `o1-*` | OpenAI |
| `claude-*` | Anthropic |
| `gemini-*` | Google |
| `deepseek-*` | DeepSeek |
| `glm-*` | ZhipuAI |
| `ernie-*` | Baidu |
| `qwen-*` | Alibaba |
| `abab*`, `MiniMax-*` | Minimax |
| `llama*`, `mistral*`, `qwen2*` | Ollama (local) |

---

## 4. Auth 模块设计

### 4.1 API Key 认证

**Header:** `Authorization: Bearer <api_key>`

**流程：**
```
请求 → 解析 Header → 透明传递 API Key → 调用 LLM → 返回响应
```

**说明：** managing-up 不存储 Key，仅透传

### 4.2 OAuth 2.0 流程

**端点：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/auth/authorize` | 授权页面 |
| POST | `/auth/token` | 获取 Token |
| POST | `/auth/refresh` | 刷新 Token |
| POST | `/auth/revoke` | 撤销 Token |

**支持的 Grant Types：**
- `authorization_code` - 授权码
- `client_credentials` - 客户端凭证
- `refresh_token` - 刷新 Token

### 4.3 认证优先级

1. **API Key** - Bearer Token 优先
2. **OAuth Token** - 如果无 API Key，验证 OAuth Token
3. **可选认证** - 部分公开接口（如 health）无需认证

---

## 5. Session 机制 (Skill 执行)

### 5.1 同步执行

```
请求 → LLM 同步调用 → 返回结果
```

### 5.2 后台执行 (Session)

**创建 Session：**
```json
POST /api/v1/sessions
{
  "skill_id": "xxx",
  "input": {...},
  "background": true
}

Response:
{
  "session_id": "sess_xxx",
  "status": "running"
}
```

**查询 Session：**
```json
GET /api/v1/sessions/{session_id}

Response:
{
  "session_id": "sess_xxx",
  "status": "completed| running| failed",
  "result": {...}
}
```

### 5.3 Skill 执行中的 LLM 调用

```
Skill Step (type: condition)
    │
    ▼
LLM 同步评估
    │
    ├── 条件满足 → 继续下一步
    │
    └── 条件不满足 → 执行 on_failure
```

---

## 6. 多 Provider 配置

### 6.1 调用方指定

**方式：** model 字段包含 provider 前缀

```
model: "openai:gpt-4o"
model: "anthropic:claude-sonnet-4"
```

### 6.2 调用方 API Key 传递

**Header：**
```
OpenAI: x-api-key: sk-xxx
Anthropic: x-api-key: sk-ant-xxx
```

### 6.3 多租户隔离

**策略：** API Key 隔离

- managing-up 不管理配额
- 每个调用方用自己的 Key
- 仅做路由和透明代理

---

## 7. 错误处理

### 7.1 统一错误格式

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "...",
    "param": "model",
    "type": "invalid_request_error"
  }
}
```

### 7.2 Provider 错误映射

| Provider 错误 | 映射为 |
|---------------|--------|
| 401 | `invalid_request_error` |
| 429 | `rate_limit_error` |
| 500 | `internal_error` |
| 503 | `service_unavailable_error` |

---

## 8. 配置项

### 8.1 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `PORT` | 服务端口 | `8080` |
| `LLM_DEFAULT_PROVIDER` | 默认 Provider | `openai` |
| `LLM_DEFAULT_MODEL` | 默认 Model | `gpt-4o` |
| `OAUTH_CLIENT_ID` | OAuth 客户端 ID | - |
| `OAUTH_CLIENT_SECRET` | OAuth 客户端密钥 | - |
| `OAUTH_REDIRECT_URI` | OAuth 回调地址 | - |

### 8.2 配置文件 (config.yaml)

```yaml
server:
  port: 8080

llm:
  default_provider: openai
  default_model: gpt-4o

auth:
  oauth:
    enabled: true
    client_id: ${OAUTH_CLIENT_ID}
    client_secret: ${OAUTH_CLIENT_SECRET}
    redirect_uri: ${OAUTH_REDIRECT_URI}
```

---

## 9. 实现计划

### Phase 1: LLM Gateway 核心

- [ ] OpenAI /v1/chat/completions 实现
- [ ] Anthropic /v1/messages 实现
- [ ] Provider 路由逻辑
- [ ] API Key 透明传递

### Phase 2: Auth 模块

- [ ] API Key 中间件
- [ ] OAuth 2.0 授权码流程
- [ ] Token 验证

### Phase 3: Session 机制

- [ ] Session 创建与管理
- [ ] 后台执行支持
- [ ] 轮询查询接口

### Phase 4: Skill 引擎 (后续)

- [ ] Skill 注册与执行
- [ ] LLM 动态决策
- [ ] sop-to-skill 集成

---

## 10. API 端点汇总

### LLM Gateway

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/chat/completions` | OpenAI 聊天 |
| POST | `/v1/completions` | OpenAI 补全 |
| POST | `/v1/embeddings` | OpenAI 嵌入 |
| POST | `/v1/messages` | Anthropic 消息 |
| GET | `/v1/models` | 模型列表 |

### Auth

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/auth/authorize` | 授权 |
| POST | `/auth/token` | Token |
| POST | `/auth/refresh` | 刷新 |
| POST | `/auth/revoke` | 撤销 |

### Session (Skill 执行)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/sessions` | 创建 |
| GET | `/api/v1/sessions/{id}` | 查询 |
| DELETE | `/api/v1/sessions/{id}` | 取消 |

### 公开

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/api/v1/meta` | 服务信息 |

---

## 11. 参考

- OpenAI API: https://platform.openai.com/docs/api-reference
- Anthropic API: https://docs.anthropic.com/en/api/reference
- OAuth 2.0: RFC 6749
