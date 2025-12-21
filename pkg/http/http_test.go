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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gcmp "github.com/google/go-cmp/cmp"

	"github.com/piusalfred/whatsapp/internal/test"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func ExampleNewSender() {
	customHTTPClient := &http.Client{
		Timeout: 2 * time.Second,
	}

	// Define middleware that logs request information
	loggingMiddleware := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
		return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
			fmt.Println("request logger init called before core middlewares")

			return next(ctx, req, decoder)
		}
	}

	methodPrinter := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
		return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
			fmt.Printf("from core middleware request method is: %s\n", req.Method)

			return next(ctx, req, decoder)
		}
	}

	loggingMiddleware2 := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
		return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
			fmt.Println("called after core middleware request logger final")

			return next(ctx, req, decoder)
		}
	}

	loggingMiddleware3 := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
		return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
			err := next(ctx, req, decoder)
			fmt.Println("called after request send execution and the err is:", err)

			return err
		}
	}

	reqInterceptor := func(_ context.Context, req *http.Request) error {
		fmt.Println("Just intercepted the request and the method is:", req.Method)

		return nil
	}

	resInterceptor := func(_ context.Context, res *http.Response) error {
		fmt.Println("Just intercepted the response and status code:", res.StatusCode)

		return nil
	}

	resBodyInterceptor := func(_ context.Context, res *http.Response) error {
		body, _ := io.ReadAll(res.Body)
		defer func() {
			res.Body = io.NopCloser(bytes.NewReader(body))
		}()
		fmt.Println("from response body interceptor:", string(body))

		return nil
	}

	options := []whttp.CoreClientOption[TestMessage]{
		whttp.WithCoreClientHTTPClient[TestMessage](customHTTPClient),
		whttp.WithCoreClientMiddlewares[TestMessage](methodPrinter),
		whttp.WithCoreClientRequestInterceptor[TestMessage](nil),
		whttp.WithCoreClientResponseInterceptor[TestMessage](resBodyInterceptor),
	}

	sender := whttp.NewSender(
		options...,
	)

	longLiveClient := &http.Client{Timeout: time.Hour}

	sender.PrependMiddlewares(loggingMiddleware) // they will be executed first before previous set middlewares if any
	sender.AppendMiddlewares(loggingMiddleware2, loggingMiddleware3)
	sender.SetHTTPClient(longLiveClient)
	sender.SetRequestInterceptor(reqInterceptor)  // overrides the default interceptors
	sender.SetResponseInterceptor(resInterceptor) // overrides the default interceptors

	echoHandler := func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "could not read body", http.StatusInternalServerError)

			return
		}
		defer r.Body.Close()

		w.WriteHeader(http.StatusOK)
		w.Write(bodyBytes)
	}

	server := httptest.NewServer(http.HandlerFunc(echoHandler))
	defer server.Close()

	req := &whttp.Request[TestMessage]{
		Method:    http.MethodPost,
		BaseURL:   server.URL,
		Endpoints: []string{"send"},
		Message:   &TestMessage{Name: "Hello", Value: 123},
		Bearer:    "example-token",
	}

	// Response decoder to handle the response
	decoder := whttp.ResponseDecoderJSON(&TestMessage{}, whttp.DecodeOptions{})

	// Send the request using the sender
	err := sender.Send(context.Background(), req, decoder)
	if err != nil {
		fmt.Println("Error sending request:", err)
	}

	// Output:
	// request logger init called before core middlewares
	// from core middleware request method is: POST
	// called after core middleware request logger final
	// Just intercepted the request and the method is: POST
	// Just intercepted the response and status code: 200
	// called after request send execution and the err is: <nil>
}
func TestCoreClient_SetMethods(t *testing.T) {
	t.Parallel()

	t.Run("SetHTTPClient", func(t *testing.T) {
		t.Parallel()

		client := whttp.NewSender[TestMessage]()
		customHTTP := &http.Client{Timeout: 5 * time.Second}

		client.SetHTTPClient(customHTTP)
	})

	t.Run("SetRequestInterceptor", func(t *testing.T) {
		t.Parallel()

		client := whttp.NewSender[TestMessage]()
		interceptor := whttp.RequestInterceptorFunc(func(ctx context.Context, req *http.Request) error {
			return nil
		})

		client.SetRequestInterceptor(interceptor)
	})

	t.Run("SetResponseInterceptor", func(t *testing.T) {
		t.Parallel()

		client := whttp.NewSender[TestMessage]()
		interceptor := whttp.ResponseInterceptorFunc(func(ctx context.Context, resp *http.Response) error {
			return nil
		})

		client.SetResponseInterceptor(interceptor)
	})

	t.Run("AppendMiddlewares", func(t *testing.T) {
		t.Parallel()

		client := whttp.NewSender[TestMessage]()
		middleware := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
			return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
				return next(ctx, req, decoder)
			}
		}

		client.AppendMiddlewares(middleware)
	})

	t.Run("PrependMiddlewares", func(t *testing.T) {
		t.Parallel()

		client := whttp.NewSender[TestMessage]()
		middleware := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
			return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
				return next(ctx, req, decoder)
			}
		}

		client.PrependMiddlewares(middleware)
	})
}
func TestSenderFunc_Send(t *testing.T) {
	t.Parallel()

	called := false
	sender := whttp.SenderFunc[TestMessage](func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
		called = true
		return nil
	})

	req := &whttp.Request[TestMessage]{
		Method:  http.MethodGet,
		BaseURL: "https://api.example.com",
	}

	decoder := whttp.ResponseDecoderFunc(func(ctx context.Context, resp *http.Response) error {
		return nil
	})

	err := sender.Send(context.Background(), req, decoder)
	test.AssertNoError(t, "Send", err)

	if !called {
		t.Error("expected sender function to be called")
	}
}
func TestMiddlewareExecution(t *testing.T) {
	t.Parallel()

	t.Run("middleware execution order", func(t *testing.T) {
		t.Parallel()

		var order []int

		middleware1 := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
			return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
				order = append(order, 1)
				err := next(ctx, req, decoder)
				order = append(order, 4)
				return err
			}
		}

		middleware2 := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
			return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
				order = append(order, 2)
				err := next(ctx, req, decoder)
				order = append(order, 3)
				return err
			}
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"name":"test","value":1}`))
		}))
		defer server.Close()

		client := whttp.NewSender[TestMessage](
			whttp.WithCoreClientMiddlewares[TestMessage](middleware1, middleware2),
		)

		req := &whttp.Request[TestMessage]{
			Method:  http.MethodGet,
			BaseURL: server.URL,
		}

		msg := &TestMessage{}
		decoder := whttp.ResponseDecoderJSON(msg, whttp.DecodeOptions{})

		err := client.Send(context.Background(), req, decoder)
		test.AssertNoError(t, "Send", err)

		expectedOrder := []int{1, 2, 3, 4}
		if diff := gcmp.Diff(expectedOrder, order); diff != "" {
			t.Errorf("middleware execution order mismatch (-want +got):\n%s", diff)
		}
	})
}
func TestNewSenderIntegration(t *testing.T) {
	t.Parallel()

	t.Run("complete request lifecycle with interceptors", func(t *testing.T) {
		t.Parallel()

		requestIntercepted := false
		responseIntercepted := false

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Custom-Header") != "custom-value" {
				t.Error("expected custom header to be present")
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TestMessage{Name: "response", Value: 200})
		}))
		defer server.Close()

		reqInterceptor := whttp.RequestInterceptorFunc(func(ctx context.Context, req *http.Request) error {
			requestIntercepted = true
			req.Header.Set("X-Custom-Header", "custom-value")
			return nil
		})

		resInterceptor := whttp.ResponseInterceptorFunc(func(ctx context.Context, resp *http.Response) error {
			responseIntercepted = true
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			}
			return nil
		})

		sender := whttp.NewSender[TestMessage](
			whttp.WithCoreClientRequestInterceptor[TestMessage](reqInterceptor),
			whttp.WithCoreClientResponseInterceptor[TestMessage](resInterceptor),
		)

		req := &whttp.Request[TestMessage]{
			Method:    http.MethodPost,
			BaseURL:   server.URL,
			Endpoints: []string{"api", "messages"},
			Message:   &TestMessage{Name: "request", Value: 100},
			Bearer:    "test-token",
		}

		result := &TestMessage{}
		decoder := whttp.ResponseDecoderJSON(result, whttp.DecodeOptions{})

		err := sender.Send(context.Background(), req, decoder)
		test.AssertNoError(t, "Send", err)

		if !requestIntercepted {
			t.Error("expected request interceptor to be called")
		}
		if !responseIntercepted {
			t.Error("expected response interceptor to be called")
		}
		if result.Name != "response" || result.Value != 200 {
			t.Errorf("expected Name=response and Value=200, got Name=%s and Value=%d", result.Name, result.Value)
		}
	})

	t.Run("error handling in interceptors", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		reqInterceptor := whttp.RequestInterceptorFunc(func(ctx context.Context, req *http.Request) error {
			return errors.New("interceptor error")
		})

		sender := whttp.NewSender[TestMessage](
			whttp.WithCoreClientRequestInterceptor[TestMessage](reqInterceptor),
		)

		req := &whttp.Request[TestMessage]{
			Method:  http.MethodGet,
			BaseURL: server.URL,
		}

		decoder := whttp.ResponseDecoderJSON(&TestMessage{}, whttp.DecodeOptions{})

		err := sender.Send(context.Background(), req, decoder)
		test.AssertError(t, "should return interceptor error", err)
		if !strings.Contains(err.Error(), "interceptor error") {
			t.Errorf("expected error to contain 'interceptor error', got %v", err)
		}
	})
}
func TestCoreClientOptions(t *testing.T) {
	t.Parallel()

	t.Run("WithCoreClientHTTPClient", func(t *testing.T) {
		t.Parallel()

		customClient := &http.Client{Timeout: 10 * time.Second}
		sender := whttp.NewSender[TestMessage](
			whttp.WithCoreClientHTTPClient[TestMessage](customClient),
		)

		if sender == nil {
			t.Error("expected sender to be created")
		}
	})

	t.Run("all options combined", func(t *testing.T) {
		t.Parallel()

		customClient := &http.Client{Timeout: 5 * time.Second}
		reqInterceptor := whttp.RequestInterceptorFunc(func(ctx context.Context, req *http.Request) error {
			return nil
		})
		resInterceptor := whttp.ResponseInterceptorFunc(func(ctx context.Context, resp *http.Response) error {
			return nil
		})
		middleware := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
			return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
				return next(ctx, req, decoder)
			}
		}

		sender := whttp.NewSender[TestMessage](
			whttp.WithCoreClientHTTPClient[TestMessage](customClient),
			whttp.WithCoreClientRequestInterceptor[TestMessage](reqInterceptor),
			whttp.WithCoreClientResponseInterceptor[TestMessage](resInterceptor),
			whttp.WithCoreClientMiddlewares[TestMessage](middleware),
		)

		if sender == nil {
			t.Error("expected sender to be created with all options")
		}
	})
}
func TestAnySender(t *testing.T) {
	t.Parallel()

	t.Run("AnySenderFunc", func(t *testing.T) {
		t.Parallel()

		called := false
		sender := whttp.AnySenderFunc(func(ctx context.Context, req *whttp.Request[any], decoder whttp.ResponseDecoder) error {
			called = true
			return nil
		})

		req := &whttp.Request[any]{
			Method:  http.MethodGet,
			BaseURL: "https://api.example.com",
		}

		decoder := whttp.ResponseDecoderFunc(func(ctx context.Context, resp *http.Response) error {
			return nil
		})

		err := sender.Send(context.Background(), req, decoder)
		test.AssertNoError(t, "Send", err)

		if !called {
			t.Error("expected sender to be called")
		}
	})

	t.Run("NewAnySender", func(t *testing.T) {
		t.Parallel()

		sender := whttp.NewAnySender()
		if sender == nil {
			t.Error("expected non-nil sender")
		}
	})
}
func TestCoreClientWithBaseSender(t *testing.T) {
	t.Parallel()

	called := false
	baseSender := whttp.SenderFunc[TestMessage](func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
		called = true
		return nil
	})

	client := whttp.NewSender[TestMessage]()
	client.SetBaseSender(baseSender)

	req := &whttp.Request[TestMessage]{
		Method:  http.MethodGet,
		BaseURL: "https://api.example.com",
	}

	decoder := whttp.ResponseDecoderFunc(func(ctx context.Context, resp *http.Response) error {
		return nil
	})

	err := client.Send(context.Background(), req, decoder)
	test.AssertNoError(t, "Send", err)

	if !called {
		t.Error("expected base sender to be called")
	}
}
func TestNewAnySender_WithOptions(t *testing.T) {
	t.Parallel()

	t.Run("with nil option", func(t *testing.T) {
		t.Parallel()

		sender := whttp.NewAnySender(nil)
		if sender == nil {
			t.Error("expected non-nil sender even with nil option")
		}
	})

	t.Run("with custom http client", func(t *testing.T) {
		t.Parallel()

		customClient := &http.Client{Timeout: 30 * time.Second}
		sender := whttp.NewAnySender(
			whttp.WithCoreClientHTTPClient[any](customClient),
		)
		if sender == nil {
			t.Error("expected non-nil sender")
		}
	})
}
func TestSendFuncWithInterceptors_ErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("http client Do error", func(t *testing.T) {
		t.Parallel()

		client := &http.Client{
			Transport: &errorTransport{},
		}

		req := &whttp.Request[TestMessage]{
			Method:  http.MethodGet,
			BaseURL: "https://api.example.com",
		}

		decoder := whttp.ResponseDecoderJSON(&TestMessage{}, whttp.DecodeOptions{})

		sendFunc := whttp.SendFuncWithInterceptors[TestMessage](client, nil, nil)
		err := sendFunc(context.Background(), req, decoder)

		test.AssertError(t, "should return HTTP client error", err)
	})

	t.Run("response hook read error handling", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"name":"test","value":123}`))
		}))
		defer server.Close()

		responseHookCalled := false
		resHook := whttp.ResponseInterceptorFunc(func(ctx context.Context, resp *http.Response) error {
			responseHookCalled = true
			return nil
		})

		client := &http.Client{}
		req := &whttp.Request[TestMessage]{
			Method:  http.MethodGet,
			BaseURL: server.URL,
		}

		result := &TestMessage{}
		decoder := whttp.ResponseDecoderJSON(result, whttp.DecodeOptions{})

		sendFunc := whttp.SendFuncWithInterceptors[TestMessage](client, nil, resHook)
		err := sendFunc(context.Background(), req, decoder)

		test.AssertNoError(t, "Send", err)
		if !responseHookCalled {
			t.Error("expected response hook to be called")
		}
	})
}

type errorTransport struct{}

func (t *errorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("transport error")
}

func TestSendFuncWithInterceptors_RequestHookError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"test","value":123}`))
	}))
	defer server.Close()

	expectedErr := errors.New("request hook failed")
	reqHook := whttp.RequestInterceptorFunc(func(ctx context.Context, req *http.Request) error {
		return expectedErr
	})

	client := &http.Client{}
	req := &whttp.Request[TestMessage]{
		Method:  http.MethodGet,
		BaseURL: server.URL,
	}

	result := &TestMessage{}
	decoder := whttp.ResponseDecoderJSON(result, whttp.DecodeOptions{})

	sendFunc := whttp.SendFuncWithInterceptors[TestMessage](client, reqHook, nil)
	err := sendFunc(context.Background(), req, decoder)

	test.AssertError(t, "should return request hook error", err)
	test.AssertErrorIs(t, "should be the expected error", err, expectedErr)
}

func TestSendFuncWithInterceptors_ResponseHookError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"test","value":123}`))
	}))
	defer server.Close()

	expectedErr := errors.New("response hook failed")
	resHook := whttp.ResponseInterceptorFunc(func(ctx context.Context, resp *http.Response) error {
		return expectedErr
	})

	client := &http.Client{}
	req := &whttp.Request[TestMessage]{
		Method:  http.MethodGet,
		BaseURL: server.URL,
	}

	result := &TestMessage{}
	decoder := whttp.ResponseDecoderJSON(result, whttp.DecodeOptions{})

	sendFunc := whttp.SendFuncWithInterceptors[TestMessage](client, nil, resHook)
	err := sendFunc(context.Background(), req, decoder)

	test.AssertError(t, "should return response hook error", err)
	test.AssertErrorIs(t, "should be the expected error", err, expectedErr)
}

func TestSendFuncWithInterceptors_DecoderError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	client := &http.Client{}
	req := &whttp.Request[TestMessage]{
		Method:  http.MethodGet,
		BaseURL: server.URL,
	}

	result := &TestMessage{}
	decoder := whttp.ResponseDecoderJSON(result, whttp.DecodeOptions{DisallowUnknownFields: true})

	sendFunc := whttp.SendFuncWithInterceptors[TestMessage](client, nil, nil)
	err := sendFunc(context.Background(), req, decoder)

	test.AssertError(t, "should return decoder error", err)
}

func TestCoreClient_SendWithMiddlewareChain(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Middleware-1") != "applied" {
			http.Error(w, "middleware 1 not applied", http.StatusBadRequest)
			return
		}
		if r.Header.Get("X-Middleware-2") != "applied" {
			http.Error(w, "middleware 2 not applied", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"success","value":200}`))
	}))
	defer server.Close()

	middleware1 := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
		return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
			if req.Headers == nil {
				req.Headers = make(map[string]string)
			}
			req.Headers["X-Middleware-1"] = "applied"
			return next(ctx, req, decoder)
		}
	}

	middleware2 := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
		return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
			if req.Headers == nil {
				req.Headers = make(map[string]string)
			}
			req.Headers["X-Middleware-2"] = "applied"
			return next(ctx, req, decoder)
		}
	}

	sender := whttp.NewSender(
		whttp.WithCoreClientMiddlewares[TestMessage](middleware1, middleware2),
	)

	req := &whttp.Request[TestMessage]{
		Method:  http.MethodGet,
		BaseURL: server.URL,
	}

	result := &TestMessage{}
	decoder := whttp.ResponseDecoderJSON(result, whttp.DecodeOptions{})

	err := sender.Send(context.Background(), req, decoder)
	test.AssertNoError(t, "Send", err)

	if result.Name != "success" || result.Value != 200 {
		t.Errorf("expected Name=success and Value=200, got Name=%s and Value=%d", result.Name, result.Value)
	}
}

func TestCoreClient_MiddlewareErrorShortCircuit(t *testing.T) {
	t.Parallel()

	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	expectedErr := errors.New("middleware error")
	errorMiddleware := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
		return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
			return expectedErr
		}
	}

	sender := whttp.NewSender(
		whttp.WithCoreClientMiddlewares[TestMessage](errorMiddleware),
	)

	req := &whttp.Request[TestMessage]{
		Method:  http.MethodGet,
		BaseURL: server.URL,
	}

	decoder := whttp.ResponseDecoderJSON(&TestMessage{}, whttp.DecodeOptions{})

	err := sender.Send(context.Background(), req, decoder)
	test.AssertError(t, "should return middleware error", err)
	test.AssertErrorIs(t, "should be the expected error", err, expectedErr)

	if serverCalled {
		t.Error("server should not have been called when middleware returns error")
	}
}

func TestCoreClient_InterceptorsWithBody(t *testing.T) {
	t.Parallel()

	responseBody := `{"name":"response","value":999}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	}))
	defer server.Close()

	var interceptedReqBody string
	var interceptedResBody string

	reqHook := whttp.RequestInterceptorFunc(func(ctx context.Context, req *http.Request) error {
		if req.Body != nil {
			body, _ := io.ReadAll(req.Body)
			interceptedReqBody = string(body)
			req.Body = io.NopCloser(bytes.NewReader(body))
		}
		return nil
	})

	resHook := whttp.ResponseInterceptorFunc(func(ctx context.Context, resp *http.Response) error {
		body, _ := io.ReadAll(resp.Body)
		interceptedResBody = string(body)
		resp.Body = io.NopCloser(bytes.NewReader(body))
		return nil
	})

	sender := whttp.NewSender(
		whttp.WithCoreClientRequestInterceptor[TestMessage](reqHook),
		whttp.WithCoreClientResponseInterceptor[TestMessage](resHook),
	)

	req := &whttp.Request[TestMessage]{
		Method:  http.MethodPost,
		BaseURL: server.URL,
		Message: &TestMessage{Name: "request", Value: 42},
	}

	result := &TestMessage{}
	decoder := whttp.ResponseDecoderJSON(result, whttp.DecodeOptions{})

	err := sender.Send(context.Background(), req, decoder)
	test.AssertNoError(t, "Send", err)

	if interceptedReqBody == "" {
		t.Error("request body should have been intercepted")
	}

	if interceptedResBody != responseBody {
		t.Errorf("expected intercepted response body %q, got %q", responseBody, interceptedResBody)
	}

	if result.Name != "response" || result.Value != 999 {
		t.Errorf("expected Name=response and Value=999, got Name=%s and Value=%d", result.Name, result.Value)
	}
}

func TestAnySenderFunc_Send(t *testing.T) {
	t.Parallel()

	called := false
	fn := whttp.AnySenderFunc(func(ctx context.Context, req *whttp.Request[any], decoder whttp.ResponseDecoder) error {
		called = true
		return nil
	})

	decoder := whttp.ResponseDecoderFunc(func(ctx context.Context, resp *http.Response) error {
		return nil
	})

	err := fn.Send(context.Background(), &whttp.Request[any]{}, decoder)
	test.AssertNoError(t, "Send", err)

	if !called {
		t.Error("expected AnySenderFunc to be called")
	}
}

func TestCoreClient_NilMiddlewaresSkipped(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"test","value":1}`))
	}))
	defer server.Close()

	executionOrder := []string{}

	mw1 := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
		return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
			executionOrder = append(executionOrder, "mw1")
			return next(ctx, req, decoder)
		}
	}

	mw3 := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
		return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
			executionOrder = append(executionOrder, "mw3")
			return next(ctx, req, decoder)
		}
	}

	sender := whttp.NewSender(
		whttp.WithCoreClientMiddlewares[TestMessage](mw1, nil, mw3),
	)

	req := &whttp.Request[TestMessage]{
		Method:  http.MethodGet,
		BaseURL: server.URL,
	}

	result := &TestMessage{}
	decoder := whttp.ResponseDecoderJSON(result, whttp.DecodeOptions{})

	err := sender.Send(context.Background(), req, decoder)
	test.AssertNoError(t, "Send", err)

	if len(executionOrder) != 2 || executionOrder[0] != "mw1" || executionOrder[1] != "mw3" {
		t.Errorf("expected execution order [mw1, mw3], got %v", executionOrder)
	}
}

func TestSendFuncWithInterceptors_InvalidRequest(t *testing.T) {
	t.Parallel()

	client := &http.Client{}

	sendFunc := whttp.SendFuncWithInterceptors[TestMessage](client, nil, nil)
	err := sendFunc(context.Background(), nil, nil)

	test.AssertError(t, "should return error for nil request", err)
}
