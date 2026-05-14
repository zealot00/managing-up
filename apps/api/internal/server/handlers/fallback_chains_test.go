package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/gateway"
	"github.com/zealot/managing-up/apps/api/internal/llm"
)

// --- Mocks ---

type mockFallbackChainRepo struct {
	chains map[string]FallbackChainDTO
}

func newMockRepo() *mockFallbackChainRepo {
	return &mockFallbackChainRepo{chains: make(map[string]FallbackChainDTO)}
}

func (m *mockFallbackChainRepo) ListFallbackChains() ([]FallbackChainDTO, error) {
	var result []FallbackChainDTO
	for _, c := range m.chains {
		result = append(result, c)
	}
	return result, nil
}

func (m *mockFallbackChainRepo) GetFallbackChain(id string) (FallbackChainDTO, bool, error) {
	c, ok := m.chains[id]
	return c, ok, nil
}

func (m *mockFallbackChainRepo) CreateFallbackChain(chain FallbackChainDTO) (FallbackChainDTO, error) {
	chain.ID = "test-id-" + chain.Model
	chain.CreatedAt = time.Now()
	chain.UpdatedAt = time.Now()
	for i := range chain.Targets {
		chain.Targets[i].ID = "target-" + chain.Targets[i].Provider
		chain.Targets[i].ChainID = chain.ID
	}
	m.chains[chain.ID] = chain
	return chain, nil
}

func (m *mockFallbackChainRepo) UpdateFallbackChain(chain FallbackChainDTO) (FallbackChainDTO, error) {
	if _, ok := m.chains[chain.ID]; !ok {
		return FallbackChainDTO{}, fmt.Errorf("not found")
	}
	chain.UpdatedAt = time.Now()
	m.chains[chain.ID] = chain
	return chain, nil
}

func (m *mockFallbackChainRepo) DeleteFallbackChain(id string) error {
	delete(m.chains, id)
	return nil
}

type mockReloader struct {
	lastChains map[llm.Model][]gateway.FallbackTarget
}

func (m *mockReloader) SetFallbackChains(chains map[llm.Model][]gateway.FallbackTarget) {
	m.lastChains = chains
}

func TestFallbackChainHandler_List(t *testing.T) {
	repo := newMockRepo()
	repo.CreateFallbackChain(FallbackChainDTO{Model: "gpt-4o", IsEnabled: true, Targets: []FallbackTargetDTO{
		{Provider: "anthropic", Model: "claude-sonnet-4", Priority: 0, IsEnabled: true},
	}})

	reloader := &mockReloader{}
	h := NewFallbackChainHandler(repo, reloader)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestFallbackChainHandler_Create(t *testing.T) {
	repo := newMockRepo()
	reloader := &mockReloader{}
	h := NewFallbackChainHandler(repo, reloader)

	body, _ := json.Marshal(createFallbackChainRequest{
		Model:     "gpt-4o",
		IsEnabled: ptrBool(true),
		Targets: []FallbackTargetDTO{
			{Provider: "anthropic", Model: "claude-sonnet-4", Priority: 0, IsEnabled: true},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Verify reload was called
	if reloader.lastChains == nil {
		t.Fatal("expected reloader to be called")
	}
	if _, ok := reloader.lastChains[llm.Model("gpt-4o")]; !ok {
		t.Fatal("expected gpt-4o in reloaded chains")
	}
}

func TestFallbackChainHandler_Create_MissingModel(t *testing.T) {
	repo := newMockRepo()
	reloader := &mockReloader{}
	h := NewFallbackChainHandler(repo, reloader)

	body, _ := json.Marshal(createFallbackChainRequest{})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestFallbackChainHandler_Update(t *testing.T) {
	repo := newMockRepo()
	created, _ := repo.CreateFallbackChain(FallbackChainDTO{Model: "gpt-4o", IsEnabled: true})

	reloader := &mockReloader{}
	h := NewFallbackChainHandler(repo, reloader)

	body, _ := json.Marshal(updateFallbackChainRequest{
		Targets: []FallbackTargetDTO{
			{Provider: "anthropic", Model: "claude-sonnet-4", Priority: 0, IsEnabled: true},
			{Provider: "ollama", Model: "qwen2.5", Priority: 1, IsEnabled: true},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/"+created.ID, bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify reload
	if len(reloader.lastChains) == 0 {
		t.Fatal("expected reloader to be called")
	}
}

func TestFallbackChainHandler_Delete(t *testing.T) {
	repo := newMockRepo()
	created, _ := repo.CreateFallbackChain(FallbackChainDTO{Model: "gpt-4o", IsEnabled: true})

	reloader := &mockReloader{}
	h := NewFallbackChainHandler(repo, reloader)

	req := httptest.NewRequest(http.MethodDelete, "/"+created.ID, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify deleted
	_, found, _ := repo.GetFallbackChain(created.ID)
	if found {
		t.Fatal("expected chain to be deleted")
	}
}

func TestFallbackChainHandler_Get_NotFound(t *testing.T) {
	repo := newMockRepo()
	reloader := &mockReloader{}
	h := NewFallbackChainHandler(repo, reloader)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestFallbackChain_ReloadChains_DisabledChain(t *testing.T) {
	repo := newMockRepo()
	repo.CreateFallbackChain(FallbackChainDTO{Model: "gpt-4o", IsEnabled: false, Targets: []FallbackTargetDTO{
		{Provider: "anthropic", Model: "claude-sonnet-4", IsEnabled: true},
	}})

	reloader := &mockReloader{}
	h := NewFallbackChainHandler(repo, reloader)

	h.reloadChains()

	// Disabled chain should not be in the reloaded chains
	if len(reloader.lastChains) != 0 {
		t.Fatal("expected no chains when all are disabled")
	}
}

func TestFallbackChain_ReloadChains_DisabledTarget(t *testing.T) {
	repo := newMockRepo()
	repo.CreateFallbackChain(FallbackChainDTO{Model: "gpt-4o", IsEnabled: true, Targets: []FallbackTargetDTO{
		{Provider: "anthropic", Model: "claude-sonnet-4", IsEnabled: false},
		{Provider: "ollama", Model: "qwen2.5", IsEnabled: true},
	}})

	reloader := &mockReloader{}
	h := NewFallbackChainHandler(repo, reloader)

	h.reloadChains()

	chain := reloader.lastChains[llm.Model("gpt-4o")]
	if len(chain) != 1 {
		t.Fatalf("expected 1 target (disabled one filtered), got %d", len(chain))
	}
	if chain[0].Provider != llm.ProviderOllama {
		t.Errorf("expected ollama, got %s", chain[0].Provider)
	}
}

func ptrBool(b bool) *bool { return &b }
