package gateway

import "strings"

func ParseClientName(ua string) string {
	if ua == "" {
		return "unknown"
	}
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "python"):
		return "python"
	case strings.Contains(ua, "node"):
		return "node"
	case strings.Contains(ua, "curl"):
		return "curl"
	case strings.Contains(ua, "opencode/"):
		return "opencode"
	case strings.Contains(ua, "opencode"):
		return "opencode"
	case strings.Contains(ua, "go-http-client"), strings.Contains(ua, "go-built"):
		return "go"
	case strings.Contains(ua, "anthropic"):
		return "anthropic-sdk"
	case strings.Contains(ua, "openai"):
		return "openai-sdk"
	case strings.Contains(ua, "langchain"):
		return "langchain"
	case strings.Contains(ua, "lanchain"):
		return "langchain"
	default:
		if idx := strings.Index(ua, "/"); idx > 0 {
			return ua[:idx]
		}
		if idx := strings.Index(ua, " "); idx > 0 {
			return ua[:idx]
		}
		return "other"
	}
}
