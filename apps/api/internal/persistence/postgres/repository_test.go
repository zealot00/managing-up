package postgres

import (
	"os"
	"testing"

	"github.com/zealot/managing-up/apps/api/internal/server"
)

func TestNew(t *testing.T) {
	t.Parallel()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	repo, err := New(dsn)
	if err != nil {
		t.Fatalf("expected postgres repository to connect: %v", err)
	}
	defer repo.Close()
}

// skipWithoutDB skips the test if TEST_DATABASE_URL is not set.
func skipWithoutDB(t *testing.T) {
	t.Helper()
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
}

// setupRepo creates a repository and returns it along with a cleanup function.
func setupRepo(t *testing.T) (*Repository, func()) {
	t.Helper()
	skipWithoutDB(t)

	repo, err := New(os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	return repo, func() {
		repo.Close()
	}
}

// createTestSkill creates a skill for testing and returns it.
func createTestSkill(t *testing.T, repo *Repository, name, team, riskLevel string) server.Skill {
	t.Helper()
	req := server.CreateSkillRequest{
		Name:      name,
		OwnerTeam: team,
		RiskLevel: riskLevel,
	}
	skill := repo.CreateSkill(req)
	if skill.ID == "" {
		t.Fatalf("expected non-empty skill ID after create")
	}
	return skill
}

// cleanupSkill deletes a skill by ID.
func cleanupSkill(t *testing.T, repo *Repository, id string) {
	t.Helper()
	_, err := repo.db.Exec(`DELETE FROM skills WHERE id = $1`, id)
	if err != nil {
		t.Logf("warning: failed to cleanup skill %s: %v", id, err)
	}
}

// cleanupExecution deletes an execution by ID.
func cleanupExecution(t *testing.T, repo *Repository, id string) {
	t.Helper()
	_, err := repo.db.Exec(`DELETE FROM executions WHERE id = $1`, id)
	if err != nil {
		t.Logf("warning: failed to cleanup execution %s: %v", id, err)
	}
}

// cleanupApproval deletes an approval by execution ID.
func cleanupApproval(t *testing.T, repo *Repository, executionID string) {
	t.Helper()
	_, err := repo.db.Exec(`DELETE FROM approvals WHERE execution_id = $1`, executionID)
	if err != nil {
		t.Logf("warning: failed to cleanup approval for execution %s: %v", executionID, err)
	}
}

func TestCreateSkill(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	skill := createTestSkill(t, repo, "test_skill_create", "test_team", "low")
	defer cleanupSkill(t, repo, skill.ID)

	if skill.Name != "test_skill_create" {
		t.Errorf("expected skill name 'test_skill_create', got '%s'", skill.Name)
	}
	if skill.OwnerTeam != "test_team" {
		t.Errorf("expected owner team 'test_team', got '%s'", skill.OwnerTeam)
	}
	if skill.RiskLevel != "low" {
		t.Errorf("expected risk level 'low', got '%s'", skill.RiskLevel)
	}
	if skill.Status != "draft" {
		t.Errorf("expected status 'draft', got '%s'", skill.Status)
	}

	retrieved, ok := repo.GetSkill(skill.ID)
	if !ok {
		t.Fatalf("expected to retrieve skill %s after creation", skill.ID)
	}
	if retrieved.Name != skill.Name {
		t.Errorf("expected retrieved skill name '%s', got '%s'", skill.Name, retrieved.Name)
	}
}

func TestCreateExecutionWithInvalidSkill(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	req := server.CreateExecutionRequest{
		SkillID:     "nonexistent_skill_id",
		TriggeredBy: "test_user",
		Input:       map[string]any{"key": "value"},
	}

	execution, ok := repo.CreateExecution(req)
	if ok {
		t.Errorf("expected CreateExecution with invalid skill ID to return ok=false, got ok=true with execution ID %s", execution.ID)
	}
	if execution.ID != "" {
		t.Errorf("expected empty execution ID when skill not found, got '%s'", execution.ID)
	}
}

func TestApproveExecutionUpdatesStatus(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	skill := createTestSkill(t, repo, "test_skill_approval", "test_team", "medium")
	defer cleanupSkill(t, repo, skill.ID)

	execReq := server.CreateExecutionRequest{
		SkillID:     skill.ID,
		TriggeredBy: "test_user",
		Input:       map[string]any{"server_id": "srv-test"},
	}
	execution, ok := repo.CreateExecution(execReq)
	if !ok {
		t.Fatalf("expected successful execution creation for skill %s", skill.ID)
	}
	defer cleanupExecution(t, repo, execution.ID)

	approvalReq := server.ApproveExecutionRequest{
		Approver: "admin_user",
		Decision: "approved",
		Note:     "test approval",
	}

	approval, ok := repo.ApproveExecution(execution.ID, approvalReq)
	if !ok {
		t.Fatalf("expected ApproveExecution to return ok=true")
	}

	if approval.Status != "approved" {
		t.Errorf("expected approval status 'approved', got '%s'", approval.Status)
	}

	updatedExec, ok := repo.GetExecution(execution.ID)
	if !ok {
		t.Fatalf("expected to retrieve execution after approval")
	}
	if updatedExec.Status != "running" {
		t.Errorf("expected execution status 'running' after approval, got '%s'", updatedExec.Status)
	}
	if updatedExec.CurrentStepID != "resumed_after_approval" {
		t.Errorf("expected current step 'resumed_after_approval', got '%s'", updatedExec.CurrentStepID)
	}
}

func TestListSkillsWithStatus(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	skill1 := createTestSkill(t, repo, "test_skill_status_a", "team_a", "low")
	defer cleanupSkill(t, repo, skill1.ID)

	skill2 := createTestSkill(t, repo, "test_skill_status_b", "team_b", "medium")
	defer cleanupSkill(t, repo, skill2.ID)

	_, err := repo.db.Exec(`UPDATE skills SET status = 'published' WHERE id = $1`, skill2.ID)
	if err != nil {
		t.Fatalf("failed to update skill status: %v", err)
	}

	draftSkills := repo.ListSkills("draft")
	foundDraft := false
	for _, s := range draftSkills {
		if s.ID == skill1.ID {
			foundDraft = true
			break
		}
	}
	if !foundDraft {
		t.Errorf("expected skill %s to be in draft list", skill1.ID)
	}
	for _, s := range draftSkills {
		if s.ID == skill2.ID && s.Status == "draft" {
			t.Errorf("expected skill %s NOT to be in draft list", skill2.ID)
		}
	}

	publishedSkills := repo.ListSkills("published")
	foundPublished := false
	for _, s := range publishedSkills {
		if s.ID == skill2.ID {
			foundPublished = true
			break
		}
	}
	if !foundPublished {
		t.Errorf("expected skill %s to be in published list", skill2.ID)
	}

	allSkills := repo.ListSkills("")
	if len(allSkills) < 2 {
		t.Errorf("expected at least 2 skills in total list, got %d", len(allSkills))
	}
}

func TestDashboardAggregatesMetrics(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	skill := createTestSkill(t, repo, "test_skill_dashboard", "test_team", "low")
	defer cleanupSkill(t, repo, skill.ID)

	_, err := repo.db.Exec(`UPDATE skills SET status = 'published', current_version = 'v1' WHERE id = $1`, skill.ID)
	if err != nil {
		t.Fatalf("failed to update skill: %v", err)
	}

	execReq := server.CreateExecutionRequest{
		SkillID:     skill.ID,
		TriggeredBy: "test_user",
		Input:       map[string]any{},
	}
	execution, ok := repo.CreateExecution(execReq)
	if !ok {
		t.Fatalf("failed to create execution for dashboard test")
	}
	defer cleanupExecution(t, repo, execution.ID)

	_, err = repo.db.Exec(`UPDATE executions SET status = 'running' WHERE id = $1`, execution.ID)
	if err != nil {
		t.Fatalf("failed to update execution status: %v", err)
	}

	dashboard := repo.Dashboard()

	if dashboard.Summary.ActiveSkills < 1 {
		t.Errorf("expected at least 1 active skill, got %d", dashboard.Summary.ActiveSkills)
	}

	if dashboard.Summary.RunningExecutions < 1 {
		t.Errorf("expected at least 1 running execution, got %d", dashboard.Summary.RunningExecutions)
	}

	foundExec := false
	for _, exec := range dashboard.RecentExecutions {
		if exec.ID == execution.ID {
			foundExec = true
			break
		}
	}
	if !foundExec {
		t.Errorf("expected created execution %s to appear in recent executions", execution.ID)
	}
}

// TestRejectExecutionUpdatesStatus tests that rejection updates execution status correctly.
func TestRejectExecutionUpdatesStatus(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	skill := createTestSkill(t, repo, "test_skill_reject", "test_team", "high")
	defer cleanupSkill(t, repo, skill.ID)

	execReq := server.CreateExecutionRequest{
		SkillID:     skill.ID,
		TriggeredBy: "test_user",
		Input:       map[string]any{"server_id": "srv-test"},
	}
	execution, ok := repo.CreateExecution(execReq)
	if !ok {
		t.Fatalf("expected successful execution creation")
	}
	defer cleanupExecution(t, repo, execution.ID)

	rejectReq := server.ApproveExecutionRequest{
		Approver: "admin_user",
		Decision: "rejected",
		Note:     "rejected for testing",
	}

	approval, ok := repo.ApproveExecution(execution.ID, rejectReq)
	if !ok {
		t.Fatalf("expected ApproveExecution to succeed for rejection")
	}

	if approval.Status != "rejected" {
		t.Errorf("expected approval status 'rejected', got '%s'", approval.Status)
	}

	updatedExec, ok := repo.GetExecution(execution.ID)
	if !ok {
		t.Fatalf("expected to retrieve execution after rejection")
	}
	if updatedExec.Status != "failed" {
		t.Errorf("expected execution status 'failed' after rejection, got '%s'", updatedExec.Status)
	}
	if updatedExec.CurrentStepID != "approval_rejected" {
		t.Errorf("expected current step 'approval_rejected', got '%s'", updatedExec.CurrentStepID)
	}
}

// TestGetSkillNotFound tests that GetSkill returns false for nonexistent skills.
func TestGetSkillNotFound(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	_, ok := repo.GetSkill("nonexistent_skill_id")
	if ok {
		t.Errorf("expected GetSkill to return false for nonexistent skill")
	}
}

// TestGetExecutionNotFound tests that GetExecution returns false for nonexistent executions.
func TestGetExecutionNotFound(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	_, ok := repo.GetExecution("nonexistent_exec_id")
	if ok {
		t.Errorf("expected GetExecution to return false for nonexistent execution")
	}
}

// TestListExecutionsWithStatus tests filtering executions by status.
func TestListExecutionsWithStatus(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	skill := createTestSkill(t, repo, "test_skill_exec_filter", "test_team", "low")
	defer cleanupSkill(t, repo, skill.ID)

	exec1, ok := repo.CreateExecution(server.CreateExecutionRequest{
		SkillID:     skill.ID,
		TriggeredBy: "user1",
		Input:       map[string]any{},
	})
	if !ok {
		t.Fatalf("failed to create first execution")
	}
	defer cleanupExecution(t, repo, exec1.ID)

	exec2, ok := repo.CreateExecution(server.CreateExecutionRequest{
		SkillID:     skill.ID,
		TriggeredBy: "user2",
		Input:       map[string]any{},
	})
	if !ok {
		t.Fatalf("failed to create second execution")
	}
	defer cleanupExecution(t, repo, exec2.ID)

	_, err := repo.db.Exec(`UPDATE executions SET status = 'succeeded' WHERE id = $1`, exec1.ID)
	if err != nil {
		t.Fatalf("failed to update execution status: %v", err)
	}

	pendingExecs := repo.ListExecutions("pending")
	foundPending := false
	for _, e := range pendingExecs {
		if e.ID == exec2.ID {
			foundPending = true
			break
		}
	}
	if !foundPending {
		t.Errorf("expected execution %s to be in pending list", exec2.ID)
	}

	succeededExecs := repo.ListExecutions("succeeded")
	foundSucceeded := false
	for _, e := range succeededExecs {
		if e.ID == exec1.ID {
			foundSucceeded = true
			break
		}
	}
	if !foundSucceeded {
		t.Errorf("expected execution %s to be in succeeded list", exec1.ID)
	}
}

// TestCreateExecutionWithValidSkill tests successful execution creation.
func TestCreateExecutionWithValidSkill(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	skill := createTestSkill(t, repo, "test_skill_exec", "test_team", "medium")
	defer cleanupSkill(t, repo, skill.ID)

	exec, ok := repo.CreateExecution(server.CreateExecutionRequest{
		SkillID:     skill.ID,
		TriggeredBy: "test_operator",
		Input:       map[string]any{"env": "test"},
	})
	if !ok {
		t.Fatalf("expected CreateExecution to succeed with valid skill")
	}
	defer cleanupExecution(t, repo, exec.ID)

	if exec.SkillID != skill.ID {
		t.Errorf("expected execution skill ID '%s', got '%s'", skill.ID, exec.SkillID)
	}
	if exec.SkillName != skill.Name {
		t.Errorf("expected execution skill name '%s', got '%s'", skill.Name, exec.SkillName)
	}
	if exec.Status != "pending" {
		t.Errorf("expected execution status 'pending', got '%s'", exec.Status)
	}
	if exec.TriggeredBy != "test_operator" {
		t.Errorf("expected triggered by 'test_operator', got '%s'", exec.TriggeredBy)
	}
}
