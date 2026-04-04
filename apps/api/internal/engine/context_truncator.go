package engine

const (
	MaxContextTokens = 128000
	SafetyWaterline  = 0.85
	SummaryTrigger   = 0.90
)

type ContextTruncator struct {
	tokenizer Tokenizer
	maxTokens int
}

func NewContextTruncator(tokenizer Tokenizer, maxTokens int) *ContextTruncator {
	return &ContextTruncator{
		tokenizer: tokenizer,
		maxTokens: maxTokens,
	}
}

func (ct *ContextTruncator) TruncateIfNeeded(messages []Message) ([]Message, bool, error) {
	totalTokens := ct.tokenizer.CountMessages(messages)
	safeLimit := int(float64(ct.maxTokens) * SafetyWaterline)

	if totalTokens <= safeLimit {
		return messages, false, nil
	}

	truncated, err := ct.truncateToTokenCount(messages, safeLimit)
	return truncated, true, err
}

func (ct *ContextTruncator) truncateToTokenCount(messages []Message, maxTokens int) ([]Message, error) {
	result := make([]Message, 0, len(messages))

	var systemMsg Message
	otherMessages := make([]Message, 0)

	for _, msg := range messages {
		if msg.Role == "system" {
			systemMsg = msg
		} else {
			otherMessages = append(otherMessages, msg)
		}
	}

	currentTokens := 0
	if systemMsg.Content != "" {
		currentTokens += ct.tokenizer.Count(systemMsg.Content)
		result = append(result, systemMsg)
	}

	for i := len(otherMessages) - 1; i >= 0; i-- {
		msg := otherMessages[i]
		msgTokens := ct.tokenizer.Count(msg.Content) + 10

		if currentTokens+msgTokens <= maxTokens {
			result = append([]Message{msg}, result[1:]...)
			result[0] = systemMsg
			currentTokens += msgTokens
		} else {
			break
		}
	}

	for len(result) > 2 {
		if ct.tokenizer.CountMessages(result) <= maxTokens {
			break
		}
		result = result[1:]
	}

	return result, nil
}

func (ct *ContextTruncator) NeedsSummarization(messages []Message) bool {
	totalTokens := ct.tokenizer.CountMessages(messages)
	threshold := int(float64(ct.maxTokens) * SummaryTrigger)
	return totalTokens > threshold
}

func (ct *ContextTruncator) Summarize(messages []Message, summarizeFn func([]Message) (string, error)) ([]Message, error) {
	if !ct.NeedsSummarization(messages) {
		return messages, nil
	}

	systemMsg := Message{}
	var recentMessages []Message
	var oldMessages []Message

	for i, msg := range messages {
		if msg.Role == "system" {
			systemMsg = msg
		} else if i < len(messages)-5 {
			oldMessages = append(oldMessages, msg)
		} else {
			recentMessages = append(recentMessages, msg)
		}
	}

	summary, err := summarizeFn(oldMessages)
	if err != nil {
		return nil, err
	}

	result := []Message{systemMsg}
	result = append(result, Message{
		Role:    "system",
		Content: "[Earlier conversation summarized: " + summary + "]",
	})
	result = append(result, recentMessages...)

	return result, nil
}
