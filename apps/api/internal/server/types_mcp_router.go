package server

type JSONMap map[string]any

type MCPRouterCatalog struct {
	ID            string   `json:"id"`
	ServerID      string   `json:"server_id"`
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	TransportType string   `json:"transport_type"`
	Command       string   `json:"command,omitempty"`
	Args          []string `json:"args,omitempty"`
	URL           string   `json:"url,omitempty"`
	TaskTypes     []string `json:"task_types"`
	Tags          []string `json:"tags,omitempty"`
	Capabilities  JSONMap  `json:"capabilities"`
	RoutingConfig JSONMap  `json:"routing_config,omitempty"`
	Status        string   `json:"status"`
	TrustScore    float64  `json:"trust_score"`
	UseCount      int64    `json:"use_count"`
	ErrorCount    int64    `json:"error_count"`
	LastUsedAt    *string  `json:"last_used_at,omitempty"`
	SyncedAt      string   `json:"synced_at"`
	CreatedAt     string   `json:"created_at"`
}

type MCPRouterLog struct {
	ID              string   `json:"id"`
	CorrelationID   string   `json:"correlation_id"`
	AgentID         string   `json:"agent_id,omitempty"`
	TaskType        string   `json:"task_type,omitempty"`
	TaskTags        []string `json:"task_tags,omitempty"`
	TaskComplexity  string   `json:"task_complexity,omitempty"`
	RawDescription  string   `json:"raw_description,omitempty"`
	Matched         bool     `json:"matched"`
	MatchedServerID string   `json:"matched_server_id,omitempty"`
	MatchScore      float64  `json:"match_score,omitempty"`
	MatchLatencyMS  int      `json:"match_latency_ms,omitempty"`
	Status          string   `json:"status,omitempty"`
	ErrorCode       string   `json:"error_code,omitempty"`
	ErrorMessage    string   `json:"error_message,omitempty"`
	DurationMS      int      `json:"duration_ms,omitempty"`
	CreatedAt       string   `json:"created_at"`
}

type RouteRequest struct {
	Task          RouteTask `json:"task"`
	AgentID       string    `json:"agent_id,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
}

type RouteTask struct {
	Description string         `json:"description,omitempty"`
	Structured  TaskStructured `json:"structured,omitempty"`
}

type TaskStructured struct {
	TaskType   string   `json:"task_type,omitempty"`
	Language   string   `json:"language,omitempty"`
	Complexity string   `json:"complexity,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

type RouteResponse struct {
	Matched       bool         `json:"matched"`
	Target        *RouteTarget `json:"target,omitempty"`
	MatchScore    float64      `json:"match_score,omitempty"`
	RoutingTimeMS int          `json:"routing_time_ms,omitempty"`
}

type RouteTarget struct {
	ServerID   string `json:"server_id"`
	ServerName string `json:"server_name"`
	Transport  string `json:"transport"`
	Endpoint   string `json:"endpoint,omitempty"`
}
