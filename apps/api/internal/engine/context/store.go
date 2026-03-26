package context

import "fmt"

type Store interface {
	Save(ctx Context) error
	Load(id ContextID, ctype ContextType) (Context, error)
	Delete(id ContextID) error
	List(ctype ContextType) ([]ContextID, error)
}

type MemoryStore struct {
	contexts map[ContextID]Context
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		contexts: make(map[ContextID]Context),
	}
}

func (s *MemoryStore) Save(ctx Context) error {
	s.contexts[ctx.ID()] = ctx
	return nil
}

func (s *MemoryStore) Load(id ContextID, ctype ContextType) (Context, error) {
	ctx, ok := s.contexts[id]
	if !ok {
		return nil, fmt.Errorf("context not found: %s", id)
	}
	if ctx.Type() != ctype {
		return nil, fmt.Errorf("context type mismatch: expected %s, got %s", ctype, ctx.Type())
	}
	return ctx, nil
}

func (s *MemoryStore) Delete(id ContextID) error {
	delete(s.contexts, id)
	return nil
}

func (s *MemoryStore) List(ctype ContextType) ([]ContextID, error) {
	var ids []ContextID
	for id, ctx := range s.contexts {
		if ctype == "" || ctx.Type() == ctype {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

type AgentContext struct {
	Conversation *ConversationContext
	Memory       *MemoryContext
	Working      *WorkingContext
}

func NewAgentContext(agentID string) *AgentContext {
	id := ContextID(agentID)
	return &AgentContext{
		Conversation: NewConversationContext(id + "_conv"),
		Memory:       NewMemoryContext(id + "_mem"),
		Working:      NewWorkingContext(id + "_work"),
	}
}
