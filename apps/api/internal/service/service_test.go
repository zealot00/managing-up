package service

import (
	"context"
	"errors"
	"testing"
)

// Mock implementations

type mockSkillRepo struct {
	skills []Skill
	err    error
}

func (m *mockSkillRepo) ListSkills(status string) []Skill { return m.skills }
func (m *mockSkillRepo) GetSkill(id string) (Skill, bool) {
	if len(m.skills) > 0 {
		return m.skills[0], true
	}
	return Skill{}, false
}
func (m *mockSkillRepo) CreateSkill(req CreateSkillRequest) Skill {
	return Skill{ID: "new_skill", Name: req.Name}
}
func (m *mockSkillRepo) ListSkillsByCategory(ctx context.Context, category, search string) ([]Skill, error) {
	return m.skills, m.err
}
func (m *mockSkillRepo) ListDependencies(ctx context.Context, skillID string) ([]SkillDependency, error) {
	return nil, nil
}
func (m *mockSkillRepo) ResolveDepTree(ctx context.Context, skillID string) ([]DependencyNode, error) {
	return nil, nil
}
func (m *mockSkillRepo) UpsertRating(ctx context.Context, skillID, userID string, rating int, comment string) error {
	return nil
}
func (m *mockSkillRepo) GetRatingStats(ctx context.Context, skillID string) (float64, int, error) {
	return 0, 0, nil
}
func (m *mockSkillRepo) GetInstallCount(ctx context.Context, skillID string) (int, error) {
	return 0, nil
}

type mockExecutionRepo struct {
	execution Execution
	skill     Skill
}

func (m *mockExecutionRepo) GetSkill(id string) (Skill, bool) {
	if m.skill.ID != "" {
		return m.skill, true
	}
	return Skill{}, false
}
func (m *mockExecutionRepo) CreateExecution(req CreateExecutionRequest) (Execution, bool) {
	return m.execution, true
}
func (m *mockExecutionRepo) ApproveExecution(id string, req ApproveExecutionRequest) (Approval, bool) {
	return Approval{ID: "approval_001", Status: req.Decision}, true
}

// SkillService Tests

func TestCreateSkill_Success(t *testing.T) {
	repo := &mockSkillRepo{skills: []Skill{}}
	svc := NewSkillService(repo)

	req := CreateSkillRequest{
		Name:      "TestSkill",
		OwnerTeam: "team-a",
		RiskLevel: "low",
	}

	skill, err := svc.CreateSkill(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if skill.Name != req.Name {
		t.Errorf("expected skill name %q, got %q", req.Name, skill.Name)
	}
}

func TestCreateSkill_EmptyName(t *testing.T) {
	repo := &mockSkillRepo{skills: []Skill{}}
	svc := NewSkillService(repo)

	req := CreateSkillRequest{
		Name:      "",
		OwnerTeam: "team-a",
		RiskLevel: "low",
	}

	_, err := svc.CreateSkill(req)

	if !errors.Is(err, ErrSkillNameRequired) {
		t.Errorf("expected ErrSkillNameRequired, got %v", err)
	}
}

func TestCreateSkill_EmptyOwnerTeam(t *testing.T) {
	repo := &mockSkillRepo{skills: []Skill{}}
	svc := NewSkillService(repo)

	req := CreateSkillRequest{
		Name:      "TestSkill",
		OwnerTeam: "",
		RiskLevel: "low",
	}

	_, err := svc.CreateSkill(req)

	if !errors.Is(err, ErrOwnerTeamRequired) {
		t.Errorf("expected ErrOwnerTeamRequired, got %v", err)
	}
}

func TestCreateSkill_InvalidRiskLevel(t *testing.T) {
	repo := &mockSkillRepo{skills: []Skill{}}
	svc := NewSkillService(repo)

	req := CreateSkillRequest{
		Name:      "TestSkill",
		OwnerTeam: "team-a",
		RiskLevel: "invalid",
	}

	_, err := svc.CreateSkill(req)

	if !errors.Is(err, ErrInvalidRiskLevel) {
		t.Errorf("expected ErrInvalidRiskLevel, got %v", err)
	}
}

func TestCreateSkill_DuplicateName(t *testing.T) {
	repo := &mockSkillRepo{
		skills: []Skill{{Name: "ExistingSkill"}},
	}
	svc := NewSkillService(repo)

	req := CreateSkillRequest{
		Name:      "ExistingSkill",
		OwnerTeam: "team-a",
		RiskLevel: "low",
	}

	_, err := svc.CreateSkill(req)

	if !errors.Is(err, ErrDuplicateSkillName) {
		t.Errorf("expected ErrDuplicateSkillName, got %v", err)
	}
}

// ExecutionService Tests

func TestCreateExecution_Success(t *testing.T) {
	repo := &mockExecutionRepo{
		execution: Execution{ID: "exec-001"},
		skill:     Skill{ID: "skill-001", Name: "TestSkill"},
	}
	svc := NewExecutionService(repo)

	req := CreateExecutionRequest{
		SkillID:     "skill-001",
		TriggeredBy: "user-001",
		Input:       map[string]any{"key": "value"},
	}

	exec, err := svc.CreateExecution(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if exec.ID != "exec-001" {
		t.Errorf("expected execution ID %q, got %q", "exec-001", exec.ID)
	}
}

func TestCreateExecution_SkillNotFound(t *testing.T) {
	repo := &mockExecutionRepo{
		execution: Execution{},
		skill:     Skill{},
	}
	svc := NewExecutionService(repo)

	req := CreateExecutionRequest{
		SkillID:     "nonexistent",
		TriggeredBy: "user-001",
	}

	_, err := svc.CreateExecution(req)

	if !errors.Is(err, ErrSkillNotFound) {
		t.Errorf("expected ErrSkillNotFound, got %v", err)
	}
}

func TestCreateExecution_EmptySkillID(t *testing.T) {
	repo := &mockExecutionRepo{}
	svc := NewExecutionService(repo)

	req := CreateExecutionRequest{
		SkillID:     "",
		TriggeredBy: "user-001",
	}

	_, err := svc.CreateExecution(req)

	if !errors.Is(err, ErrSkillIDRequired) {
		t.Errorf("expected ErrSkillIDRequired, got %v", err)
	}
}

func TestCreateExecution_EmptyTriggeredBy(t *testing.T) {
	repo := &mockExecutionRepo{
		skill: Skill{ID: "skill-001"},
	}
	svc := NewExecutionService(repo)

	req := CreateExecutionRequest{
		SkillID:     "skill-001",
		TriggeredBy: "",
	}

	_, err := svc.CreateExecution(req)

	if !errors.Is(err, ErrTriggeredByRequired) {
		t.Errorf("expected ErrTriggeredByRequired, got %v", err)
	}
}

func TestApproveExecution_Success(t *testing.T) {
	repo := &mockExecutionRepo{}
	svc := NewExecutionService(repo)

	req := ApproveExecutionRequest{
		Approver: "admin-001",
		Decision: "approved",
		Note:     "looks good",
	}

	approval, err := svc.ApproveExecution("exec-001", req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if approval.ID != "approval_001" {
		t.Errorf("expected approval ID %q, got %q", "approval_001", approval.ID)
	}
	if approval.Status != "approved" {
		t.Errorf("expected status %q, got %q", "approved", approval.Status)
	}
}

func TestApproveExecution_EmptyApprover(t *testing.T) {
	repo := &mockExecutionRepo{}
	svc := NewExecutionService(repo)

	req := ApproveExecutionRequest{
		Approver: "",
		Decision: "approved",
	}

	_, err := svc.ApproveExecution("exec-001", req)

	if !errors.Is(err, ErrApproverRequired) {
		t.Errorf("expected ErrApproverRequired, got %v", err)
	}
}

func TestApproveExecution_InvalidDecision(t *testing.T) {
	repo := &mockExecutionRepo{}
	svc := NewExecutionService(repo)

	req := ApproveExecutionRequest{
		Approver: "admin-001",
		Decision: "maybe",
	}

	_, err := svc.ApproveExecution("exec-001", req)

	if !errors.Is(err, ErrInvalidDecision) {
		t.Errorf("expected ErrInvalidDecision, got %v", err)
	}
}

// Error Types Tests

func TestErrors_AreDistinct(t *testing.T) {
	errs := []error{
		ErrSkillNotFound,
		ErrExecutionNotFound,
		ErrInvalidRiskLevel,
		ErrInvalidDecision,
		ErrSkillNameRequired,
		ErrOwnerTeamRequired,
		ErrSkillIDRequired,
		ErrTriggeredByRequired,
		ErrApproverRequired,
		ErrDuplicateSkillName,
	}

	// Check all errors are distinct
	for i, err1 := range errs {
		for j, err2 := range errs {
			if i != j && errors.Is(err1, err2) {
				t.Errorf("errors %v and %v should be distinct", err1, err2)
			}
		}
	}

	// Check errors.Is works correctly
	if !errors.Is(ErrSkillNotFound, ErrSkillNotFound) {
		t.Errorf("errors.Is(ErrSkillNotFound, ErrSkillNotFound) should be true")
	}
	if !errors.Is(ErrInvalidRiskLevel, ErrInvalidRiskLevel) {
		t.Errorf("errors.Is(ErrInvalidRiskLevel, ErrInvalidRiskLevel) should be true")
	}
}
