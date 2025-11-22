//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the “Software”), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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
