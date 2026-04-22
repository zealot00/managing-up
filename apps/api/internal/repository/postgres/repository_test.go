package postgres

import (
	"context"
	"os"
	"testing"
	"time"

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

func cleanupSkillDependency(t *testing.T, repo *Repository, id string) {
	t.Helper()
	_, err := repo.db.Exec(`DELETE FROM skill_dependencies WHERE id = $1`, id)
	if err != nil {
		t.Logf("warning: failed to cleanup skill dependency %s: %v", id, err)
	}
}

func cleanupSkillInstall(t *testing.T, repo *Repository, id string) {
	t.Helper()
	_, err := repo.db.Exec(`DELETE FROM skill_installs WHERE id = $1`, id)
	if err != nil {
		t.Logf("warning: failed to cleanup skill install %s: %v", id, err)
	}
}

func cleanupSkillRating(t *testing.T, repo *Repository, skillID, userID string) {
	t.Helper()
	_, err := repo.db.Exec(`DELETE FROM skill_ratings WHERE skill_id = $1 AND user_id = $2`, skillID, userID)
	if err != nil {
		t.Logf("warning: failed to cleanup skill rating for skill %s user %s: %v", skillID, userID, err)
	}
}

func ensureTestUser(t *testing.T, repo *Repository, id string) {
	t.Helper()
	_, err := repo.db.Exec(`
		INSERT INTO users (id, username, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, 'hash', 'user', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, id, id)
	if err != nil {
		t.Fatalf("failed to ensure test user %s: %v", id, err)
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

func TestSkillEnterpriseQueries(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	ctx := context.Background()
	skillA := createTestSkill(t, repo, "test_skill_enterprise_a", "team_a", "medium")
	skillB := createTestSkill(t, repo, "test_skill_enterprise_b", "team_b", "low")
	defer cleanupSkill(t, repo, skillA.ID)
	defer cleanupSkill(t, repo, skillB.ID)

	ensureTestUser(t, repo, "enterprise-user")

	_, err := repo.db.Exec(`
		UPDATE skills
		SET status = 'published', current_version = 'v1', category = 'operations', trust_score = 0.93
		WHERE id = $1
	`, skillA.ID)
	if err != nil {
		t.Fatalf("failed to update skillA enterprise fields: %v", err)
	}
	_, err = repo.db.Exec(`
		UPDATE skills
		SET status = 'published', current_version = 'v2', category = 'observability', trust_score = 0.81
		WHERE id = $1
	`, skillB.ID)
	if err != nil {
		t.Fatalf("failed to update skillB enterprise fields: %v", err)
	}

	depID := "dep_enterprise_test"
	installID := "install_enterprise_test"
	defer cleanupSkillDependency(t, repo, depID)
	defer cleanupSkillInstall(t, repo, installID)
	defer cleanupSkillRating(t, repo, skillA.ID, "enterprise-user")

	_, err = repo.db.Exec(`
		INSERT INTO skill_dependencies (id, skill_id, dependency_skill_id, version_constraint, created_at)
		VALUES ($1, $2, $3, '>=v2', NOW())
	`, depID, skillA.ID, skillB.ID)
	if err != nil {
		t.Fatalf("failed to insert skill dependency: %v", err)
	}

	_, err = repo.db.Exec(`
		INSERT INTO skill_installs (id, skill_id, user_id, version, environment, installed_at, skill_snapshot)
		VALUES ($1, $2, $3, 'v1', 'production', NOW(), '{}'::jsonb)
	`, installID, skillA.ID, "enterprise-user")
	if err != nil {
		t.Fatalf("failed to insert skill install: %v", err)
	}

	if err := repo.UpsertRating(ctx, skillA.ID, "enterprise-user", 5, "excellent"); err != nil {
		t.Fatalf("failed to upsert rating: %v", err)
	}

	deps, err := repo.ListDependencies(ctx, skillA.ID)
	if err != nil {
		t.Fatalf("expected dependencies query to succeed: %v", err)
	}
	if len(deps) != 1 || deps[0].DependencySkillID != skillB.ID {
		t.Fatalf("expected dependency on %s, got %+v", skillB.ID, deps)
	}

	market, err := repo.ListSkillsByCategory(ctx, "operations", "enterprise")
	if err != nil {
		t.Fatalf("expected market query to succeed: %v", err)
	}
	if len(market) != 1 || market[0].ID != skillA.ID {
		t.Fatalf("expected market result for %s, got %+v", skillA.ID, market)
	}

	avg, count, err := repo.GetRatingStats(ctx, skillA.ID)
	if err != nil {
		t.Fatalf("expected rating stats query to succeed: %v", err)
	}
	if count != 1 || avg != 5 {
		t.Fatalf("expected avg=5 count=1, got avg=%v count=%d", avg, count)
	}

	installs, err := repo.GetInstallCount(ctx, skillA.ID)
	if err != nil {
		t.Fatalf("expected install count query to succeed: %v", err)
	}
	if installs != 1 {
		t.Fatalf("expected install count 1, got %d", installs)
	}

	tree, err := repo.ResolveDepTree(ctx, skillA.ID)
	if err != nil {
		t.Fatalf("expected dependency tree query to succeed: %v", err)
	}
	if len(tree) != 1 || tree[0].SkillID != skillB.ID {
		t.Fatalf("expected dependency tree root %s, got %+v", skillB.ID, tree)
	}

	if _, err := time.Parse(time.RFC3339, deps[0].CreatedAt); err != nil {
		t.Fatalf("expected dependency created_at to be RFC3339, got %q", deps[0].CreatedAt)
	}
}

func cleanupGatewaySession(t *testing.T, repo *Repository, id string) {
	t.Helper()
	_, err := repo.db.Exec(`DELETE FROM mcp_gateway_sessions WHERE id = $1`, id)
	if err != nil {
		t.Logf("warning: failed to cleanup gateway session %s: %v", id, err)
	}
}

func TestCreateGatewaySession(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	ctx := context.Background()
	session := &server.GatewaySession{
		ID:             "gs_test_create",
		SessionType:    "router",
		AgentID:        "agent_test",
		CorrelationID:  "corr_test_123",
		TaskIntent:     map[string]any{"task": "test_task"},
		RiskLevel:      "medium",
		PolicyDecision: map[string]any{"allowed": true},
		Status:         "active",
		StartedAt:      time.Now(),
		Metadata:       map[string]any{"key": "value"},
	}

	if err := repo.CreateGatewaySession(ctx, session); err != nil {
		t.Fatalf("expected CreateGatewaySession to succeed: %v", err)
	}
	defer cleanupGatewaySession(t, repo, session.ID)

	retrieved, err := repo.GetGatewaySession(ctx, session.ID)
	if err != nil {
		t.Fatalf("expected GetGatewaySession to succeed: %v", err)
	}
	if retrieved.SessionType != session.SessionType {
		t.Errorf("expected session type %q, got %q", session.SessionType, retrieved.SessionType)
	}
	if retrieved.AgentID != session.AgentID {
		t.Errorf("expected agent ID %q, got %q", session.AgentID, retrieved.AgentID)
	}
	if retrieved.CorrelationID != session.CorrelationID {
		t.Errorf("expected correlation ID %q, got %q", session.CorrelationID, retrieved.CorrelationID)
	}
	if retrieved.RiskLevel != session.RiskLevel {
		t.Errorf("expected risk level %q, got %q", session.RiskLevel, retrieved.RiskLevel)
	}
	if retrieved.Status != session.Status {
		t.Errorf("expected status %q, got %q", session.Status, retrieved.Status)
	}
}

func TestCreateGatewaySessionWithNilPolicyDecision(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	ctx := context.Background()
	session := &server.GatewaySession{
		ID:            "gs_test_nil_policy",
		SessionType:   "router",
		AgentID:       "agent_test",
		CorrelationID: "corr_nil_policy",
		TaskIntent:    map[string]any{"task": "test"},
		RiskLevel:     "low",
		Status:        "active",
		StartedAt:     time.Now(),
		Metadata:      map[string]any{},
	}

	if err := repo.CreateGatewaySession(ctx, session); err != nil {
		t.Fatalf("expected CreateGatewaySession with nil policy decision to succeed: %v", err)
	}
	defer cleanupGatewaySession(t, repo, session.ID)

	retrieved, err := repo.GetGatewaySession(ctx, session.ID)
	if err != nil {
		t.Fatalf("expected GetGatewaySession to succeed: %v", err)
	}
	if retrieved.PolicyDecision != nil {
		t.Errorf("expected nil policy decision, got %+v", retrieved.PolicyDecision)
	}
}

func TestGetGatewaySessionNotFound(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	ctx := context.Background()
	_, err := repo.GetGatewaySession(ctx, "nonexistent_session")
	if err == nil {
		t.Fatalf("expected GetGatewaySession to return error for nonexistent session")
	}
}

func TestEndGatewaySession(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	ctx := context.Background()
	session := &server.GatewaySession{
		ID:            "gs_test_end",
		SessionType:   "router",
		AgentID:       "agent_test",
		CorrelationID:  "corr_end_123",
		TaskIntent:    map[string]any{},
		RiskLevel:     "low",
		Status:        "active",
		StartedAt:     time.Now(),
		Metadata:      map[string]any{},
	}

	if err := repo.CreateGatewaySession(ctx, session); err != nil {
		t.Fatalf("expected CreateGatewaySession to succeed: %v", err)
	}
	defer cleanupGatewaySession(t, repo, session.ID)

	if err := repo.EndGatewaySession(ctx, session.ID); err != nil {
		t.Fatalf("expected EndGatewaySession to succeed: %v", err)
	}

	retrieved, err := repo.GetGatewaySession(ctx, session.ID)
	if err != nil {
		t.Fatalf("expected GetGatewaySession to succeed: %v", err)
	}
	if retrieved.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", retrieved.Status)
	}
	if retrieved.EndedAt == nil {
		t.Errorf("expected ended_at to be set, got nil")
	}
}

func TestUpdateGatewaySessionPolicyDecision(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	ctx := context.Background()
	session := &server.GatewaySession{
		ID:            "gs_test_update_policy",
		SessionType:   "router",
		AgentID:       "agent_test",
		CorrelationID:  "corr_policy_123",
		TaskIntent:    map[string]any{},
		RiskLevel:     "medium",
		Status:        "active",
		StartedAt:     time.Now(),
		Metadata:      map[string]any{},
	}

	if err := repo.CreateGatewaySession(ctx, session); err != nil {
		t.Fatalf("expected CreateGatewaySession to succeed: %v", err)
	}
	defer cleanupGatewaySession(t, repo, session.ID)

	newDecision := map[string]any{"allowed": true, "reason": "approved by policy"}
	if err := repo.UpdateGatewaySessionPolicyDecision(ctx, session.ID, newDecision); err != nil {
		t.Fatalf("expected UpdateGatewaySessionPolicyDecision to succeed: %v", err)
	}

	retrieved, err := repo.GetGatewaySession(ctx, session.ID)
	if err != nil {
		t.Fatalf("expected GetGatewaySession to succeed: %v", err)
	}
	if retrieved.PolicyDecision == nil {
		t.Fatalf("expected policy decision to be set, got nil")
	}
	allowed, ok := retrieved.PolicyDecision["allowed"].(bool)
	if !ok || !allowed {
		t.Errorf("expected policy decision allowed=true, got %+v", retrieved.PolicyDecision)
	}
}

func TestListGatewaySessions(t *testing.T) {
	t.Parallel()
	repo, cleanup := setupRepo(t)
	defer cleanup()

	ctx := context.Background()
	agentID := "agent_list_test"

	session1 := &server.GatewaySession{
		ID:            "gs_test_list_1",
		SessionType:   "router",
		AgentID:       agentID,
		CorrelationID:  "corr_list_1",
		TaskIntent:    map[string]any{"task": "test1"},
		RiskLevel:     "low",
		Status:        "active",
		StartedAt:     time.Now().Add(-2 * time.Hour),
		Metadata:      map[string]any{},
	}
	session2 := &server.GatewaySession{
		ID:            "gs_test_list_2",
		SessionType:   "router",
		AgentID:       agentID,
		CorrelationID:  "corr_list_2",
		TaskIntent:    map[string]any{"task": "test2"},
		RiskLevel:     "medium",
		Status:        "completed",
		StartedAt:     time.Now().Add(-1 * time.Hour),
		Metadata:      map[string]any{},
	}
	session3 := &server.GatewaySession{
		ID:            "gs_test_list_3",
		SessionType:   "router",
		AgentID:       "other_agent",
		CorrelationID:  "corr_list_3",
		TaskIntent:    map[string]any{"task": "test3"},
		RiskLevel:     "high",
		Status:        "active",
		StartedAt:     time.Now(),
		Metadata:      map[string]any{},
	}

	for _, s := range []*server.GatewaySession{session1, session2, session3} {
		if err := repo.CreateGatewaySession(ctx, s); err != nil {
			t.Fatalf("expected CreateGatewaySession to succeed: %v", err)
		}
		defer cleanupGatewaySession(t, repo, s.ID)
	}

	sessions, err := repo.ListGatewaySessions(ctx, agentID, 10)
	if err != nil {
		t.Fatalf("expected ListGatewaySessions to succeed: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions for agent %q, got %d", agentID, len(sessions))
	}

	allSessions, err := repo.ListGatewaySessions(ctx, "", 10)
	if err != nil {
		t.Fatalf("expected ListGatewaySessions with empty agent ID to succeed: %v", err)
	}
	if len(allSessions) < 3 {
		t.Errorf("expected at least 3 sessions total, got %d", len(allSessions))
	}

	if sessions[0].CorrelationID != session2.CorrelationID && sessions[1].CorrelationID != session2.CorrelationID {
		t.Errorf("expected sessions to be ordered by started_at DESC, got %+v", sessions)
	}
}
