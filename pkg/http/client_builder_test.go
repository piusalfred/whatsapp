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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/piusalfred/whatsapp/internal/test"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func TestCoreClientBuilder_BuildSender(t *testing.T) {
	t.Parallel()

	t.Run("build with defaults", func(t *testing.T) {
		t.Parallel()

		b := whttp.NewCoreClientBuilder()
		client := whttp.BuildSender[TestMessage](b)

		if client == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("build with custom http client", func(t *testing.T) {
		t.Parallel()

		customClient := &http.Client{Timeout: 5 * time.Second}
		b := whttp.NewCoreClientBuilder().WithHTTPClient(customClient)
		client := whttp.BuildSender[TestMessage](b)

		if client == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("build with interceptors", func(t *testing.T) {
		t.Parallel()

		reqCalled := false
		resCalled := false

		b := whttp.NewCoreClientBuilder().
			WithRequestInterceptor(func(ctx context.Context, req *http.Request) error {
				reqCalled = true
				return nil
			}).
			WithResponseInterceptor(func(ctx context.Context, resp *http.Response) error {
				resCalled = true
				return nil
			})

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"name":"ok","value":1}`))
		}))
		defer server.Close()

		client := whttp.BuildSender[TestMessage](b)

		req := &whttp.Request[TestMessage]{
			Method:  http.MethodGet,
			BaseURL: server.URL,
		}

		decoder := whttp.ResponseDecoderJSON(&TestMessage{}, whttp.DecodeOptions{})
		err := client.Send(context.Background(), req, decoder)
		test.AssertNoError(t, "Send", err)

		if !reqCalled {
			t.Error("expected request interceptor to be called")
		}
		if !resCalled {
			t.Error("expected response interceptor to be called")
		}
	})

	t.Run("build with middlewares", func(t *testing.T) {
		t.Parallel()

		mwCalled := false
		mw := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
			return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
				mwCalled = true
				return next(ctx, req, decoder)
			}
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"name":"ok","value":1}`))
		}))
		defer server.Close()

		b := whttp.NewCoreClientBuilder()
		client := whttp.BuildSender(b, mw)

		req := &whttp.Request[TestMessage]{
			Method:  http.MethodGet,
			BaseURL: server.URL,
		}

		decoder := whttp.ResponseDecoderJSON(&TestMessage{}, whttp.DecodeOptions{})
		err := client.Send(context.Background(), req, decoder)
		test.AssertNoError(t, "Send", err)

		if !mwCalled {
			t.Error("expected middleware to be called")
		}
	})

	t.Run("build with limits", func(t *testing.T) {
		t.Parallel()

		b := whttp.NewCoreClientBuilder().
			WithMaxBodyBytes(1024).
			WithMaxHeaderBytes(512)

		client := whttp.BuildSender[TestMessage](b)
		if client == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("build any client", func(t *testing.T) {
		t.Parallel()

		b := whttp.NewCoreClientBuilder()
		client := whttp.BuildAnySender(b)

		if client == nil {
			t.Fatal("expected non-nil any client")
		}
	})

	t.Run("nil http client is ignored", func(t *testing.T) {
		t.Parallel()

		b := whttp.NewCoreClientBuilder().WithHTTPClient(nil)
		client := whttp.BuildSender[TestMessage](b)

		if client == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("zero limits are ignored", func(t *testing.T) {
		t.Parallel()

		b := whttp.NewCoreClientBuilder().
			WithMaxBodyBytes(0).
			WithMaxHeaderBytes(0)

		client := whttp.BuildSender[TestMessage](b)
		if client == nil {
			t.Fatal("expected non-nil client")
		}
	})
}

func TestCoreClientBuilder_FluentChaining(t *testing.T) {
	t.Parallel()

	b := whttp.NewCoreClientBuilder().
		WithHTTPClient(&http.Client{Timeout: 10 * time.Second}).
		WithRequestInterceptor(func(ctx context.Context, req *http.Request) error { return nil }).
		WithResponseInterceptor(func(ctx context.Context, resp *http.Response) error { return nil }).
		WithMaxBodyBytes(5 << 20).
		WithMaxHeaderBytes(1 << 20)

	client := whttp.BuildSender[TestMessage](b)
	if client == nil {
		t.Fatal("expected non-nil client after fluent chaining")
	}
}
