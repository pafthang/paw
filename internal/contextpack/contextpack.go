package contextpack

import (
	"strings"
	"unicode/utf8"

	"github.com/pafthang/paw/internal/llm"
)

const DefaultSystemPrompt = "You are Paw, a local coding assistant. Be concise, practical, and preserve useful session context."
const DefaultMaxContextChars = 24000
const MinContextChars = 1000

func Pack(systemPrompt string, history []llm.Message, incoming []llm.Message, maxChars int) []llm.Message {
	if maxChars <= 0 {
		maxChars = DefaultMaxContextChars
	}
	if maxChars < MinContextChars {
		maxChars = MinContextChars
	}
	if strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = DefaultSystemPrompt
	}

	result := []llm.Message{{Role: "system", Content: systemPrompt}}
	incomingChars := messagesLen(incoming)
	budget := maxChars - messageLen(result[0]) - incomingChars
	if budget < 0 {
		budget = 0
	}

	keptHistory := make([]llm.Message, 0, len(history))
	used := 0
	for i := len(history) - 1; i >= 0; i-- {
		msg := normalize(history[i])
		if msg.Role == "" || msg.Content == "" {
			continue
		}
		cost := messageLen(msg)
		if used+cost > budget && len(keptHistory) > 0 {
			break
		}
		if used+cost > budget && len(keptHistory) == 0 {
			continue
		}
		used += cost
		keptHistory = append(keptHistory, msg)
	}
	for i := len(keptHistory) - 1; i >= 0; i-- {
		result = append(result, keptHistory[i])
	}
	for _, msg := range incoming {
		msg = normalize(msg)
		if msg.Role == "" || msg.Content == "" {
			continue
		}
		result = append(result, msg)
	}
	return result
}

func Stats(messages []llm.Message) map[string]any {
	return map[string]any{
		"messages": len(messages),
		"chars":    messagesLen(messages),
	}
}

func messagesLen(messages []llm.Message) int {
	total := 0
	for _, msg := range messages {
		total += messageLen(msg)
	}
	return total
}

func messageLen(msg llm.Message) int {
	return utf8.RuneCountInString(msg.Role) + utf8.RuneCountInString(msg.Content) + 8
}

func normalize(msg llm.Message) llm.Message {
	msg.Role = strings.TrimSpace(msg.Role)
	msg.Content = strings.TrimSpace(msg.Content)
	return msg
}
