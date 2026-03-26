package approval

import (
	"fmt"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/engine/tool"
)

// Status represents the status of an approval request.
type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

// Request represents a pending approval request.
type Request struct {
	ID           string
	ToolName     string
	ToolDef      *tool.ToolDefinition
	Args         map[string]any
	RiskLevel    tool.RiskLevel
	Status       Status
	RequestedAt  time.Time
	ReviewedAt   *time.Time
	Reviewer     string
	Result       any
	RejectReason string
}

// ApprovalService manages tool execution approvals.
type ApprovalService interface {
	CreateRequest(toolDef *tool.ToolDefinition, args map[string]any) (*Request, error)
	GetRequest(id string) (*Request, bool)
	Approve(id string, reviewer string, result any) error
	Reject(id string, reviewer string, reason string) error
	ListPending() []*Request
}

type approvalService struct {
	requests map[string]*Request
}

// NewApprovalService creates a new in-memory approval service.
func NewApprovalService() ApprovalService {
	return &approvalService{
		requests: make(map[string]*Request),
	}
}

func (s *approvalService) CreateRequest(toolDef *tool.ToolDefinition, args map[string]any) (*Request, error) {
	id := fmt.Sprintf("apr_%d", time.Now().UnixNano())
	req := &Request{
		ID:          id,
		ToolName:    toolDef.Name,
		ToolDef:     toolDef,
		Args:        args,
		RiskLevel:   toolDef.RiskLevel,
		Status:      StatusPending,
		RequestedAt: time.Now(),
	}
	s.requests[id] = req
	return req, nil
}

func (s *approvalService) GetRequest(id string) (*Request, bool) {
	req, ok := s.requests[id]
	return req, ok
}

func (s *approvalService) Approve(id string, reviewer string, result any) error {
	req, ok := s.requests[id]
	if !ok {
		return fmt.Errorf("approval request not found: %s", id)
	}
	if req.Status != StatusPending {
		return fmt.Errorf("approval request already processed: %s", req.Status)
	}
	now := time.Now()
	req.Status = StatusApproved
	req.ReviewedAt = &now
	req.Reviewer = reviewer
	req.Result = result
	return nil
}

func (s *approvalService) Reject(id string, reviewer string, reason string) error {
	req, ok := s.requests[id]
	if !ok {
		return fmt.Errorf("approval request not found: %s", id)
	}
	if req.Status != StatusPending {
		return fmt.Errorf("approval request already processed: %s", req.Status)
	}
	now := time.Now()
	req.Status = StatusRejected
	req.ReviewedAt = &now
	req.Reviewer = reviewer
	req.RejectReason = reason
	return nil
}

func (s *approvalService) ListPending() []*Request {
	var pending []*Request
	for _, req := range s.requests {
		if req.Status == StatusPending {
			pending = append(pending, req)
		}
	}
	return pending
}
