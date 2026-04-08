package service

import (
	"context"
	"fmt"
	"time"
)

type MCPServer struct {
	ID            string
	ServerID      string
	Name          string
	TrustScore    float64
	TransportType string
	URL           string
}

type RouterCatalogEntry struct {
	ID            string
	ServerID      string
	Name          string
	TrustScore    float64
	TransportType string
	URL           string
}

type MCPRouterRepository interface {
	FindMatchingServers(ctx context.Context, taskTypes []string, tags []string) ([]MCPServer, error)
	IncrementUseCount(ctx context.Context, id string)
	SyncServer(ctx context.Context, serverID string, approvedBy string) error
	ListCatalog(ctx context.Context) ([]RouterCatalogEntry, error)
}

type MCPRouterService struct {
	repo             MCPRouterRepository
	metricsCollector *MetricsCollector
}

func NewMCPRouterService(repo MCPRouterRepository, mc *MetricsCollector) *MCPRouterService {
	return &MCPRouterService{repo: repo, metricsCollector: mc}
}

func (s *MCPRouterService) MatchTask(ctx context.Context, taskTypes []string, tags []string) (*MatchResult, error) {
	start := time.Now()

	servers, err := s.repo.FindMatchingServers(ctx, taskTypes, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching servers: %w", err)
	}

	if len(servers) == 0 {
		return &MatchResult{Matched: false}, nil
	}

	best := servers[0]

	s.repo.IncrementUseCount(ctx, best.ID)

	_ = time.Since(start)

	return &MatchResult{
		Matched:    true,
		ServerID:   best.ServerID,
		ServerName: best.Name,
		Transport:  best.TransportType,
		Endpoint:   best.URL,
		MatchScore: best.TrustScore,
	}, nil
}

type MatchResult struct {
	Matched    bool
	ServerID   string
	ServerName string
	Transport  string
	Endpoint   string
	MatchScore float64
}

func (s *MCPRouterService) SyncFromMCPServer(ctx context.Context, serverID string, approvedBy string) error {
	return s.repo.SyncServer(ctx, serverID, approvedBy)
}

func (s *MCPRouterService) GetCatalog(ctx context.Context) ([]RouterCatalogEntry, error) {
	return s.repo.ListCatalog(ctx)
}
