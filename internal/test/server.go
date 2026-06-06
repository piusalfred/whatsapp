/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the "Software"), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
)

// CapturedRequest holds details about a single request received by MockServer.
type CapturedRequest struct {
	Method      string
	Path        string
	QueryParams url.Values
	Header      http.Header
	Body        []byte
}

// MockBehavior configures the response a MockServer returns.
type MockBehavior struct {
	StatusCode int
	Payload    []byte
	Headers    map[string]string
}

// MockServer wraps httptest.Server and records every incoming request so tests
// can assert on HTTP method, path, headers, query parameters and body.
type MockServer struct {
	Server   *httptest.Server
	Behavior MockBehavior
	mu       sync.Mutex
	Requests []CapturedRequest
}

// NewMockServer starts a new mock server with the given behavior.
func NewMockServer(behavior MockBehavior) *MockServer {
	m := &MockServer{
		Behavior: behavior,
		Requests: make([]CapturedRequest, 0),
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		m.mu.Lock()
		m.Requests = append(m.Requests, CapturedRequest{
			Method:      r.Method,
			Path:        r.URL.Path,
			QueryParams: r.URL.Query(),
			Header:      r.Header.Clone(),
			Body:        body,
		})
		m.mu.Unlock()

		for k, v := range m.Behavior.Headers {
			w.Header().Set(k, v)
		}

		statusCode := m.Behavior.StatusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		w.WriteHeader(statusCode)

		if m.Behavior.Payload != nil {
			_, _ = w.Write(m.Behavior.Payload)
		}
	})

	m.Server = httptest.NewServer(handler)
	return m
}

// GetRequests returns a copy of all captured requests.
func (m *MockServer) GetRequests() []CapturedRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	reqs := make([]CapturedRequest, len(m.Requests))
	copy(reqs, m.Requests)
	return reqs
}

// Close shuts down the server.
func (m *MockServer) Close() {
	m.Server.Close()
}
