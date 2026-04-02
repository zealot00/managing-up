# SEH 评测模块 HTTP API 实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 根据 api-mock-integration.md 文档实现 SEH 评测模块缺失的 HTTP API 接口，将 Mock 数据存储升级为真实实现。

**Architecture:** SEH Server 使用文件存储 (Store) 作为后端，需要扩展 Store 方法和 Handlers 实现完整的 RESTful API。当前 handlers.go 使用 s.repo (PostgreSQL)，但 repo 可能为 nil。需要确保 Store 和 Repo 两套存储都能工作。

**Tech Stack:** Go HTTP Handlers, JSON File Store (store.go), PostgreSQL Repo (repo.go)

---

## 待实现接口清单

| 接口 | 优先级 | 当前状态 | 需要修改 |
|------|--------|----------|----------|
| `POST /datasets` | P2 | ❌ 未实现 | store.go + handlers.go |
| `DELETE /datasets/:dataset_id` | P2 | ❌ 未实现 | store.go + handlers.go |
| `POST /releases/:release_id/approve` | P2 | ⚠️ 硬编码返回 | store.go + handlers.go |
| `POST /releases/:release_id/reject` | P2 | ⚠️ 硬编码返回 | store.go + handlers.go |
| `POST /releases/:release_id/rollback` | P2 | ⚠️ 硬编码返回 | store.go + handlers.go |
| `GET /cases/:case_id/lineage` | P1 | ⚠️ 返回空 | store.go + handlers.go |
| `GET /datasets/:dataset_id/lineage` | P1 | ❌ 未实现 | store.go + handlers.go |
| `POST /policies` | P1 | ⚠️ 仅写入内存 | store.go + handlers.go |

---

## Task 1: 扩展 Store - Dataset 存储方法

**Files:**
- Modify: `apps/api/internal/seh/store.go:264-316`

**Step 1: 添加 CreateDataset 方法**

在 `store.go` 第 316 行后添加:

```go
func (s *Store) CreateDataset(dataset DatasetDetailDTO) error {
	datasets, err := s.ListDatasetsDetail()
	if err != nil {
		datasets = []DatasetDetailDTO{}
	}

	dataset.DatasetID = "ds_" + randomID()
	dataset.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	datasets = append(datasets, dataset)

	return s.writeJSON(DatasetsFile, datasets)
}

func (s *Store) ListDatasetsDetail() ([]DatasetDetailDTO, error) {
	var datasets []DatasetDetailDTO
	if err := s.readJSON(DatasetsFile, &datasets); err != nil {
		return nil, err
	}
	return datasets, nil
}

func (s *Store) DeleteDataset(datasetID string) error {
	datasets, err := s.ListDatasetsDetail()
	if err != nil {
		return err
	}

	for i, d := range datasets {
		if d.DatasetID == datasetID {
			datasets = append(datasets[:i], datasets[i+1:]...)
			return s.writeJSON(DatasetsFile, datasets)
		}
	}
	return fmt.Errorf("dataset not found: %s", datasetID)
}
```

**Step 2: 验证代码**

Run: `go build apps/api/...`
Expected: 无编译错误

---

## Task 2: 扩展 Store - Release 审批状态存储

**Files:**
- Modify: `apps/api/internal/seh/store.go`

**Step 1: 添加 UpdateReleaseWithApproval 方法**

在 `store.go` 末尾 (第 535 行后) 添加:

```go
func (s *Store) GetReleaseForUpdate(releaseID string) (map[string]interface{}, error) {
	releases, err := s.ListReleases()
	if err != nil {
		return nil, err
	}
	for _, r := range releases {
		if r["release_id"] == releaseID {
			return r, nil
		}
	}
	return nil, fmt.Errorf("release not found: %s", releaseID)
}

func (s *Store) ApproveRelease(releaseID, approvedBy string) error {
	releases, err := s.ListReleases()
	if err != nil {
		return err
	}
	for i, r := range releases {
		if r["release_id"] == releaseID {
			r["status"] = "approved"
			r["approved_by"] = approvedBy
			r["approved_at"] = time.Now().UTC().Format(time.RFC3339)
			releases[i] = r
			return s.writeJSON("releases.json", releases)
		}
	}
	return fmt.Errorf("release not found: %s", releaseID)
}

func (s *Store) RejectRelease(releaseID, rejectedReason string) error {
	releases, err := s.ListReleases()
	if err != nil {
		return err
	}
	for i, r := range releases {
		if r["release_id"] == releaseID {
			r["status"] = "rejected"
			r["rejected_reason"] = rejectedReason
			releases[i] = r
			return s.writeJSON("releases.json", releases)
		}
	}
	return fmt.Errorf("release not found: %s", releaseID)
}

func (s *Store) RollbackRelease(releaseID string) error {
	releases, err := s.ListReleases()
	if err != nil {
		return err
	}
	for i, r := range releases {
		if r["release_id"] == releaseID {
			r["status"] = "rolled_back"
			releases[i] = r
			return s.writeJSON("releases.json", releases)
		}
	}
	return fmt.Errorf("release not found: %s", releaseID)
}
```

**Step 2: 添加 Lineage 存储方法**

在上述方法后添加:

```go
func (s *Store) GetCaseLineage(caseID string) (map[string]interface{}, error) {
	// 从 lineage_cases.json 读取
	var lineage map[string]interface{}
	if err := s.readJSON("lineage_cases.json", &lineage); err != nil {
		// 如果文件不存在，返回空血缘
		return map[string]interface{}{
			"ancestors":    []interface{}{},
			"descendants":  []interface{}{},
		}, nil
	}
	return lineage, nil
}

func (s *Store) GetDatasetLineage(datasetID string) (map[string]interface{}, error) {
	// 从 lineage_datasets.json 读取
	var lineage map[string]interface{}
	if err := s.readJSON("lineage_datasets.json", &lineage); err != nil {
		// 如果文件不存在，返回空血缘
		return map[string]interface{}{
			"versions": []interface{}{},
		}, nil
	}
	return lineage, nil
}
```

**Step 3: 验证代码**

Run: `go build apps/api/...`
Expected: 无编译错误

---

## Task 3: 扩展 Store - Policy 持久化

**Files:**
- Modify: `apps/api/internal/seh/store.go`

**Step 1: 添加 CreatePolicy 方法**

在 `store.go` 第 415 行 (ListPolicies 方法后) 添加:

```go
func (s *Store) CreatePolicy(policy GovernancePolicyDTO) error {
	policies, err := s.ListPolicies()
	if err != nil {
		policies = []GovernancePolicyDTO{}
	}

	policy.PolicyID = "pol_" + randomID()
	policy.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	policies = append(policies, policy)

	return s.writeJSON(PoliciesFile, policies)
}
```

**Step 2: 验证代码**

Run: `go build apps/api/...`
Expected: 无编译错误

---

## Task 4: 实现 POST /datasets 和 DELETE /datasets/:dataset_id

**Files:**
- Modify: `apps/api/internal/seh/handlers.go:125-200`

**Step 1: 在 HandleDatasets 中添加 POST /datasets**

在 `handlers.go` 第 131 行 HandleDatasets 方法开始处添加:

```go
func (s *Server) HandleDatasets(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// POST /datasets - Create new dataset
	if r.Method == http.MethodPost && path == "/datasets" {
		var req struct {
			Name        string          `json:"name"`
			Version     string          `json:"version"`
			Owner       string          `json:"owner"`
			Description string          `json:"description"`
			Manifest    DatasetManifest `json:"manifest"`
			Tags        []string        `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid JSON", http.StatusBadRequest, "BAD_REQUEST")
			return
		}
		if req.Name == "" {
			writeError(w, "name is required", http.StatusUnprocessableEntity, "VALIDATION_ERROR")
			return
		}

		dataset := DatasetDetailDTO{
			Name:        req.Name,
			Version:     req.Version,
			Owner:       req.Owner,
			Description: req.Description,
			Manifest:    req.Manifest,
			CaseCount:   0,
			Checksum:    "",
			CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		}

		if s.store != nil {
			if err := s.store.CreateDataset(dataset); err != nil {
				writeError(w, "Failed to create dataset", http.StatusInternalServerError, "INTERNAL_ERROR")
				return
			}
		} else if s.repo != nil {
			// Use repo if available
			// TODO: implement CreateDataset in repo
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"dataset_id": dataset.DatasetID,
			"created_at": dataset.CreatedAt,
		})
		return
	}

	// DELETE /datasets/:dataset_id
	if r.Method == http.MethodDelete && strings.HasPrefix(path, "/datasets/") {
		datasetID := strings.TrimPrefix(path, "/datasets/")
		if datasetID == "" || strings.Contains(datasetID, "/") {
			writeError(w, "Invalid dataset ID", http.StatusBadRequest, "BAD_REQUEST")
			return
		}

		if s.store != nil {
			if err := s.store.DeleteDataset(datasetID); err != nil {
				writeError(w, "Dataset not found", http.StatusNotFound, "NOT_FOUND")
				return
			}
		} else if s.repo != nil {
			// Use repo if available
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}
	// ... existing GET logic continues
```

**Step 2: 添加 store 字段到 Server 结构**

修改 `handlers.go` 第 13-16 行:

```go
type Server struct {
	repo       *Repo
	store      *Store  // 添加: 文件存储 (mock 模式)
	authConfig AuthConfig
}
```

修改 `NewServer` 函数 (第 18-28 行):

```go
func NewServer(authConfig AuthConfig) *Server {
	dsn := os.Getenv("DATABASE_URL")
	var repo *Repo
	var store *Store

	if dsn != "" {
		var err error
		repo, err = NewRepo(dsn)
		if err != nil {
			panic(fmt.Sprintf("failed to connect to SEH database: %v", err))
		}
	} else {
		// 使用文件存储 (mock 模式)
		store = NewStore("")
		if err := store.InitMockData(); err != nil {
			panic(fmt.Sprintf("failed to init mock data: %v", err))
		}
	}
	return &Server{repo: repo, store: store, authConfig: authConfig}
}
```

修改 `NewServerWithRepo` 函数 (第 31-33 行):

```go
func NewServerWithRepo(repo *Repo, authConfig AuthConfig) *Server {
	return &Server{repo: repo, authConfig: authConfig}
}

func NewServerWithStore(store *Store, authConfig AuthConfig) *Server {
	return &Server{store: store, authConfig: authConfig}
}
```

**Step 3: 验证代码**

Run: `go build apps/api/...`
Expected: 无编译错误

---

## Task 5: 实现 Release 审批接口 (approve/reject/rollback)

**Files:**
- Modify: `apps/api/internal/seh/handlers.go:865-891`

**Step 1: 修改 HandleReleaseByID 中的 approve/reject/rollback**

替换 `handlers.go` 第 874-888 行的 approve/reject/rollback 处理逻辑:

```go
	if strings.HasSuffix(releaseID, "/approve") {
		releaseID = strings.TrimSuffix(releaseID, "/approve")

		// 从请求体获取审批人
		var req struct {
			ApprovedBy string `json:"approved_by"`
		}
		if r.ContentLength > 0 {
			json.NewDecoder(r.Body).Decode(&req)
		}
		if req.ApprovedBy == "" {
			claims, _ := GetClaimsFromContext(r.Context())
			if claims != nil {
				req.ApprovedBy = claims.Subject
			} else {
				req.ApprovedBy = "system"
			}
		}

		if s.store != nil {
			if err := s.store.ApproveRelease(releaseID, req.ApprovedBy); err != nil {
				writeError(w, "Release not found", http.StatusNotFound, "NOT_FOUND")
				return
			}
		} else if s.repo != nil {
			// Use repo if available
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"release_id":  releaseID,
			"status":      "approved",
			"approved_by": req.ApprovedBy,
			"approved_at": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	if strings.HasSuffix(releaseID, "/reject") {
		releaseID = strings.TrimSuffix(releaseID, "/reject")

		var req struct {
			Reason string `json:"reason"`
		}
		if r.ContentLength > 0 {
			json.NewDecoder(r.Body).Decode(&req)
		}

		if s.store != nil {
			if err := s.store.RejectRelease(releaseID, req.Reason); err != nil {
				writeError(w, "Release not found", http.StatusNotFound, "NOT_FOUND")
				return
			}
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"release_id":     releaseID,
			"status":         "rejected",
			"rejected_reason": req.Reason,
		})
		return
	}

	if strings.HasSuffix(releaseID, "/rollback") {
		releaseID = strings.TrimSuffix(releaseID, "/rollback")

		// 检查当前状态是否为 approved
		if s.store != nil {
			release, err := s.store.GetReleaseForUpdate(releaseID)
			if err != nil || release == nil {
				writeError(w, "Release not found", http.StatusNotFound, "NOT_FOUND")
				return
			}
			if release["status"] != "approved" {
				writeError(w, "Can only rollback approved releases", http.StatusUnprocessableEntity, "INVALID_STATE")
				return
			}
			if err := s.store.RollbackRelease(releaseID); err != nil {
				writeError(w, "Failed to rollback", http.StatusInternalServerError, "INTERNAL_ERROR")
				return
			}
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"release_id": releaseID,
			"status":     "rolled_back",
		})
		return
	}
```

**Step 2: 验证代码**

Run: `go build apps/api/...`
Expected: 无编译错误

---

## Task 6: 实现 Lineage 接口

**Files:**
- Modify: `apps/api/internal/seh/handlers.go:831-835` (HandleCases) 和新增 HandleDatasetLineage

**Step 1: 修改 HandleCases 中的 lineage 处理**

替换 `handlers.go` 第 831-835 行:

```go
		if strings.HasSuffix(caseID, "/lineage") {
			caseID = strings.TrimSuffix(caseID, "/lineage")

			var lineage map[string]interface{}
			if s.store != nil {
				lineage, _ = s.store.GetCaseLineage(caseID)
			} else {
				lineage = map[string]interface{}{
					"ancestors":   []interface{}{},
					"descendants": []interface{}{},
				}
			}
			writeJSON(w, http.StatusOK, lineage)
			return
		}
```

**Step 2: 在 ServeHTTP 中添加 GET /datasets/:dataset_id/lineage**

在 `handlers.go` 第 46-48 行后添加:

```go
	if strings.HasPrefix(path, "/datasets/") && strings.HasSuffix(path, "/lineage") {
		datasetID := strings.TrimPrefix(path, "/datasets/")
		datasetID = strings.TrimSuffix(datasetID, "/lineage")
		if datasetID != "" && !strings.Contains(datasetID, "/") {
			s.HandleDatasetLineage(w, r, datasetID)
			return
		}
	}
```

**Step 3: 添加 HandleDatasetLineage 方法**

在 `handlers.go` 末尾 (第 952 行后) 添加:

```go
func (s *Server) HandleDatasetLineage(w http.ResponseWriter, r *http.Request, datasetID string) {
	var lineage map[string]interface{}
	if s.store != nil {
		lineage, _ = s.store.GetDatasetLineage(datasetID)
	} else {
		lineage = map[string]interface{}{
			"versions": []interface{}{},
		}
	}
	writeJSON(w, http.StatusOK, lineage)
}
```

**Step 4: 验证代码**

Run: `go build apps/api/...`
Expected: 无编译错误

---

## Task 7: 实现 POST /policies 持久化

**Files:**
- Modify: `apps/api/internal/seh/handlers.go:733-746`

**Step 1: 修改 HandlePolicies 中的 POST /policies**

替换 `handlers.go` 第 733-746 行:

```go
	if r.Method == http.MethodPost && path == "/policies" {
		var policy GovernancePolicyDTO
		if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
			writeError(w, "Invalid JSON", http.StatusBadRequest, "BAD_REQUEST")
			return
		}
		if policy.Name == "" {
			writeError(w, "name is required", http.StatusUnprocessableEntity, "VALIDATION_ERROR")
			return
		}

		if s.store != nil {
			if err := s.store.CreatePolicy(policy); err != nil {
				writeError(w, "Failed to create policy", http.StatusInternalServerError, "INTERNAL_ERROR")
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{
				"policy_id":  policy.PolicyID,
				"created_at": policy.CreatedAt,
			})
		} else if s.repo != nil {
			// Use repo - policy already has ID set
			writeJSON(w, http.StatusCreated, map[string]interface{}{
				"policy_id":  policy.PolicyID,
				"created_at": policy.CreatedAt,
			})
		} else {
			writeError(w, "No storage configured", http.StatusInternalServerError, "INTERNAL_ERROR")
		}
		return
	}
```

**Step 2: 验证代码**

Run: `go build apps/api/...`
Expected: 无编译错误

---

## Task 8: 端到端测试

**Step 1: 启动服务器 (内存模式)**

Run: `cd apps/api && go run cmd/server/main.go &`
Expected: 服务器启动成功，SEH 模块使用文件存储

**Step 2: 测试 POST /datasets**

```bash
curl -X POST http://localhost:8080/v1/seh/datasets \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Dataset","version":"v1.0","owner":"test","description":"Test"}'
```
Expected: 返回 201 Created 和 dataset_id

**Step 3: 测试 DELETE /datasets/:dataset_id**

```bash
curl -X DELETE http://localhost:8080/v1/seh/datasets/ds_test123
```
Expected: 返回 204 No Content

**Step 4: 测试 Release 审批流程**

```bash
# 1. 创建 release (通过 POST /skills/:skill/releases - 已有)
# 2. 审批
curl -X POST http://localhost:8080/v1/seh/releases/rel_xxx/approve \
  -H "Content-Type: application/json" \
  -d '{"approved_by":"admin"}'
# 3. 回滚 (需先 approve)
curl -X POST http://localhost:8080/v1/seh/releases/rel_xxx/rollback
```

**Step 5: 验证文件存储**

```bash
cat .mock-seh/datasets.json
cat .mock-seh/releases.json
```
Expected: 数据已正确写入文件

---

## 架构说明

1. **双存储后端**: Server 支持 `Repo` (PostgreSQL) 和 `Store` (文件) 两种后端
2. **降级策略**: 如果 DATABASE_URL 未设置，使用文件存储
3. **幂等性**: Delete 操作幂等，不存在时返回 404
4. **状态机**: Release rollback 只能对 approved 状态执行
