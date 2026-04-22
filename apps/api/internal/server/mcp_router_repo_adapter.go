package server

import (
	"context"
	"sync"

	"github.com/zealot/managing-up/apps/api/internal/models"
	"github.com/zealot/managing-up/apps/api/internal/service"
)

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
