/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the "Software"), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package http

import (
	"net/http"
	"strings"
)

type (
	// DebugMessage is a single debug message returned by the WhatsApp API when
	// debug mode is enabled. Link points to related documentation; Message
	// contains a human-readable explanation.
	DebugMessage struct {
		Link    string `json:"link"`
		Message string `json:"message"`
		Type    string `json:"type"`
	}

	// DebugLogLevel controls the verbosity of debug output appended to API
	// responses. Set it on a Request to receive debug information.
	DebugLogLevel string

	// DebugHeaders holds debug information extracted from response headers.
	// These are separate from the body-parsed __debug__ messages and are
	// captured without any additional body I/O.
	DebugHeaders struct {
		FacebookAPIVersion string   `json:"facebook_api_version,omitempty"`
		FBTraceID          []string `json:"fbtrace_id,omitempty"`
		FBRev              []string `json:"fb_rev,omitempty"`
		FBDebug            []string `json:"fb_debug,omitempty"`
	}

	// DebugDetails is a container for debug messages returned in API responses
	// when a Request has a non-none DebugLogLevel. Messages are populated from
	// the __debug__ JSON body.
	DebugDetails struct {
		Messages []DebugMessage `json:"messages,omitempty"`
	}

	// DebugHeadersCapturer is implemented by response types that can receive
	// debug headers after the response body has been decoded. The decoder
	// calls OnDebugHeaders when the target implements this interface.
	DebugHeadersCapturer interface {
		OnDebugHeaders(h DebugHeaders)
	}
)

// debugHeadersFromResponse extracts debug headers from an HTTP response.
func debugHeadersFromResponse(resp *http.Response) DebugHeaders {
	return DebugHeaders{
		FacebookAPIVersion: resp.Header.Get("Facebook-Api-Version"),
		FBTraceID:          resp.Header.Values("X-Fb-Trace-Id"),
		FBRev:              resp.Header.Values("X-Fb-Rev"),
		FBDebug:            resp.Header.Values("X-Fb-Debug"),
	}
}

const (
	DebugLogLevelInfo    DebugLogLevel = "info"
	DebugLogLevelAll     DebugLogLevel = "all"
	DebugLogLevelWarning DebugLogLevel = "warning"
	DebugLogLevelNone    DebugLogLevel = "none"
)

// WithRequestDebugLogLevel sets the debug log level on a request.
func WithRequestDebugLogLevel[T any](level DebugLogLevel) RequestOption[T] {
	return func(request *Request[T]) {
		request.debugLogLevel = level
	}
}

// ParseDebugLogLevel parses a string into a DebugLogLevel. Parsing is
// case-insensitive and trims whitespace. Unrecognized values fall back
// to DebugLogLevelNone.
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
