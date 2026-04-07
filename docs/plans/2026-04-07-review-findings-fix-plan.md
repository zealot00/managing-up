# Review Findings Fix Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix all critical bugs identified in code review: Redis memory leak, Worker blocking, Embedding evaluator, Fallback router, Pricing performance, and MCP critical issues.

**Architecture:** 
- **Phase 1 (Critical):** MCP bugs (Context cancellation, Fat Lock) - these can crash the system
- **Phase 2 (High):** Redis leak, Worker blocking, Embedding evaluator with env config
- **Phase 3 (Medium):** Fallback router, Pricing O(N), Stdio orphan process

**Tech Stack:** Go 1.21+, Redis, MCP protocol

---

## Phase 1: MCP Critical Bugs

### Task 1: Fix Context Cancellation Bug in RegisterHTTP

**Files:**
- Modify: `apps/api/internal/engine/executors/mcp_client.go:194-269`

**Step 1: Read the current implementation**
```bash
# Verify current bug location
```

**Step 2: Fix the context timeout issue**

The bug is that `defer cancel()` fires immediately when `RegisterHTTP` returns, but the MCP client background goroutines still need the context. The fix: don't use `defer cancel()` for the timeout context - instead, use a separate context for initialization only, and let the client use its own internal context.

```go
// OLD (buggy):
if defaultTimeout := 30 * time.Second; true {
    var cancel context.CancelFunc
    ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
    defer cancel()  // ❌ Cancels when function returns!
}

// NEW (fixed):
initCtx, initCancel := context.WithTimeout(ctx, 30*time.Second)
defer initCancel()  // ✓ Only cancels after init operations complete
```

**Step 3: Verify fix**
```bash
cd /Users/zealot/Code/skill-hub-ee/apps/api && go build ./...
```

---

### Task 2: Fix Context Cancellation Bug in Register (Stdio)

**Files:**
- Modify: `apps/api/internal/engine/executors/mcp_client.go:118-192`

**Step 1: Apply same fix pattern**

```go
// OLD (line 132-136):
if config.Timeout > 0 {
    var cancel context.CancelFunc
    ctx, cancel = context.WithTimeout(ctx, config.Timeout)
    defer cancel()  // ❌ Same bug
}

// NEW:
if config.Timeout > 0 {
    initCtx, initCancel := context.WithTimeout(ctx, config.Timeout)
    defer initCancel()
    ctx = initCtx  // Use initCtx for Start/Initialize/ListTools
}
```

---

### Task 3: Fix Fat Lock in MCP Registration (Two-Phase)

**Files:**
- Modify: `apps/api/internal/engine/executors/mcp_client.go:118-192`, `200-269`

**Step 1: Implement two-phase registration**

**Phase 1 - Without lock:** Do network I/O (Start, Initialize, ListTools)
**Phase 2 - With lock:** Add to clients map

```go
func (m *MCPClients) Register(ctx context.Context, name string, config MCPClientConfig) error {
    // Validate BEFORE acquiring lock
    if err := validateStdioConfig(config); err != nil {
        return fmt.Errorf("invalid stdio config: %w", err)
    }

    // Check existence WITHOUT lock first (read-only check)
    m.mu.RLock()
    if _, exists := m.clients[name]; exists {
        m.mu.RUnlock()
        return fmt.Errorf("MCP client already registered: %s", name)
    }
    m.mu.RUnlock()

    // Network I/O WITHOUT holding the write lock
    if config.Timeout > 0 {
        initCtx, initCancel := context.WithTimeout(ctx, config.Timeout)
        defer initCancel()
        ctx = initCtx
    }

    // ... (Start, Initialize, ListTools - all outside lock) ...

    // NOW acquire lock only to insert into map (nanoseconds)
    m.mu.Lock()
    if _, exists := m.clients[name]; exists {
        m.mu.Unlock()
        cleanup()
        return fmt.Errorf("MCP client already registered: %s", name)
    }
    m.clients[name] = &MCPClient{...}
    m.mu.Unlock()

    return nil
}
```

---

### Task 4: Fix Context Cancellation in Validation Functions

**Files:**
- Modify: `apps/api/internal/engine/executors/mcp_registry.go:189-260`, `262-338`

**Step 1: Apply same timeout context fix to validation functions**

Apply the same pattern to:
- `validateStdioServer` (lines 204-208)
- `validateHTTPServer` (lines 278-279)

---

## Phase 2: High Priority Fixes

### Task 5: Fix Redis Rate Limiter Memory Leak

**Files:**
- Modify: `apps/api/internal/gateway/ratelimit_redis.go:11-36`

**Step 1: Add EX/PX expiration to :reset key**

```lua
-- OLD (line 20):
redis.call('SET', key .. ':reset', now + window)

-- NEW:
redis.call('SET', key .. ':reset', now + window, 'PX', window)
```

Also fix line 32 where the same issue exists.

---

### Task 6: Fix Worker Queue Head-of-Line Blocking

**Files:**
- Modify: `apps/api/internal/engine/worker.go:49-67`

**Step 1: Make processPending asynchronous**

```go
func (w *Worker) processPending() {
    executions := w.repo.ListPendingExecutions()
    for _, exec := range executions {
        w.logger.Info("processing pending execution", "execution_id", exec.ID)
        // Spawn goroutine for each execution - non-blocking
        go func(exec server.Execution) {
            if err := w.engine.Run(context.Background(), exec.ID); err != nil {
                w.logger.Error("failed to run execution", "execution_id", exec.ID, "error", err)
            }
        }(exec)
    }
}
```

**Note:** This is a simple fix. For production, consider using a worker pool with bounded concurrency.

---

### Task 7: Implement Embedding Evaluator with Env Config

**Files:**
- Modify: `apps/api/internal/evaluator/metric.go:144-147`
- Modify: `apps/api/internal/evaluator/evaluator.go` (if exists, or create)
- Modify: `apps/api/.env.example` (add EMBEDDING_ vars)

**Step 1: Add environment variables**

```
EMBEDDING_PROVIDER=openai|ollama|azure|...
EMBEDDING_MODEL=text-embedding-3-small
EMBEDDING_API_KEY=sk-...
EMBEDDING_BASE_URL=https://api.openai.com/v1
```

**Step 2: Implement getEmbedding**

```go
func (e *EmbeddingSimilarityEvaluator) getEmbedding(ctx context.Context, text string) ([]float64, error) {
    if e.client == nil {
        return nil, nil
    }
    
    resp, err := e.client.CreateEmbeddings(ctx, &llm.EmbeddingRequest{
        Model: e.model,
        Input: text,
    })
    if err != nil {
        return nil, err
    }
    
    if len(resp.Data) == 0 {
        return nil, errors.New("no embedding returned")
    }
    
    return resp.Data[0].Embedding, nil
}
```

**Step 3: Verify client is initialized from env**

Check where `EmbeddingSimilarityEvaluator` is constructed and ensure it picks up env config.

---

## Phase 3: Medium Priority Fixes

### Task 8: Fix Fallback Router Rotation Logic

**Files:**
- Modify: `apps/api/internal/gateway/llm_router.go:79-85`

**Step 1: Implement proper rotation on failure**

```go
func (r *FallbackRouter) RecordFailure(provider llm.Provider) {
    r.mu.Lock()
    defer r.mu.Unlock()

    key := string(provider)
    _ = r.circuitBreaker.RecordFailure(context.Background(), key)
    
    // Advance to next provider in priority order
    if r.currentIndex < len(r.providers)-1 {
        r.currentIndex++
    }
}
```

---

### Task 9: Fix Pricing O(N) Performance Issue

**Files:**
- Modify: `apps/api/internal/gateway/pricing.go:37-53`

**Step 1: Use lowercase map for O(1) lookup**

```go
// Create lowercase lookup map at init
var modelPricingLower = make(map[string]ModelPricing, len(modelPricing))
for k, v := range modelPricing {
    modelPricingLower[strings.ToLower(k)] = v
}

func GetModelPricing(model string) ModelPricing {
    // Try exact match first
    if pricing, ok := modelPricing[model]; ok {
        return pricing
    }
    
    // Try lowercase lookup - O(1) not O(N)
    lowerModel := strings.ToLower(model)
    if pricing, ok := modelPricingLower[lowerModel]; ok {
        return pricing
    }
    
    return ModelPricing{
        InputCostPerToken:  0.000001,
        OutputCostPerToken: 0.000002,
    }
}
```

---

### Task 10: Fix Stdio Orphan Process (Zombie Prevention)

**Files:**
- Modify: `apps/api/internal/engine/executors/mcp_client.go:46-83`

**Step 1: Add SysProcAttr for process group cleanup**

```go
import (
    "syscall"
)

func validateStdioConfig(config MCPClientConfig) error {
    // ... existing validation ...
}

// Add process group handling in NewStdio transport setup
// Note: This requires modifying how transport.NewStdio is called
// The actual fix may need to be in the mcp-go library or we wrap the command creation

// Example wrapper (if we control the exec.Cmd creation):
cmd := exec.Command(config.Command, config.Args...)
cmd.SysProcAttr = &syscall.SysProcAttr{
    Pdeathsig: syscall.SIGKILL,
}
```

---

## Verification

After all tasks, run:

```bash
cd /Users/zealot/Code/skill-hub-ee/apps/api
go build ./...
go test ./internal/gateway/... ./internal/evaluator/... ./internal/engine/...
```

---

## Commit Strategy

Recommended commits:
1. `fix(mcp): prevent context cancellation and fat lock during registration`
2. `fix(gateway): add Redis :reset key expiration to prevent memory leak`
3. `fix(worker): process pending executions asynchronously`
4. `feat(evaluator): implement embedding similarity with env config`
5. `fix(router): advance currentIndex on failure for proper fallback`
6. `perf(pricing): use lowercase map for O(1) model lookup`
7. `fix(mcp): add process group cleanup for stdio transport`
