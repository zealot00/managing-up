package context

import "time"

type ContextID string

type Context interface {
	ID() ContextID
	Type() ContextType
	CreatedAt() time.Time
	UpdatedAt() time.Time
}

type ContextType string

const (
	TypeConversation ContextType = "conversation"
	TypeMemory       ContextType = "memory"
	TypeWorking      ContextType = "working"
)

type BaseContext struct {
	id        ContextID
	ctype     ContextType
	createdAt time.Time
	updatedAt time.Time
}

func (b *BaseContext) ID() ContextID        { return b.id }
func (b *BaseContext) Type() ContextType    { return b.ctype }
func (b *BaseContext) CreatedAt() time.Time { return b.createdAt }
func (b *BaseContext) UpdatedAt() time.Time { return b.updatedAt }

func (b *BaseContext) touch() {
	b.updatedAt = time.Now()
}
