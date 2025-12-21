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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	gcmp "github.com/google/go-cmp/cmp"

	"github.com/piusalfred/whatsapp/internal/test"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/types"
)

func TestRequestWithContext(t *testing.T) {
	t.Parallel()
	type testCase[T any] struct {
		name    string
		req     *whttp.Request[T]
		want    *http.Request
		wantErr bool
	}

	messageTests := []testCase[TestMessage]{
		{
			name: "successful POST request with JSON body",
			req: &whttp.Request[TestMessage]{
				Type:        whttp.RequestTypeSendMessage,
				Method:      http.MethodPost,
				Bearer:      "test-token",
				BaseURL:     "http://localhost/api/v1",
				Endpoints:   []string{"send-message"},
				Message:     &TestMessage{Name: "Test", Value: 123},
				Headers:     map[string]string{"X-Custom-Header": "custom-value"},
				QueryParams: map[string]string{"param1": "value1", "param2": "value2"},
			},
			want: &http.Request{
				Method: http.MethodPost,
				Header: http.Header{
					"Authorization":   []string{"Bearer test-token"},
					"X-Custom-Header": []string{"custom-value"},
					"Content-Type":    []string{"application/json"},
				},
				URL: func() *url.URL {
					u, _ := url.Parse("http://localhost/api/v1/send-message?param1=value1&param2=value2")

					return u
				}(),
				Host: "localhost",
				Body: io.NopCloser(bytes.NewReader([]byte(`{"name":"Test","value":123}`))),
			},
			wantErr: false,
		},
		{
			name: "successful GET request with no body",
			req: &whttp.Request[TestMessage]{
				Type:        whttp.RequestTypeGetPhoneNumber,
				Method:      http.MethodGet,
				Bearer:      "test-token",
				BaseURL:     "http://localhost/api/v1",
				Endpoints:   []string{"phone-number"},
				QueryParams: map[string]string{"param": "value"},
			},
			want: &http.Request{
				Method: http.MethodGet,
				Header: http.Header{
					"Authorization": []string{"Bearer test-token"},
				},
				URL: func() *url.URL {
					u, _ := url.Parse("http://localhost/api/v1/phone-number?param=value")

					return u
				}(),
				Host: "localhost",
				Body: nil,
			},
			wantErr: false,
		},
		{
			name: "successful multipart form request with file",
			req: &whttp.Request[TestMessage]{
				Type:      whttp.RequestTypeUploadMedia,
				Method:    http.MethodPost,
				Bearer:    "test-token",
				BaseURL:   "http://localhost/api/v1",
				Endpoints: []string{"upload"},
				Form: &whttp.RequestForm{
					Fields: map[string]string{
						"title": "example",
					},
					FormFile: &whttp.FormFile{
						Name: "file",
						Path: "testfile.txt",
					},
				},
				Headers: map[string]string{"X-Custom-Header": "custom-value"},
			},
			want: &http.Request{
				Method: http.MethodPost,
				Header: http.Header{
					"Authorization":   []string{"Bearer test-token"},
					"X-Custom-Header": []string{"custom-value"},
				},
				URL: func() *url.URL {
					u, _ := url.Parse("http://localhost/api/v1/upload")

					return u
				}(),
				Host: "localhost",
			},
			wantErr: false,
		},
	}

	for _, tt := range messageTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := whttp.RequestWithContext(t.Context(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequestWithContext() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if diff := gcmp.Diff(got.URL.String(), tt.want.URL.String()); diff != "" {
				t.Errorf("RequestWithContext() URL mismatch (-want +got):\n%s", diff)
			}
			if diff := gcmp.Diff(got.Method, tt.want.Method); diff != "" {
				t.Errorf("RequestWithContext() Method mismatch (-want +got):\n%s", diff)
			}
			for key := range tt.want.Header {
				if got.Header.Get(key) != tt.want.Header.Get(key) {
					t.Errorf("RequestWithContext() header %v mismatch, want %v,"+
						" got %v", key, tt.want.Header.Get(key), got.Header.Get(key))
				}
			}

			if tt.want.Body != nil {
				var gotBody TestMessage
				var wantBody TestMessage

				gotBodyBytes, err := io.ReadAll(got.Body)
				test.AssertNoError(t, "read body bytes", err)
				if err := json.Unmarshal(gotBodyBytes, &gotBody); err != nil {
					t.Fatalf("failed to unmarshal request body: %v", err)
				}

				wantBodyBytes, _ := json.Marshal(tt.req.Message)
				json.Unmarshal(wantBodyBytes, &wantBody)

				if diff := gcmp.Diff(gotBody, wantBody); diff != "" {
					t.Errorf("RequestWithContext() Body mismatch (-want +got):\n%s", diff)
				}
			}

			if strings.Contains(got.Header.Get("Content-Type"), "multipart/form-data") {
				err = got.ParseMultipartForm(10 << 20) // 10 MB
				if err != nil {
					t.Fatalf("failed to parse multipart form: %v", err)
				}

				for key, expectedVal := range tt.req.Form.Fields {
					if gotVal := got.FormValue(key); gotVal != expectedVal {
						t.Errorf("expected form field %s = %v, got %v", key, expectedVal, gotVal)
					}
				}

				file, _, err := got.FormFile(tt.req.Form.FormFile.Name)
				if err != nil {
					t.Fatalf("failed to get form file: %v", err)
				}
				defer file.Close()

				fileContent, _ := io.ReadAll(file)
				expectedFileContent, _ := os.ReadFile(tt.req.Form.FormFile.Path)
				if string(fileContent) != string(expectedFileContent) {
					t.Errorf("expected file content = %v, got %v", string(expectedFileContent), string(fileContent))
				}
			}
		})
	}
}
func TestRequest_URL(t *testing.T) {
	t.Parallel()
	testBaseURL := "https://api.example.com"

	type testCase[T any] struct {
		name    string
		method  string
		baseURL string
		want    string
		wantErr bool
		options []whttp.RequestOption[T]
	}
	tests := []testCase[any]{
		{
			name:    "base only",
			method:  http.MethodGet,
			baseURL: testBaseURL,
			want:    "https://api.example.com/",
		},
		{
			name:    "with endpoints",
			method:  http.MethodGet,
			baseURL: "https://api.example.com",
			want:    "https://api.example.com/v1/users",
			options: []whttp.RequestOption[any]{
				whttp.WithRequestEndpoints[any]("v1", "users"),
			},
		},
		{
			name:    "with query params",
			method:  http.MethodGet,
			baseURL: "https://api.example.com",
			want:    "https://api.example.com/v1/users?age=30&name=John",
			options: []whttp.RequestOption[any]{
				whttp.WithRequestEndpoints[any]("v1", "users"),
				whttp.WithRequestQueryParams[any](map[string]string{
					"name": "John",
					"age":  "30",
				}),
			},
		},
		{
			name:    "with query params and headers",
			method:  http.MethodGet,
			baseURL: "https://api.example.com",
			want:    "https://api.example.com/v1/users?age=30&debug=warning&name=John",
			options: []whttp.RequestOption[any]{
				whttp.WithRequestEndpoints[any]("v1", "users"),
				whttp.WithRequestQueryParams[any](map[string]string{
					"name": "John",
					"age":  "30",
				}),
				whttp.WithRequestDebugLogLevel[any](whttp.DebugLogLevelWarning),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			request := whttp.MakeRequest(tt.method, tt.baseURL, tt.options...)
			got, err := request.URL()
			if (err != nil) != tt.wantErr {
				t.Errorf("URL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("URL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMakeRequest(t *testing.T) {
	t.Parallel()

	t.Run("basic request creation", func(t *testing.T) {
		t.Parallel()

		req := whttp.MakeRequest(
			http.MethodPost,
			"https://api.example.com",
			whttp.WithRequestType[TestMessage](whttp.RequestTypeSendMessage),
			whttp.WithRequestBearer[TestMessage]("test-token"),
			whttp.WithRequestEndpoints[TestMessage]("v1", "messages"),
		)

		if req.Method != http.MethodPost {
			t.Errorf("expected Method=%s, got %s", http.MethodPost, req.Method)
		}
		if req.BaseURL != "https://api.example.com" {
			t.Errorf("expected BaseURL=https://api.example.com, got %s", req.BaseURL)
		}
		if req.Bearer != "test-token" {
			t.Errorf("expected Bearer=test-token, got %s", req.Bearer)
		}
		if len(req.Endpoints) != 2 || req.Endpoints[0] != "v1" || req.Endpoints[1] != "messages" {
			t.Errorf("expected Endpoints=[v1, messages], got %v", req.Endpoints)
		}
	})

	t.Run("request with message", func(t *testing.T) {
		t.Parallel()

		msg := &TestMessage{Name: "test", Value: 42}
		req := whttp.MakeRequest(
			http.MethodPost,
			"https://api.example.com",
			whttp.WithRequestMessage[TestMessage](msg),
		)

		if req.Message != msg {
			t.Errorf("expected Message to be set")
		}
	})

	t.Run("request with headers and query params", func(t *testing.T) {
		t.Parallel()

		headers := map[string]string{"X-Custom": "value"}
		queryParams := map[string]string{"key": "value"}

		req := whttp.MakeRequest(
			http.MethodGet,
			"https://api.example.com",
			whttp.WithRequestHeaders[TestMessage](headers),
			whttp.WithRequestQueryParams[TestMessage](queryParams),
		)

		if req.Headers["X-Custom"] != "value" {
			t.Errorf("expected header X-Custom=value")
		}
		if req.QueryParams["key"] != "value" {
			t.Errorf("expected query param key=value")
		}
	})
}
func TestMakeDownloadRequest(t *testing.T) {
	t.Parallel()

	downloadURL := "https://cdn.example.com/file.pdf"
	req := whttp.MakeDownloadRequest[any](
		downloadURL,
		whttp.WithRequestBearer[any]("test-token"),
		whttp.WithRequestHeaders[any](map[string]string{"X-Download": "true"}),
	)

	if req.Method != http.MethodGet {
		t.Errorf("expected Method=%s, got %s", http.MethodGet, req.Method)
	}
	if req.DownloadURL != downloadURL {
		t.Errorf("expected DownloadURL=%s, got %s", downloadURL, req.DownloadURL)
	}
	if req.Headers["X-Download"] != "true" {
		t.Errorf("expected header X-Download=true")
	}
}
func TestNewRequestWithContext(t *testing.T) {
	t.Parallel()

	t.Run("creates request with context", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		msg := &TestMessage{Name: "test", Value: 99}

		req, err := whttp.NewRequestWithContext[TestMessage](
			ctx,
			http.MethodPost,
			"https://api.example.com",
			whttp.WithRequestMessage[TestMessage](msg),
		)

		test.AssertNoError(t, "NewRequestWithContext", err)

		if req.Method != http.MethodPost {
			t.Errorf("expected Method=%s, got %s", http.MethodPost, req.Method)
		}
		if req.URL.String() != "https://api.example.com/" {
			t.Errorf("expected URL=https://api.example.com/, got %s", req.URL.String())
		}
	})
}
func TestWithRequestOptions(t *testing.T) {
	t.Parallel()

	t.Run("WithRequestMetadata", func(t *testing.T) {
		t.Parallel()

		metadata := types.Metadata{"key": "value"}
		req := whttp.MakeRequest(
			http.MethodPost,
			"https://api.example.com",
			whttp.WithRequestMetadata[TestMessage](metadata),
		)

		if req.Metadata == nil || req.Metadata["key"] != "value" {
			t.Error("expected metadata to be set")
		}
	})

	t.Run("WithRequestForm", func(t *testing.T) {
		t.Parallel()

		form := &whttp.RequestForm{
			Fields: map[string]string{"field": "value"},
		}
		req := whttp.MakeRequest(
			http.MethodPost,
			"https://api.example.com",
			whttp.WithRequestForm[TestMessage](form),
		)

		if req.Form == nil || req.Form.Fields["field"] != "value" {
			t.Error("expected form to be set")
		}
	})

	t.Run("WithRequestAppSecret", func(t *testing.T) {
		t.Parallel()

		req := whttp.MakeRequest(
			http.MethodPost,
			"https://api.example.com",
			whttp.WithRequestAppSecret[TestMessage]("secret123"),
		)

		if req.AppSecret != "secret123" {
			t.Errorf("expected AppSecret=secret123, got %s", req.AppSecret)
		}
	})

	t.Run("WithRequestSecured", func(t *testing.T) {
		t.Parallel()

		req := whttp.MakeRequest(
			http.MethodPost,
			"https://api.example.com",
			whttp.WithRequestSecured[TestMessage](true),
		)

		if !req.SecureRequests {
			t.Error("expected SecureRequests to be true")
		}
	})

	t.Run("WithRequestBodyReader", func(t *testing.T) {
		t.Parallel()

		reader := strings.NewReader("test body")
		req := whttp.MakeRequest(
			http.MethodPost,
			"https://api.example.com",
			whttp.WithRequestBodyReader[TestMessage](reader),
		)

		if req.BodyReader == nil {
			t.Error("expected BodyReader to be set")
		}
	})
}
func TestRequest_SetEndpoints(t *testing.T) {
	t.Parallel()

	req := &whttp.Request[TestMessage]{
		Method:  http.MethodGet,
		BaseURL: "https://api.example.com",
	}

	req.SetEndpoints("v1", "messages", "123")

	if len(req.Endpoints) != 3 {
		t.Errorf("expected 3 endpoints, got %d", len(req.Endpoints))
	}
	if req.Endpoints[0] != "v1" || req.Endpoints[1] != "messages" || req.Endpoints[2] != "123" {
		t.Errorf("unexpected endpoints: %v", req.Endpoints)
	}
}
func TestRequest_SetRequestMessage(t *testing.T) {
	t.Parallel()

	req := &whttp.Request[TestMessage]{
		Method:  http.MethodPost,
		BaseURL: "https://api.example.com",
	}

	msg := &TestMessage{Name: "test", Value: 999}
	req.SetRequestMessage(msg)

	if req.Message != msg {
		t.Error("expected message to be set")
	}
	if req.Message.Name != "test" || req.Message.Value != 999 {
		t.Errorf("expected Name=test and Value=999, got Name=%s and Value=%d", req.Message.Name, req.Message.Value)
	}
}
func TestRequest_SetDebugLogLevel(t *testing.T) {
	t.Parallel()

	req := &whttp.Request[TestMessage]{
		Method:  http.MethodGet,
		BaseURL: "https://api.example.com",
	}

	req.SetDebugLogLevel(whttp.DebugLogLevelAll)

	url, err := req.URL()
	test.AssertNoError(t, "URL", err)

	if !strings.Contains(url, "debug=all") {
		t.Errorf("expected URL to contain debug=all, got %s", url)
	}
}
func TestRequestWithContext_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("nil request", func(t *testing.T) {
		t.Parallel()

		_, err := whttp.RequestWithContext[TestMessage](context.Background(), nil)
		test.AssertError(t, "should error on nil request", err)
	})

	t.Run("request without app secret verification", func(t *testing.T) {
		t.Parallel()

		req := &whttp.Request[TestMessage]{
			Method:  http.MethodPost,
			BaseURL: "https://api.example.com",
			Message: &TestMessage{Name: "test", Value: 1},
		}

		httpReq, err := whttp.RequestWithContext(context.Background(), req)
		test.AssertNoError(t, "RequestWithContext", err)

		if httpReq == nil {
			t.Fatal("expected non-nil http request")
		}
	})
}
func TestRequest_URL_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("download URL takes precedence", func(t *testing.T) {
		t.Parallel()

		req := &whttp.Request[TestMessage]{
			DownloadURL: "https://download.example.com/file.pdf",
			BaseURL:     "https://api.example.com",
			Endpoints:   []string{"should", "be", "ignored"},
		}

		url, err := req.URL()
		test.AssertNoError(t, "URL", err)

		if url != "https://download.example.com/file.pdf" {
			t.Errorf("expected download URL, got %s", url)
		}
	})

	t.Run("URL with multiple query params", func(t *testing.T) {
		t.Parallel()

		req := &whttp.Request[TestMessage]{
			BaseURL:   "https://api.example.com",
			Endpoints: []string{"v1", "users"},
			QueryParams: map[string]string{
				"filter": "active",
				"sort":   "name",
				"limit":  "10",
			},
		}

		url, err := req.URL()
		test.AssertNoError(t, "URL", err)

		if !strings.Contains(url, "filter=active") || !strings.Contains(url, "sort=name") {
			t.Errorf("expected URL to contain all query params, got %s", url)
		}
	})

	t.Run("URL with secure requests but missing bearer token", func(t *testing.T) {
		t.Parallel()

		req := &whttp.Request[TestMessage]{
			BaseURL:        "https://api.example.com",
			SecureRequests: true,
			AppSecret:      "secret",
			Bearer:         "",
		}

		_, err := req.URL()
		test.AssertError(t, "should error without bearer token", err)
	})
}
func TestRequestWithContext_MoreEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("request with invalid base URL", func(t *testing.T) {
		t.Parallel()

		req := &whttp.Request[TestMessage]{
			Method:  http.MethodGet,
			BaseURL: "://invalid-url",
		}

		_, err := whttp.RequestWithContext(context.Background(), req)
		test.AssertError(t, "should error on invalid URL", err)
	})

	t.Run("form encoding error propagation", func(t *testing.T) {
		t.Parallel()

		req := &whttp.Request[TestMessage]{
			Method:  http.MethodPost,
			BaseURL: "https://api.example.com",
			Form: &whttp.RequestForm{
				Fields: map[string]string{"field": "value"},
				FormFile: &whttp.FormFile{
					Name: "file",
					Path: "/nonexistent/path/file.txt",
				},
			},
		}

		_, err := whttp.RequestWithContext(context.Background(), req)
		test.AssertError(t, "should error on form encoding failure", err)
	})
}
