# Bridge Adapter 设计文档

**版本**：v1.0
**日期**：2026-04-23
**状态**：实现中

## 1. 概述

Bridge Adapter 是 Any-to-MCP 的桥接层，允许企业将现有的 REST API（通过 OpenAPI/Swagger 规范）快速适配为 MCP Server，无需编写额外代码。

## 2. 目标

1. **OpenAPI 导入** - 解析 Swagger/OpenAPI 2.0/3.0 规范
2. **REST → MCP Adapter** - 将 REST API 方法映射为 MCP Tool
3. **请求参数插值** - 支持路径参数、查询参数、body 插值
4. **响应裁剪** - 字段白名单、大小阈值裁剪、摘要生成

## 3. 架构

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  OpenAPI Spec   │────►│ Bridge Adapter  │────►│   MCP Server    │
│  (YAML/JSON)    │     │                 │     │   (stdio/http)  │
└─────────────────┘     │  - Parser       │     └─────────────────┘
                        │  - Template     │
                        │  - Response Opt │
                        └──────────────────┘
```

## 4. 核心组件

### 4.1 OpenAPI Parser

解析 OpenAPI 2.0/3.0 规范，提取：
- API 元信息（title, version, baseUrl）
- Endpoints（path, method, parameters, requestBody, responses）
- Schema 定义

### 4.2 Adapter Template

将 endpoint 映射为 MCP tool：

```yaml
tools:
  - name: get_user
    description: Get user by ID
    inputSchema:
      type: object
      properties:
        user_id:
          type: string
          description: User ID
    endpoint:
      path: /users/{user_id}
      method: GET
      parameters:
        - name: user_id
          in: path
          required: true
```

### 4.3 Request Interpolation

支持 `{{variable}}` 语法的参数插值：

```
GET /users/{{user_id}}?filter={{filter}}
```

### 4.4 Response Optimizer

- **字段白名单** - 只返回指定字段
- **大小阈值** - 超过 N bytes 则裁剪
- **摘要生成** - 调用 LLM 生成摘要（可选）

## 5. 数据模型

### 5.1 Adapter Template

```go
type AdapterTemplate struct {
    ID          string           `json:"id"`
    Name        string           `json:"name"`
    OpenAPISpec string           `json:"openapi_spec"`
    Mappings    []ToolMapping    `json:"mappings"`
    Options     AdapterOptions   `json:"options"`
    CreatedAt   time.Time        `json:"created_at"`
}

type ToolMapping struct {
    ToolName        string            `json:"tool_name"`
    OperationID     string            `json:"operation_id"`
    InputMappings   []FieldMapping    `json:"input_mappings"`
    OutputMappings   []FieldMapping    `json:"output_mappings"`
}

type AdapterOptions struct {
    ResponseLimitBytes int64  `json:"response_limit_bytes"` // 0 = unlimited
    UseSummary          bool   `json:"use_summary"`
    SummaryThresholdBytes int64 `json:"summary_threshold_bytes"`
}
```

## 6. API 设计

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/adapters` | 创建 Adapter |
| GET | `/api/v1/adapters` | 列出 Adapters |
| GET | `/api/v1/adapters/{id}` | 获取 Adapter |
| DELETE | `/api/v1/adapters/{id}` | 删除 Adapter |
| POST | `/api/v1/adapters/{id}/test` | 测试 Adapter |
| GET | `/api/v1/adapters/{id}/tools` | 列出生成的 MCP Tools |

## 7. 实现计划

### Phase 1: OpenAPI Parser
- 解析 OpenAPI 2.0 JSON/YAML
- 解析 OpenAPI 3.0 JSON/YAML
- 提取 endpoints 和 schemas

### Phase 2: Adapter Template Engine
- Template CRUD
- Tool mapping
- Request interpolation

### Phase 3: Response Optimizer
- 字段白名单过滤
- 大小阈值裁剪
- 摘要生成（可选）

## 8. 文件结构

```
apps/api/internal/bridge/
├── parser/
│   ├── openapi.go        # OpenAPI 解析器
│   └── schema.go         # Schema 转换
├── template/
│   ├── template.go       # Adapter Template
│   └── mapping.go        # Field mapping
├── optimizer/
│   ├── optimizer.go      # 响应优化器
│   └── summarizer.go     # 摘要生成（可选）
├── adapter.go            # 主 adapter
└── handler.go            # HTTP handler
```