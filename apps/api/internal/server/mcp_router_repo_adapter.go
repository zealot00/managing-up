package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/zealot/managing-up/apps/api/internal/models"
	"github.com/zealot/managing-up/apps/api/internal/repository"
	"github.com/zealot/managing-up/apps/api/internal/server/handlers"
	"github.com/zealot/managing-up/apps/api/internal/service"
)

// postgresMCPRouterRepoAdapter adapts repository.MCPRouterRepository
// (which uses repository-level types) to service.MCPRouterRepository
// (which uses service-level types).
type postgresMCPRouterRepoAdapter struct {
	inner *repository.PostgresMCPRouterRepository
}

var _ service.MCPRouterRepository = (*postgresMCPRouterRepoAdapter)(nil)

func newPostgresMCPRouterRepoAdapter(inner *repository.PostgresMCPRouterRepository) *postgresMCPRouterRepoAdapter {
	return &postgresMCPRouterRepoAdapter{inner: inner}
}

func (a *postgresMCPRouterRepoAdapter) FindMatchingServers(ctx context.Context, taskTypes []string, tags []string) ([]service.MCPServer, error) {
	entries, err := a.inner.FindMatchingServers(ctx, taskTypes, tags)
	if err != nil {
		return nil, err
	}
	servers := make([]service.MCPServer, len(entries))
	for i, e := range entries {
		servers[i] = service.MCPServer{
			ID:            e.ID,
			ServerID:      e.ServerID,
			Name:          e.Name,
			TrustScore:    e.TrustScore,
			TransportType: e.TransportType,
			URL:           e.URL,
		}
	}
	return servers, nil
}

func (a *postgresMCPRouterRepoAdapter) IncrementUseCount(ctx context.Context, id string) {
	_ = a.inner.IncrementUseCount(ctx, id)
}

func (a *postgresMCPRouterRepoAdapter) SyncServer(ctx context.Context, server service.MCPServer, approvedBy string) error {
	return a.inner.SyncServer(ctx, server.ServerID, approvedBy)
}

func (a *postgresMCPRouterRepoAdapter) ListCatalog(ctx context.Context) ([]service.RouterCatalogEntry, error) {
	entries, err := a.inner.ListCatalog(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]service.RouterCatalogEntry, len(entries))
	for i, e := range entries {
		result[i] = service.RouterCatalogEntry{
			ID:            e.ID,
			ServerID:      e.ServerID,
			Name:          e.Name,
			TrustScore:    e.TrustScore,
			TransportType: e.TransportType,
			URL:           e.URL,
		}
	}
	return result, nil
}

// postgresRouteLoggerAdapter adapts repository.PostgresMCPRouterRepository
// to the handlers.RouteLogger interface.
type postgresRouteLoggerAdapter struct {
	inner *repository.PostgresMCPRouterRepository
}

func newPostgresRouteLoggerAdapter(inner *repository.PostgresMCPRouterRepository) *postgresRouteLoggerAdapter {
	return &postgresRouteLoggerAdapter{inner: inner}
}

func (a *postgresRouteLoggerAdapter) LogRoute(ctx context.Context, log *handlers.RouteLogEntry) error {
	repoLog := &repository.RouteLogEntry{
		SessionID:       log.SessionID,
		CorrelationID:   log.CorrelationID,
		AgentID:         log.AgentID,
		TaskType:        log.TaskType,
		TaskTags:        log.TaskTags,
		Matched:         log.Matched,
		MatchedServerID: log.MatchedServerID,
		MatchScore:      log.MatchScore,
		MatchLatencyMS:  log.MatchLatencyMS,
		Status:          log.Status,
	}
	return a.inner.LogRouteWithSession(ctx, repoLog)
}

type inMemoryMCPRouterRepo struct {
	mu       sync.RWMutex
	servers  map[string]service.MCPServer
	catalog  []service.RouterCatalogEntry
	useCount map[string]int64
}

var _ service.MCPRouterRepository = (*inMemoryMCPRouterRepo)(nil)

func newInMemoryMCPRouterRepo() *inMemoryMCPRouterRepo {
	return &inMemoryMCPRouterRepo{
		servers:  make(map[string]service.MCPServer),
		catalog:  make([]service.RouterCatalogEntry, 0),
		useCount: make(map[string]int64),
	}
}

func (r *inMemoryMCPRouterRepo) FindMatchingServers(ctx context.Context, taskTypes []string, tags []string) ([]service.MCPServer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(taskTypes) == 0 && len(tags) == 0 {
		return nil, nil
	}

	var best *service.MCPServer
	var bestScore float64 = -1

	for _, server := range r.servers {
		score := server.TrustScore
		if score > bestScore {
			bestScore = score
			best = &server
		}
	}

	if best == nil {
		return nil, nil
	}

	return []service.MCPServer{*best}, nil
}

func (r *inMemoryMCPRouterRepo) IncrementUseCount(ctx context.Context, id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.useCount[id]++
}

func (r *inMemoryMCPRouterRepo) SyncServer(ctx context.Context, server service.MCPServer, approvedBy string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.servers[server.ID] = server
	r.catalog = append(r.catalog, service.RouterCatalogEntry{
		ID:            server.ID,
		ServerID:      server.ID,
		Name:          server.Name,
		TrustScore:    server.TrustScore,
		TransportType: server.TransportType,
		URL:           server.URL,
	})
	return nil
}

func (r *inMemoryMCPRouterRepo) ListCatalog(ctx context.Context) ([]service.RouterCatalogEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.catalog, nil
}

func (r *inMemoryMCPRouterRepo) AddServer(server service.MCPServer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.servers[server.ID] = server
	r.catalog = append(r.catalog, service.RouterCatalogEntry{
		ID:            server.ID,
		ServerID:      server.ServerID,
		Name:          server.Name,
		TrustScore:    server.TrustScore,
		TransportType: server.TransportType,
		URL:           server.URL,
	})
}

type inMemoryGatewaySessionRepo struct {
	mu       sync.RWMutex
	sessions map[string]*service.GatewaySession
}

var _ service.GatewaySessionRepository = (*inMemoryGatewaySessionRepo)(nil)

func newInMemoryGatewaySessionRepo() *inMemoryGatewaySessionRepo {
	return &inMemoryGatewaySessionRepo{
		sessions: make(map[string]*service.GatewaySession),
	}
}

func (r *inMemoryGatewaySessionRepo) CreateGatewaySession(ctx context.Context, session *service.GatewaySession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.ID] = session
	return nil
}

func (r *inMemoryGatewaySessionRepo) UpdatePolicyDecision(ctx context.Context, sessionID string, decision *models.PolicyDecision) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if session, ok := r.sessions[sessionID]; ok {
		session.PolicyDecision = map[string]interface{}{
			"allowed":            decision.Allowed,
			"required_approvals": decision.RequiredApprovals,
			"policy_id":          decision.PolicyID,
			"policy_version":     decision.PolicyVersion,
			"reasons":            decision.Reasons,
			"determined_at":      decision.DeterminedAt,
		}
	}
	return nil
}

func (r *inMemoryGatewaySessionRepo) ListGatewaySessions(ctx context.Context, agentID string, limit int) ([]*service.GatewaySession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*service.GatewaySession
	for _, session := range r.sessions {
		if agentID == "" || session.AgentID == agentID {
			result = append(result, session)
		}
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result, nil
}

type inMemoryMemoryRepo struct {
	mu    sync.RWMutex
	cells map[string]*models.MemoryCell
}

var _ service.MemoryRepository = (*inMemoryMemoryRepo)(nil)

func newInMemoryMemoryRepo() *inMemoryMemoryRepo {
	return &inMemoryMemoryRepo{
		cells: make(map[string]*models.MemoryCell),
	}
}

func (r *inMemoryMemoryRepo) CreateMemoryCell(ctx context.Context, cell *models.MemoryCell) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cells[cell.ID] = cell
	return nil
}

func (r *inMemoryMemoryRepo) GetMemoryCell(ctx context.Context, id string) (*models.MemoryCell, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if cell, ok := r.cells[id]; ok {
		return cell, nil
	}
	return nil, fmt.Errorf("memory cell not found: %s", id)
}

func (r *inMemoryMemoryRepo) GetMemoryCellsBySession(ctx context.Context, sessionID string, limit int) ([]models.MemoryCell, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []models.MemoryCell
	for _, cell := range r.cells {
		if cell.SessionID == sessionID {
			result = append(result, *cell)
		}
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (r *inMemoryMemoryRepo) GetMemoryCellsByAgent(ctx context.Context, agentID string, limit int) ([]models.MemoryCell, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []models.MemoryCell
	for _, cell := range r.cells {
		if cell.AgentID == agentID {
			result = append(result, *cell)
		}
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (r *inMemoryMemoryRepo) UpdateMemoryCell(ctx context.Context, cell *models.MemoryCell) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cells[cell.ID] = cell
	return nil
}

func (r *inMemoryMemoryRepo) DeleteMemoryCell(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.cells, id)
	return nil
}
