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

package http_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/piusalfred/whatsapp/internal/test"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func TestParseDebugLogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  whttp.DebugLogLevel
	}{
		{
			name:  "info",
			input: "info",
			want:  whttp.DebugLogLevelInfo,
		},
		{
			name:  "INFO with spaces",
			input: "  INFO  ",
			want:  whttp.DebugLogLevelInfo,
		},
		{
			name:  "all",
			input: "all",
			want:  whttp.DebugLogLevelAll,
		},
		{
			name:  "ALL uppercase",
			input: "ALL",
			want:  whttp.DebugLogLevelAll,
		},
		{
			name:  "warning",
			input: "warning",
			want:  whttp.DebugLogLevelWarning,
		},
		{
			name:  "Warning mixed case",
			input: "Warning",
			want:  whttp.DebugLogLevelWarning,
		},
		{
			name:  "unknown defaults to none",
			input: "unknown",
			want:  whttp.DebugLogLevelNone,
		},
		{
			name:  "empty defaults to none",
			input: "",
			want:  whttp.DebugLogLevelNone,
		},
		{
			name:  "none",
			input: "none",
			want:  whttp.DebugLogLevelNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := whttp.ParseDebugLogLevel(tt.input)
			if got != tt.want {
				t.Errorf("ParseDebugLogLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
func TestDebugLogLevelInRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		level    whttp.DebugLogLevel
		expected string
	}{
		{
			name:     "info level",
			level:    whttp.DebugLogLevelInfo,
			expected: "debug=info",
		},
		{
			name:     "all level",
			level:    whttp.DebugLogLevelAll,
			expected: "debug=all",
		},
		{
			name:     "warning level",
			level:    whttp.DebugLogLevelWarning,
			expected: "debug=warning",
		},
		{
			name:     "none level",
			level:    whttp.DebugLogLevelNone,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := whttp.MakeRequest(
				http.MethodGet,
				"https://api.example.com",
				whttp.WithRequestDebugLogLevel[TestMessage](tt.level),
			)

			url, err := req.URL()
			test.AssertNoError(t, "URL", err)

			if tt.expected == "" {
				if strings.Contains(url, "debug=") {
					t.Errorf("expected URL not to contain debug param, got %s", url)
				}
			} else {
				if !strings.Contains(url, tt.expected) {
					t.Errorf("expected URL to contain %s, got %s", tt.expected, url)
				}
			}
		})
	}
}
