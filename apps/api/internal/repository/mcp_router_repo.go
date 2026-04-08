package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type MCPServer struct {
	ID              string
	Name            string
	Description     string
	TransportType   string
	Command         string
	Args            []string
	Env             []string
	URL             string
	Headers         []string
	Tags            []string
	Status          string
	RejectionReason string
	ApprovedBy      string
	ApprovedAt      *time.Time
	IsEnabled       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type MCPRouterRepository interface {
	FindMatchingServers(ctx context.Context, taskTypes []string, tags []string) ([]RouterCatalogEntry, error)
	ListCatalog(ctx context.Context) ([]RouterCatalogEntry, error)
	IncrementUseCount(ctx context.Context, id string) error
	SyncServer(ctx context.Context, serverID string, approvedBy string) error
	GetServerByID(ctx context.Context, serverID string) (*MCPServer, error)
	LogRoute(ctx context.Context, log *RouteLogEntry) error
}

type RouterCatalogEntry struct {
	ID            string
	ServerID      string
	Name          string
	TransportType string
	Command       string
	Args          []string
	URL           string
	TaskTypes     []string
	Tags          []string
	TrustScore    float64
	UseCount      int64
}

type RouteLogEntry struct {
	CorrelationID   string
	AgentID         string
	TaskType        string
	TaskTags        []string
	RawDescription  string
	Matched         bool
	MatchedServerID string
	MatchScore      float64
	MatchLatencyMS  int
	Status          string
	ErrorCode       string
	ErrorMessage    string
	DurationMS      int
}

type PostgresMCPRouterRepository struct {
	db *sql.DB
}

func NewPostgresMCPRouterRepository(db *sql.DB) *PostgresMCPRouterRepository {
	return &PostgresMCPRouterRepository{db: db}
}

func (r *PostgresMCPRouterRepository) FindMatchingServers(ctx context.Context, taskTypes []string, tags []string) ([]RouterCatalogEntry, error) {
	query := `
		SELECT id, server_id, name, transport_type, command, args, url, task_types, tags, trust_score, use_count
		FROM mcp_router_catalog
		WHERE status = 'active'
		AND task_types @> $1
		ORDER BY trust_score DESC, use_count DESC
		LIMIT 1
	`

	var entry RouterCatalogEntry
	err := r.db.QueryRowContext(ctx, query, taskTypes).Scan(
		&entry.ID, &entry.ServerID, &entry.Name, &entry.TransportType,
		&entry.Command, &entry.Args, &entry.URL, &entry.TaskTypes, &entry.Tags,
		&entry.TrustScore, &entry.UseCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return []RouterCatalogEntry{entry}, nil
}

func (r *PostgresMCPRouterRepository) ListCatalog(ctx context.Context) ([]RouterCatalogEntry, error) {
	query := `
		SELECT id, server_id, name, transport_type, command, args, url, task_types, tags, trust_score, use_count
		FROM mcp_router_catalog
		WHERE status = 'active'
		ORDER BY trust_score DESC, use_count DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var entries []RouterCatalogEntry
	for rows.Next() {
		var entry RouterCatalogEntry
		err := rows.Scan(
			&entry.ID, &entry.ServerID, &entry.Name, &entry.TransportType,
			&entry.Command, &entry.Args, &entry.URL, &entry.TaskTypes, &entry.Tags,
			&entry.TrustScore, &entry.UseCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *PostgresMCPRouterRepository) IncrementUseCount(ctx context.Context, id string) error {
	query := `UPDATE mcp_router_catalog SET use_count = use_count + 1, last_used_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *PostgresMCPRouterRepository) SyncServer(ctx context.Context, serverID string, approvedBy string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var server MCPServer
	err = tx.QueryRowContext(ctx, `
		SELECT id, name, description, transport_type, command, args, env, url, headers
		FROM mcp_servers WHERE id = $1
	`, serverID).Scan(&server.ID, &server.Name, &server.Description, &server.TransportType,
		&server.Command, &server.Args, &server.Env, &server.URL, &server.Headers)
	if err != nil {
		return fmt.Errorf("failed to get server: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO mcp_router_catalog (server_id, name, description, transport_type, command, args, url, task_types, tags, capabilities, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, '{}', 'active')
		ON CONFLICT (server_id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			transport_type = EXCLUDED.transport_type,
			command = EXCLUDED.command,
			args = EXCLUDED.args,
			url = EXCLUDED.url,
			synced_at = NOW()
	`, server.ID, server.Name, server.Description, server.TransportType, server.Command, server.Args, server.URL, server.Tags)
	if err != nil {
		return fmt.Errorf("failed to sync catalog: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO mcp_router_sync_log (server_id, sync_type, approved_by, approved_at)
		VALUES ($1, 'approved_sync', $2, NOW())
	`, serverID, approvedBy)
	if err != nil {
		return fmt.Errorf("failed to log sync: %w", err)
	}

	return tx.Commit()
}

func (r *PostgresMCPRouterRepository) GetServerByID(ctx context.Context, serverID string) (*MCPServer, error) {
	var server MCPServer
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, transport_type, command, args, env, url, headers
		FROM mcp_servers WHERE id = $1
	`, serverID).Scan(&server.ID, &server.Name, &server.Description, &server.TransportType,
		&server.Command, &server.Args, &server.Env, &server.URL, &server.Headers)
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (r *PostgresMCPRouterRepository) LogRoute(ctx context.Context, log *RouteLogEntry) error {
	query := `
		INSERT INTO mcp_router_logs (
			correlation_id, agent_id, task_type, task_tags, raw_description,
			matched, matched_server_id, match_score, match_latency_ms,
			status, error_code, error_message, duration_ms
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.CorrelationID, log.AgentID, log.TaskType, log.TaskTags, log.RawDescription,
		log.Matched, log.MatchedServerID, log.MatchScore, log.MatchLatencyMS,
		log.Status, log.ErrorCode, log.ErrorMessage, log.DurationMS,
	)
	return err
}
