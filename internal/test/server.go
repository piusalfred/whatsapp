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
)

// MockServer wraps an httptest.Server and records the last incoming request
// so tests can assert on method, path, headers, query params and body.
type MockServer struct {
	Server      *httptest.Server
	LastRequest *http.Request
	LastBody    []byte
}

// NewMockServer starts a new server. The optional handler is called after the
// request has been recorded.
func NewMockServer(handler http.HandlerFunc) *MockServer {
	m := &MockServer{}
	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.LastRequest = r
		body, _ := io.ReadAll(r.Body)
		m.LastBody = body
		if handler != nil {
			handler(w, r)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	return m
}

// Close shuts down the server.
func (m *MockServer) Close() {
	m.Server.Close()
}
