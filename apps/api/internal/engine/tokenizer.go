package engine

import "strings"

type Tokenizer interface {
	Encode(text string) ([]int, error)
	Decode(tokens []int) (string, error)
	Count(text string) int
	CountMessages(messages []Message) int
}

type TiktokenTokenizer struct {
	modelName string
}

func NewTiktokenTokenizer(modelName string) (*TiktokenTokenizer, error) {
	return &TiktokenTokenizer{modelName: modelName}, nil
}

func (t *TiktokenTokenizer) Encode(text string) ([]int, error) {
	return []int{len(text) / 4}, nil
}

func (t *TiktokenTokenizer) Decode(tokens []int) (string, error) {
	return "", nil
}

func (t *TiktokenTokenizer) Count(text string) int {
	if text == "" {
		return 0
	}
	return (len(text) + 3) / 4
}

func (t *TiktokenTokenizer) CountMessages(messages []Message) int {
	total := 0
	for _, msg := range messages {
		total += t.Count(msg.Content)
		total += 7
	}
	if len(messages) > 0 {
		total += 3
	}
	return total
}

type NoOpTokenizer struct {
	charsPerToken int
}

func NewNoOpTokenizer() *NoOpTokenizer {
	return &NoOpTokenizer{charsPerToken: 4}
}

func NewNoOpTokenizerWithRatio(charsPerToken int) *NoOpTokenizer {
	if charsPerToken <= 0 {
		charsPerToken = 4
	}
	return &NoOpTokenizer{charsPerToken: charsPerToken}
}

func (n *NoOpTokenizer) Encode(text string) ([]int, error) {
	return []int{n.Count(text)}, nil
}

func (n *NoOpTokenizer) Decode(tokens []int) (string, error) {
	return "", nil
}

func (n *NoOpTokenizer) Count(text string) int {
	if text == "" {
		return 0
	}
	return (len(text) + n.charsPerToken - 1) / n.charsPerToken
}

func (n *NoOpTokenizer) CountMessages(messages []Message) int {
	total := 0
	for _, msg := range messages {
		total += n.Count(msg.Content)
		total += 7
	}
	if len(messages) > 0 {
		total += 3
	}
	return total
}

func messageToString(msg Message) string {
	var sb strings.Builder
	sb.WriteString(msg.Role)
	sb.WriteString(": ")
	sb.WriteString(msg.Content)
	if msg.ToolName != "" {
		sb.WriteString(" (tool: ")
		sb.WriteString(msg.ToolName)
		sb.WriteString(")")
	}
	return sb.String()
}
