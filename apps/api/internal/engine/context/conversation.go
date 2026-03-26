package context

import "time"

type ConversationContext struct {
	*BaseContext
	Messages    []Message
	MaxMessages int
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewConversationContext(id ContextID) *ConversationContext {
	now := time.Now()
	return &ConversationContext{
		BaseContext: &BaseContext{
			id:        id,
			ctype:     TypeConversation,
			createdAt: now,
			updatedAt: now,
		},
		Messages:    []Message{},
		MaxMessages: 100,
	}
}

func (c *ConversationContext) AddMessage(role, content string) {
	c.Messages = append(c.Messages, Message{Role: role, Content: content})
	c.touch()
	c.trim()
}

func (c *ConversationContext) trim() {
	if c.MaxMessages > 0 && len(c.Messages) > c.MaxMessages {
		c.Messages = c.Messages[len(c.Messages)-c.MaxMessages:]
	}
}

func (c *ConversationContext) GetMessages() []Message {
	return c.Messages
}

func (c *ConversationContext) Clear() {
	c.Messages = []Message{}
	c.touch()
}
