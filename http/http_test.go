/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type context struct {
	Method     string
	StatusCode int
	Headers    map[string]string
	Body       interface{}
}

func testServer(t *testing.T, ctx *context) *httptest.Server {
	t.Helper()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != ctx.Method {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		for key, value := range ctx.Headers {
			w.Header().Add(key, value)
		}

		w.WriteHeader(ctx.StatusCode)
		if ctx.Body != nil {
			body, err := json.Marshal(ctx.Body)
			if err != nil {
				t.Errorf("failed to marshal response body: %v", err)
				return
			}
			if _, err := w.Write(body); err != nil {
				t.Errorf("failed to write response body: %v", err)
			}
		}
	})

	return httptest.NewServer(handler)
}

func TestSend(t *testing.T) {
	t.Parallel()
	type args struct {
		method string
		url    string
		status int
		header map[string]string
		body   interface{}
	}

	tests := []struct {
		name    string
		args    args
		result  interface{}
		wantErr bool
	}{
		{
			name: "test send request",
			args: args{
				method: http.MethodGet,
				url:    "https://graph.facebook.com/v16.0/224225226/verify_code",
			},
			result: &Response{
				StatusCode: http.StatusOK,
				Body:       nil,
			},
		},
	}
}

func TestCreateRequestURL(t *testing.T) {
	t.Parallel()
	type args struct {
		baseURL    string
		apiVersion string
		senderID   string
		endpoints  []string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test create phone number verification request url",
			args: args{
				baseURL:    BaseURL,
				senderID:   "224225226",
				apiVersion: "v16.0",
				endpoints:  []string{"verify_code"},
			},
			want:    "https://graph.facebook.com/v16.0/224225226/verify_code",
			wantErr: false,
		},
		{
			name: "test create media delete request url",
			args: args{
				baseURL:    BaseURL,
				senderID:   "224225226", // this should be meda id
				apiVersion: "v16.0",
				endpoints:  nil,
			},
			want:    "https://graph.facebook.com/v16.0/224225226",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := CreateRequestURL(tt.args.baseURL, tt.args.apiVersion, tt.args.senderID, tt.args.endpoints...)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRequestURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateRequestURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJoinUrlParts(t *testing.T) {
	t.Parallel()
	type args struct {
		parts *RequestContext
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test join url parts",
			args: args{
				parts: &RequestContext{
					BaseURL:    BaseURL,
					SenderID:   "224225226",
					ApiVersion: "v16.0",
					Endpoints:  []string{"verify_code"},
				},
			},
			want:    "https://graph.facebook.com/v16.0/224225226/verify_code",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := requestURLFromContext(tt.args.parts)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestURLFromContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("requestURLFromContext() got = %v, want %v", got, tt.want)
			}
		})
	}
}
