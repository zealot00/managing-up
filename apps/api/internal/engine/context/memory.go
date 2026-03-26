package context

import "time"

type MemoryContext struct {
	*BaseContext
	Facts       []Fact
	Preferences []Preference
	Learnings   []Learning
}

type Fact struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Preference struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Learning struct {
	Content   string    `json:"content"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
}

func NewMemoryContext(id ContextID) *MemoryContext {
	now := time.Now()
	return &MemoryContext{
		BaseContext: &BaseContext{
			id:        id,
			ctype:     TypeMemory,
			createdAt: now,
			updatedAt: now,
		},
		Facts:       []Fact{},
		Preferences: []Preference{},
		Learnings:   []Learning{},
	}
}

func (m *MemoryContext) AddFact(key, value string) {
	m.Facts = append(m.Facts, Fact{Key: key, Value: value})
	m.touch()
}

func (m *MemoryContext) GetFact(key string) (string, bool) {
	for _, f := range m.Facts {
		if f.Key == key {
			return f.Value, true
		}
	}
	return "", false
}

func (m *MemoryContext) SetPreference(key, value string) {
	for i, p := range m.Preferences {
		if p.Key == key {
			m.Preferences[i].Value = value
			m.touch()
			return
		}
	}
	m.Preferences = append(m.Preferences, Preference{Key: key, Value: value})
	m.touch()
}

func (m *MemoryContext) GetPreference(key string) (string, bool) {
	for _, p := range m.Preferences {
		if p.Key == key {
			return p.Value, true
		}
	}
	return "", false
}

func (m *MemoryContext) AddLearning(content, source string) {
	m.Learnings = append(m.Learnings, Learning{
		Content:   content,
		Source:    source,
		Timestamp: time.Now(),
	})
	m.touch()
}

func (m *MemoryContext) GetRecentLearnings(n int) []Learning {
	if n <= 0 {
		return m.Learnings
	}
	if n > len(m.Learnings) {
		n = len(m.Learnings)
	}
	return m.Learnings[len(m.Learnings)-n:]
}
