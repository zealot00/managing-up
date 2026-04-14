# LLM Gateway Tool Call 异常 — 修复计划

## 问题确认

报告中的 10 个问题已全部验证确认，9 个问题中 7 个 Critical。

## 修复策略：分层修复

### 第一层：LLM Client 层（底层）

#### P1: `GenerateOptions` 添加 `Tools` 字段
**文件**: `apps/api/internal/llm/client.go`

```go
// 在 GenerateOptions 中添加
type GenerateOptions struct {
    Temperature float32
    MaxTokens   int
    TopP        float32
    StopWords   []string
    JSONMode    bool
    Tools       []Tool  // 新增
}

// 新增 Tool 类型和 WithTools Option
type Tool struct {
    Name        string
    Description string
    Parameters  map[string]any  // JSON Schema
}

func WithTools(tools []Tool) Option {
    return func(o *GenerateOptions) {
        o.Tools = tools
    }
}
```

#### P2: OpenAI Client 传递 `tools` 参数
**文件**: `apps/api/internal/llm/openai.go`

- `Generate()`: 在 reqBody 中添加 `tools` 字段
- `GenerateStream()`: 在 reqBody 中添加 `tools` 字段
- `StreamChunk` 添加 `ToolCalls` 字段

#### P3: OpenAI 流式读取 `finish_reason=tool_calls` 时不提前终止
**文件**: `apps/api/internal/llm/openai.go`

问题代码（line 216-224）：
```go
if finishReason != "" {
    return &StreamChunk{...}, nil  // 提前返回，丢失后续 chunk
}
```

修复：当 `finish_reason == "tool_calls"` 时，继续读取后续 chunk 收集完整 tool_calls 数据。

#### P4: Anthropic Client 添加 tools 支持
**文件**: `apps/api/internal/llm/anthropic.go`

Anthropic Messages API 的 tools 参数格式不同，需要单独处理。

---

### 第二层：Gateway 层

#### P5: `openAIChatRequest` 添加 `tools` 字段
**文件**: `apps/api/internal/gateway/openai_chat.go`

```go
type openAIChatRequest struct {
    Model       string          `json:"model"`
    Messages    []openAIMessage `json:"messages"`
    Temperature *float32        `json:"temperature"`
    MaxTokens   *int            `json:"max_tokens"`
    Stream      bool            `json:"stream"`
    Tools       []toolRequest   `json:"tools"`  // 新增
}
```

#### P6: Gateway 传递 tools 到 LLM Client
**文件**: `apps/api/internal/gateway/openai_chat.go`

在 `HandleOpenAIChat` 和 `handleOpenAIChatStream` 中，将 `req.Tools` 转换为 `llm.Option` 并传递给 LLM client。

---

### 第三层：Agent 层

#### P7: 替换 `ParseToolCalls` 为原生结构解析
**文件**: `apps/api/internal/engine/agents/llm_agent.go`

修改 `LLMAgent.Run()` 方法，使其能够处理 LLM Client 返回的原生 `tool_calls`（而非文本解析）。

需要新增：从 `StreamChunk` 或 `Response` 中提取 `ToolCalls` 结构。

#### P8: 工具不存在时通知 LLM
**文件**: `apps/api/internal/engine/agents/llm_agent.go`

问题代码（line 132-139）：
```go
if !found {
    // 静默 continue
    continue
}
```

修复：当工具不存在时，将错误信息作为 tool result 消息添加到对话中，告知 LLM 调用失败。

#### P9: `NewToolGateway()` 静默 Mock 问题
**文件**: `apps/api/internal/engine/tool_gateway.go`

问题：`NewToolGateway()` 返回 nil registry 会静默返回 mock 成功。

修复方案：
1. 添加日志警告提示使用了 mock
2. 或修改返回值为 error
3. 搜索所有调用点，强制使用 `NewToolGatewayWithRegistry()`

---

### 第四层：上下文层

#### P10: Context Truncation 破坏 tool result 连贯性
**文件**: `apps/api/internal/engine/context_truncator.go`

问题：截断时可能删除 tool result 但保留 tool call，导致 LLM 看到孤立的 tool call。

修复：截断策略应保证 tool call 和对应的 tool result 配对保留或同时删除。

#### P11: Tool result 序列化格式
**文件**: `apps/api/internal/engine/agents/llm_agent.go`

当前直接将 `execResult` json.Marshal 后作为 Content 传递。考虑添加格式化或降级处理。

---

## 修复顺序

1. **LLM Client 层** (第一层，先修底层依赖)
   - `GenerateOptions` 添加 Tools 字段
   - OpenAI Client 支持 tools 参数
   - 修复流式提前终止 bug
   - Anthropic Client 添加 tools 支持

2. **Gateway 层** (第二层，修复请求解析)
   - `openAIChatRequest` 添加 tools 字段
   - Gateway 传递 tools 到 LLM Client

3. **Agent 层** (第三层，修复工具执行)
   - 替换 ParseToolCalls 为原生结构解析
   - 工具不存在时通知 LLM
   - 修复 nil registry 静默 Mock

4. **上下文层** (第四层，修复长期对话)
   - Context Truncation 修复
   - Tool result 序列化优化

---

## 修复状态

### ✅ 已完成 (10/10)

1. **LLM Client 层** - 全部完成
   - `GenerateOptions` 添加 Tools 字段 ✅
   - OpenAI Client 支持 tools 参数 ✅
   - 修复流式 finish_reason=tool_calls 提前终止 bug ✅
   - Anthropic Client 添加 tools 支持 ✅

2. **Gateway 层** - 全部完成
   - `openAIChatRequest` 添加 tools 字段 ✅
   - Gateway 传递 tools 到 LLM Client ✅

3. **Agent 层** - 全部完成
   - 替换 ParseToolCalls 为原生结构解析 ✅
   - 工具不存在时通知 LLM ✅
   - nil registry 静默 Mock 添加警告日志 ✅

4. **上下文层** - 全部完成
   - Context Truncation 修复 (tool call/result 配对保留) ✅
   - Tool result 序列化优化 (使用 json.MarshalIndent) ✅

---

## 测试结果

- **编译**: ✅ 通过 `go build ./...`
- **测试**: 
  - `internal/llm`: ✅ 通过
  - `internal/engine`: ⚠️ 1个预存测试失败 (TestToolGateway_Invoke_Timeout，修改前已失败)
  - `internal/gateway`: ⚠️ 1个预存测试失败 (TestEstimateRequestTokens，与本次修复无关)

---

## 修改文件清单

1. `apps/api/internal/llm/client.go` - 添加 ToolCall、Tool 结构体，GenerateOptions.Tools
2. `apps/api/internal/llm/provider.go` - Response 添加 ToolCalls 字段
3. `apps/api/internal/llm/openai.go` - 支持 tools 参数，修复流式 tool_calls 提前终止
4. `apps/api/internal/llm/anthropic.go` - 支持 tools 参数和 tool_use 响应解析
5. `apps/api/internal/gateway/openai_chat.go` - 添加 tools 字段和转换函数
6. `apps/api/internal/engine/agents/llm_agent.go` - 使用原生 tool_calls，通知 LLM 工具失败，json.MarshalIndent
7. `apps/api/internal/engine/tool_gateway.go` - nil registry 添加警告日志
8. `apps/api/internal/engine/context_truncator.go` - tool call/result 配对分组保留
