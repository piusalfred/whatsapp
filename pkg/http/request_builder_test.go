//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package http_test

import (
	"net/http"
	"strings"
	"testing"

	gcmp "github.com/google/go-cmp/cmp"

	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/types"
)

func TestRequestBuilder_FluentBuild(t *testing.T) {
	t.Parallel()

	b := whttp.NewRequestBuilder("POST", "https://api.example.com").
		WithRequestType(whttp.RequestTypeSendMessage).
		WithEndpoints("v1", "messages").
		WithBearer("test-token").
		WithHeaders(map[string]string{"X-Custom": "value"}).
		WithQueryParams(map[string]string{"page": "1"}).
		WithAppSecret("shh", true).
		WithDebugLogLevel(whttp.DebugLogLevelAll).
		WithMetadata(types.Metadata{"key": "val"})

	msg := &TestMessage{Name: "hello", Value: 42}
	req := whttp.BuildRequest(b, msg)

	if req.Type != whttp.RequestTypeSendMessage {
		t.Errorf("Type = %v, want RequestTypeSendMessage", req.Type)
	}
	if req.Method != http.MethodPost {
		t.Errorf("Method = %q, want POST", req.Method)
	}
	if req.BaseURL != "https://api.example.com" {
		t.Errorf("BaseURL = %q, want https://api.example.com", req.BaseURL)
	}
	wantEndpoints := []string{"v1", "messages"}
	if diff := gcmp.Diff(wantEndpoints, req.Endpoints); diff != "" {
		t.Errorf("Endpoints mismatch (-want +got):\n%s", diff)
	}
	if req.Bearer != "test-token" {
		t.Errorf("Bearer = %q, want test-token", req.Bearer)
	}
	if req.Headers["X-Custom"] != "value" {
		t.Errorf("Headers mismatch")
	}
	if req.QueryParams["page"] != "1" {
		t.Errorf("QueryParams mismatch")
	}
	if req.AppSecret != "shh" {
		t.Errorf("AppSecret = %q, want shh", req.AppSecret)
	}
	if !req.SecureRequests {
		t.Error("SecureRequests should be true")
	}
	url, err := req.URL()
	if err != nil {
		t.Fatalf("URL() error: %v", err)
	}
	if !strings.Contains(url, "debug=all") {
		t.Errorf("expected URL to contain debug=all, got %s", url)
	}
	if diff := gcmp.Diff(types.Metadata{"key": "val"}, req.Metadata); diff != "" {
		t.Errorf("Metadata mismatch (-want +got):\n%s", diff)
	}
	if req.Message != msg {
		t.Error("Message pointer mismatch")
	}
}

func TestRequestBuilder_AnyRequest(t *testing.T) {
	t.Parallel()

	b := whttp.NewRequestBuilder("GET", "https://api.example.com").
		WithEndpoints("v1", "users").
		WithBearer("token")

	req := whttp.BuildAnyRequest(b)

	if req.Method != http.MethodGet {
		t.Errorf("Method = %q, want GET", req.Method)
	}
	if req.Bearer != "token" {
		t.Errorf("Bearer = %q, want token", req.Bearer)
	}
	if req.Message != nil {
		t.Error("expected nil Message for any request")
	}
}

func TestRequestBuilder_DownloadURL(t *testing.T) {
	t.Parallel()

	b := whttp.NewRequestBuilder("GET", "ignored").
		WithDownloadURL("https://cdn.example.com/file.pdf")

	req := whttp.BuildAnyRequest(b)

	url, err := req.URL()
	if err != nil {
		t.Fatalf("URL() error: %v", err)
	}
	if url != "https://cdn.example.com/file.pdf" {
		t.Errorf("URL = %q, want https://cdn.example.com/file.pdf", url)
	}
}

func TestRequestBuilder_FormAndBodyReader(t *testing.T) {
	t.Parallel()

	form := &whttp.RequestForm{
		Fields: map[string]string{"field": "value"},
	}

	b := whttp.NewRequestBuilder("POST", "https://api.example.com").
		WithForm(form).
		WithBodyReader(strings.NewReader("raw"))

	req := whttp.BuildAnyRequest(b)

	if req.Form != form {
		t.Error("Form mismatch")
	}
	if req.BodyReader == nil {
		t.Error("expected BodyReader to be set")
	}
}

func TestRequestBuilder_WithSecured(t *testing.T) {
	t.Parallel()

	b := whttp.NewRequestBuilder("GET", "https://api.example.com").
		WithSecured(true)

	req := whttp.BuildAnyRequest(b)
	if !req.SecureRequests {
		t.Error("expected SecureRequests to be true")
	}
}
