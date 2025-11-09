package http

import "strings"

type (
	DebugMessage struct {
		Link    string `json:"link"`
		Message string `json:"message"`
		Type    string `json:"type"`
	}

	DebugLogLevel string

	DebugDetails struct {
		Messages []DebugMessage `json:"messages,omitempty"`
	}
)

const (
	DebugLogLevelInfo    DebugLogLevel = "info"
	DebugLogLevelAll     DebugLogLevel = "all"
	DebugLogLevelWarning DebugLogLevel = "warning"
	DebugLogLevelNone    DebugLogLevel = "none"
)

func WithRequestDebugLogLevel[T any](level DebugLogLevel) RequestOption[T] {
	return func(request *Request[T]) {
		request.debugLogLevel = level
	}
}

// ParseDebugLogLevel parses the debug log level from a string.
func ParseDebugLogLevel(level string) DebugLogLevel {
	level = strings.TrimSpace(strings.ToLower(level))
	switch level {
	case "info":
		return DebugLogLevelInfo
	case "all":
		return DebugLogLevelAll
	case "warning":
		return DebugLogLevelWarning
	default:
		return DebugLogLevelNone
	}
}
