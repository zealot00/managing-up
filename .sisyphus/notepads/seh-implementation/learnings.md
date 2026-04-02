# SEH Implementation Learnings

## Overview
Implemented SEH (Skill Evaluation Harness) API in `apps/api/internal/seh/` package.

## Key Decisions

### Package Structure
- **types.go**: All DTOs matching the spec exactly - no custom fields added
- **store.go**: File-based JSON storage with atomic writes (write to tmp, rename)
- **auth.go**: Static token validation (mock_token_admin, mock_token_reviewer, mock_token_approver)
- **handlers.go**: HTTP handlers for all 9 P0 endpoints

### Route Registration
Added SEH routes in `server.go` after orchestrator routes:
```go
sehAuth := seh.AuthMiddleware()
mux.Handle("/v1/seh/", sehAuth(http.StripPrefix("/v1/seh", srv.sehServer)))
```

### Static Tokens
Implemented static token validation per spec section 3.2:
- mock_token_admin → role: "admin"
- mock_token_reviewer → role: "reviewer"  
- mock_token_approver → role: "approver"

### Mock Data Initialization
Store auto-initializes mock data on first access:
- 2 datasets (ds_abc12345, ds_def67890)
- 10 evaluation cases (5 per dataset)
- 3 runs (high/med/low scores)
- 2 policies (strict/relaxed)

### Gate Evaluation
Implemented policy-based gate evaluation with:
- min_golden_weight check
- min_source_diversity check
- per-source success rate thresholds
- require_provenance check
- min_success_rate (0.8) and min_score (0.75) defaults

## Files Created
```
apps/api/internal/seh/
├── types.go       # All DTO types
├── store.go       # JSON file storage
├── auth.go        # Token authentication
└── handlers.go    # HTTP handlers
```

## Files Modified
- `apps/api/internal/server/server.go` - Added SEH routes and sehServer field

## Build Verification
```bash
cd /Users/zealot/Code/skill-hub-ee/apps/api && go build ./...
# Success - no errors
```

## P0 Endpoints Implemented
1. POST /auth/token - Auth换token
2. GET /datasets - 数据集列表
3. GET /datasets/:dataset_id - 数据集详情
4. GET /datasets/:dataset_id/cases - 下载用例
5. GET /datasets/:dataset_id/verify - 校验完整性
6. POST /runs - 上报run
7. GET /runs/:run_id - 查询单次run
8. GET /runs - run列表
9. POST /runs/:run_id/gate - 门禁评估
