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

package webhooks_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piusalfred/whatsapp/webhooks"
)

var (
	errSimulatedRead     = errors.New("simulated read failure")
	errWrongSecret       = errors.New("wrong secret")
	errInvalidSignature  = errors.New("invalid signature")
	errFirstVerifier     = errors.New("first validator")
	errConfigUnavailable = errors.New("config unavailable")
	errConfigDown        = errors.New("config down")
)

func newTestConfigReader(token, secret string, validate bool) webhooks.ConfigReader {
	return webhooks.ConfigReaderFunc(func(r *http.Request) (*webhooks.Config, error) {
		return &webhooks.Config{
			Token:           token,
			AppSecret:       secret,
			ValidatePayload: validate,
		}, nil
	})
}

func okHandler() webhooks.NotificationHandler {
	return webhooks.NotificationHandlerFunc(
		func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
			return &webhooks.Response{StatusCode: http.StatusOK}
		},
	)
}

func statusHandler(code int) webhooks.NotificationHandler {
	return webhooks.NotificationHandlerFunc(
		func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
			return &webhooks.Response{StatusCode: code}
		},
	)
}

func panicHandler(msg string) webhooks.NotificationHandler {
	return webhooks.NotificationHandlerFunc(
		func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
			panic(msg)
		},
	)
}

func validWebhookBody() *bytes.Buffer {
	return bytes.NewBufferString(`{
		"object": "whatsapp_business_account",
		"entry": [{
			"id": "123456789",
			"time": 1719000000,
			"changes": [{
				"field": "messages",
				"value": {
					"messaging_product": "whatsapp",
					"metadata": {
						"display_phone_number": "15551234567",
						"phone_number_id": "987654321"
					},
					"contacts": [{
						"profile": {"name": "Test User"},
						"wa_id": "1234567890"
					}],
					"messages": [{
						"from": "1234567890",
						"id": "wamid.test123",
						"timestamp": "1719000000",
						"type": "text",
						"text": {"body": "hello"}
					}]
				}
			}]
		}]
	}`)
}

func postRequest(body *bytes.Buffer) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/webhook", body)
	r.Header.Set("Content-Type", "application/json")
	return r
}

func TestListener_Middleware_Ordering(t *testing.T) {
	t.Parallel()

	var order []int

	mw1 := func(next webhooks.NotificationHandler) webhooks.NotificationHandler {
		return webhooks.NotificationHandlerFunc(
			func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
				order = append(order, 1)
				resp := next.HandleNotification(ctx, n)
				order = append(order, 1)
				return resp
			},
		)
	}
	mw2 := func(next webhooks.NotificationHandler) webhooks.NotificationHandler {
		return webhooks.NotificationHandlerFunc(
			func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
				order = append(order, 2)
				resp := next.HandleNotification(ctx, n)
				order = append(order, 2)
				return resp
			},
		)
	}

	listener := webhooks.NewListener(okHandler(), newTestConfigReader("t", "", false), mw1, mw2)
	w := httptest.NewRecorder()
	listener.HandleNotification(w, postRequest(validWebhookBody()))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	want := []int{1, 2, 2, 1}
	if fmt.Sprint(order) != fmt.Sprint(want) {
		t.Errorf("middleware order = %v, want %v", order, want)
	}
}

func TestListener_Middleware_ShortCircuit(t *testing.T) {
	t.Parallel()

	handlerCalled := false
	h := webhooks.NotificationHandlerFunc(
		func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
			handlerCalled = true
			return &webhooks.Response{StatusCode: http.StatusOK}
		},
	)

	shortCircuit := func(next webhooks.NotificationHandler) webhooks.NotificationHandler {
		return webhooks.NotificationHandlerFunc(
			func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
				return &webhooks.Response{StatusCode: http.StatusUnauthorized}
			},
		)
	}

	listener := webhooks.NewListener(h, newTestConfigReader("t", "", false), shortCircuit)
	w := httptest.NewRecorder()
	listener.HandleNotification(w, postRequest(validWebhookBody()))

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 from short-circuit middleware, got %d", w.Code)
	}
	if handlerCalled {
		t.Error("inner handler should not be called after middleware short-circuit")
	}
}

func TestListener_Middleware_ModifyResponse(t *testing.T) {
	t.Parallel()

	modify := func(next webhooks.NotificationHandler) webhooks.NotificationHandler {
		return webhooks.NotificationHandlerFunc(
			func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
				resp := next.HandleNotification(ctx, n)
				if resp.StatusCode == http.StatusOK {
					resp.StatusCode = http.StatusCreated
				}
				return resp
			},
		)
	}

	listener := webhooks.NewListener(okHandler(), newTestConfigReader("t", "", false), modify)
	w := httptest.NewRecorder()
	listener.HandleNotification(w, postRequest(validWebhookBody()))

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestListener_Middleware_ReceivesParsedNotification(t *testing.T) {
	t.Parallel()

	var captured *webhooks.Notification
	capture := func(next webhooks.NotificationHandler) webhooks.NotificationHandler {
		return webhooks.NotificationHandlerFunc(
			func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
				captured = n
				return next.HandleNotification(ctx, n)
			},
		)
	}

	listener := webhooks.NewListener(okHandler(), newTestConfigReader("t", "", false), capture)
	w := httptest.NewRecorder()
	listener.HandleNotification(w, postRequest(validWebhookBody()))

	if captured == nil {
		t.Fatal("middleware did not receive the notification")
	}
	if captured.Object != "whatsapp_business_account" {
		t.Errorf("Object = %q, want %q", captured.Object, "whatsapp_business_account")
	}
	if len(captured.Entry) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(captured.Entry))
	}
}

func TestListener_ParseErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want int
	}{
		{"invalid JSON", "not json", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			listener := webhooks.NewListener(okHandler(), newTestConfigReader("t", "", false))
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")
			listener.HandleNotification(w, r)

			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}

func TestListener_PayloadTooLarge(t *testing.T) {
	t.Parallel()

	prefix := []byte(`{"object":"whatsapp_business_account","entry":[],"padding":"`)
	suffix := []byte(`"}`)
	padSize := webhooks.MaxPayloadBytes - len(prefix) - len(suffix) + 1
	if padSize < 0 {
		padSize = webhooks.MaxPayloadBytes + 1
	}
	padding := make([]byte, padSize)
	for i := range padding {
		padding[i] = 'x'
	}
	body := make([]byte, 0, len(prefix)+len(padding)+len(suffix))
	body = append(body, prefix...)
	body = append(body, padding...)
	body = append(body, suffix...)

	listener := webhooks.NewListener(okHandler(), newTestConfigReader("t", "", false))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	listener.HandleNotification(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for payload too large, got %d", w.Code)
	}
}

func TestListener_ReadError(t *testing.T) {
	t.Parallel()

	errReader := &errorReader{err: errSimulatedRead}
	listener := webhooks.NewListener(okHandler(), newTestConfigReader("t", "", false))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/webhook", errReader)
	listener.HandleNotification(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for read error, got %d", w.Code)
	}
}

type errorReader struct{ err error }

func (e *errorReader) Read(p []byte) (int, error) { return 0, e.err }

func TestListener_SignatureVerifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		validate  bool
		secret    string
		validator webhooks.SignatureVerifier
		want      int
	}{
		{
			"valid signature", true, "my-secret",
			webhooks.SignatureVerifierFunc(func(h http.Header, p []byte, s string) error {
				if s != "my-secret" {
					return errWrongSecret
				}
				return nil
			}),
			http.StatusOK,
		},
		{
			"rejected signature", true, "my-secret",
			webhooks.SignatureVerifierFunc(func(h http.Header, p []byte, s string) error {
				return errInvalidSignature
			}),
			http.StatusBadRequest,
		},
		{
			"skipped when validate false", false, "",
			webhooks.SignatureVerifierFunc(func(h http.Header, p []byte, s string) error {
				return errInvalidSignature
			}),
			http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			listener := webhooks.NewListener(okHandler(), newTestConfigReader("t", tt.secret, tt.validate))
			listener.SetSignatureVerifier(tt.validator)
			w := httptest.NewRecorder()
			listener.HandleNotification(w, postRequest(validWebhookBody()))

			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}

func TestListener_SetSignatureVerifier_LastWins(t *testing.T) {
	t.Parallel()

	first := webhooks.SignatureVerifierFunc(func(h http.Header, p []byte, s string) error {
		return errFirstVerifier
	})
	second := webhooks.SignatureVerifierFunc(func(h http.Header, p []byte, s string) error {
		return nil
	})

	listener := webhooks.NewListener(okHandler(), newTestConfigReader("t", "s", true))
	listener.SetSignatureVerifier(first)
	listener.SetSignatureVerifier(second)

	w := httptest.NewRecorder()
	listener.HandleNotification(w, postRequest(validWebhookBody()))

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 from second validator, got %d", w.Code)
	}
}

func TestListener_HandlerResponseCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code int
	}{
		{"200 OK", http.StatusOK},
		{"202 Accepted", http.StatusAccepted},
		{"500 Internal Server Error", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			listener := webhooks.NewListener(statusHandler(tt.code), newTestConfigReader("t", "", false))
			w := httptest.NewRecorder()
			listener.HandleNotification(w, postRequest(validWebhookBody()))

			if w.Code != tt.code {
				t.Errorf("status = %d, want %d", w.Code, tt.code)
			}
		})
	}
}

func TestListener_PanicRecovery(t *testing.T) {
	t.Parallel()

	t.Run("handler panic returns 500", func(t *testing.T) {
		t.Parallel()
		listener := webhooks.NewListener(panicHandler("boom"), newTestConfigReader("t", "", false))
		w := httptest.NewRecorder()
		listener.HandleNotification(w, postRequest(validWebhookBody()))

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500 after handler panic, got %d", w.Code)
		}
	})

	t.Run("middleware panic returns 500", func(t *testing.T) {
		t.Parallel()
		panickingMW := func(next webhooks.NotificationHandler) webhooks.NotificationHandler {
			return webhooks.NotificationHandlerFunc(
				func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
					panic("middleware panic")
				},
			)
		}
		listener := webhooks.NewListener(okHandler(), newTestConfigReader("t", "", false), panickingMW)
		w := httptest.NewRecorder()
		listener.HandleNotification(w, postRequest(validWebhookBody()))

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500 after middleware panic, got %d", w.Code)
		}
	})

	t.Run("panic notifies onError with PanicError", func(t *testing.T) {
		t.Parallel()
		var gotErr error
		listener := webhooks.NewListener(panicHandler("boom"), newTestConfigReader("t", "", false))
		listener.OnError(func(ctx context.Context, r *http.Request, err error) {
			gotErr = err
		})

		w := httptest.NewRecorder()
		listener.HandleNotification(w, postRequest(validWebhookBody()))

		if gotErr == nil {
			t.Fatal("onError was not called for panic")
		}
		pe, ok := webhooks.IsPanicError(gotErr)
		if !ok {
			t.Fatalf("error is not PanicError, got type %T: %v", gotErr, gotErr)
		}
		if pe.Value != "boom" {
			t.Errorf("panic value = %v, want 'boom'", pe.Value)
		}
	})
}

func TestListener_ConfigErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler webhooks.NotificationHandler
		reader  webhooks.ConfigReader
		want    int
	}{
		{"nil handler", nil, newTestConfigReader("t", "", false), http.StatusInternalServerError},
		{"nil config reader", okHandler(), nil, http.StatusInternalServerError},
		{"config read error", okHandler(), webhooks.ConfigReaderFunc(
			func(r *http.Request) (*webhooks.Config, error) {
				return nil, errConfigUnavailable
			},
		), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			listener := webhooks.NewListener(tt.handler, tt.reader)
			w := httptest.NewRecorder()
			listener.HandleNotification(w, postRequest(validWebhookBody()))

			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}

func TestListener_OnError(t *testing.T) {
	t.Parallel()

	t.Run("called on parse error", func(t *testing.T) {
		t.Parallel()
		var gotErr error
		listener := webhooks.NewListener(okHandler(), newTestConfigReader("t", "", false))
		listener.OnError(func(ctx context.Context, r *http.Request, err error) {
			gotErr = err
		})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("bad json"))
		listener.HandleNotification(w, r)
		if gotErr == nil {
			t.Fatal("onError was not called for parse error")
		}
	})

	t.Run("called on config reader error", func(t *testing.T) {
		t.Parallel()
		var gotErr error
		errReader := webhooks.ConfigReaderFunc(func(r *http.Request) (*webhooks.Config, error) {
			return nil, errConfigDown
		})
		listener := webhooks.NewListener(okHandler(), errReader)
		listener.OnError(func(ctx context.Context, r *http.Request, err error) {
			gotErr = err
		})
		w := httptest.NewRecorder()
		listener.HandleNotification(w, postRequest(validWebhookBody()))
		if gotErr == nil {
			t.Fatal("onError was not called for config reader error")
		}
		if !errors.Is(gotErr, errConfigDown) {
			t.Errorf("gotErr = %v, want errConfigDown", gotErr)
		}
	})

	t.Run("called with predefined errors for nil fields", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name    string
			handler webhooks.NotificationHandler
			reader  webhooks.ConfigReader
			wantErr error
		}{
			{"nil handler", nil, newTestConfigReader("t", "", false), webhooks.ErrHandlerNotConfigured},
			{"nil config reader", okHandler(), nil, webhooks.ErrConfigReaderNotConfigured},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				var gotErr error
				listener := webhooks.NewListener(tt.handler, tt.reader)
				listener.OnError(func(ctx context.Context, r *http.Request, err error) {
					gotErr = err
				})
				w := httptest.NewRecorder()
				listener.HandleNotification(w, postRequest(validWebhookBody()))
				if gotErr == nil {
					t.Fatal("onError was not called")
				}
				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("gotErr = %v, want %v", gotErr, tt.wantErr)
				}
			})
		}
	})

	t.Run("not called on success", func(t *testing.T) {
		t.Parallel()
		called := false
		listener := webhooks.NewListener(okHandler(), newTestConfigReader("t", "", false))
		listener.OnError(func(ctx context.Context, r *http.Request, err error) {
			called = true
		})
		w := httptest.NewRecorder()
		listener.HandleNotification(w, postRequest(validWebhookBody()))
		if called {
			t.Error("onError should not be called on success")
		}
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("subsequent calls overwrite previous", func(t *testing.T) {
		t.Parallel()
		var firstCalled, secondCalled bool
		listener := webhooks.NewListener(nil, newTestConfigReader("t", "", false))
		listener.OnError(func(ctx context.Context, r *http.Request, err error) { firstCalled = true })
		listener.OnError(func(ctx context.Context, r *http.Request, err error) { secondCalled = true })
		w := httptest.NewRecorder()
		listener.HandleNotification(w, postRequest(validWebhookBody()))
		if firstCalled {
			t.Error("first onError should have been overwritten")
		}
		if !secondCalled {
			t.Error("second onError was not called")
		}
	})

	t.Run("nil clears callback", func(t *testing.T) {
		t.Parallel()
		called := false
		listener := webhooks.NewListener(nil, newTestConfigReader("t", "", false))
		listener.OnError(func(ctx context.Context, r *http.Request, err error) { called = true })
		listener.OnError(nil)
		w := httptest.NewRecorder()
		listener.HandleNotification(w, postRequest(validWebhookBody()))
		if called {
			t.Error("onError should not be called after clearing with nil")
		}
	})

	t.Run("receives context and request", func(t *testing.T) {
		t.Parallel()
		var gotReq *http.Request
		listener := webhooks.NewListener(nil, newTestConfigReader("t", "", false))
		listener.OnError(func(ctx context.Context, r *http.Request, err error) {
			gotReq = r
		})
		w := httptest.NewRecorder()
		r := postRequest(validWebhookBody())
		r.Header.Set("X-Custom", "test-value")
		listener.HandleNotification(w, r)
		if gotReq == nil {
			t.Fatal("onError request is nil")
		}
		if gotReq.Header.Get("X-Custom") != "test-value" {
			t.Error("onError request does not have expected header")
		}
	})
}

func TestListener_SubscriptionVerification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
		token string
		want  int
	}{
		{"valid token", "hub.mode=subscribe&hub.challenge=abc123&hub.verify_token=my-token", "my-token", http.StatusOK},
		{
			"wrong token",
			"hub.mode=subscribe&hub.challenge=abc123&hub.verify_token=wrong",
			"correct",
			http.StatusForbidden,
		},
		{"wrong mode", "hub.mode=invalid&hub.challenge=abc123&hub.verify_token=t", "t", http.StatusForbidden},
		{"missing token", "hub.mode=subscribe&hub.challenge=abc123", "t", http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			listener := webhooks.NewListener(okHandler(), newTestConfigReader(tt.token, "", false))
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/webhook?"+tt.query, nil)
			listener.HandleSubscriptionVerification(w, r)

			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
			if tt.want == http.StatusOK && w.Body.String() != "abc123" {
				t.Errorf("challenge = %q, want %q", w.Body.String(), "abc123")
			}
		})
	}

	t.Run("HTML-escapes challenge", func(t *testing.T) {
		t.Parallel()
		listener := webhooks.NewListener(okHandler(), newTestConfigReader("t", "", false))
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet,
			"/webhook?hub.mode=subscribe&hub.challenge=<script>&hub.verify_token=t", nil)
		listener.HandleSubscriptionVerification(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if strings.Contains(w.Body.String(), "<script>") {
			t.Error("challenge was not HTML-escaped")
		}
	})

	t.Run("nil config reader returns 500", func(t *testing.T) {
		t.Parallel()
		listener := webhooks.NewListener(okHandler(), nil)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/webhook?hub.mode=subscribe&hub.challenge=c&hub.verify_token=t", nil)
		listener.HandleSubscriptionVerification(w, r)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500 for nil config reader, got %d", w.Code)
		}
	})

	t.Run("config read error returns 500", func(t *testing.T) {
		t.Parallel()
		errReader := webhooks.ConfigReaderFunc(func(r *http.Request) (*webhooks.Config, error) {
			return nil, errConfigDown
		})
		listener := webhooks.NewListener(okHandler(), errReader)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/webhook?hub.mode=subscribe&hub.challenge=c&hub.verify_token=t", nil)
		listener.HandleSubscriptionVerification(w, r)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500 for config error, got %d", w.Code)
		}
	})

	t.Run("onError fires on verification failure", func(t *testing.T) {
		t.Parallel()
		var gotErr error
		listener := webhooks.NewListener(okHandler(), newTestConfigReader("t", "", false))
		listener.OnError(func(ctx context.Context, r *http.Request, err error) {
			gotErr = err
		})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(
			http.MethodGet,
			"/webhook?hub.mode=subscribe&hub.challenge=c&hub.verify_token=wrong",
			nil,
		)
		listener.HandleSubscriptionVerification(w, r)
		if gotErr == nil {
			t.Fatal("onError was not called for verification failure")
		}
	})
}

func TestListener_ResponseBodyNeverLeaksInternalError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler webhooks.NotificationHandler
		reader  webhooks.ConfigReader
		body    string
		want    int
	}{
		{"nil handler", nil, newTestConfigReader("t", "", false), `{}`, http.StatusInternalServerError},
		{"nil config reader", okHandler(), nil, `{}`, http.StatusInternalServerError},
		{"bad json", okHandler(), newTestConfigReader("t", "", false), `not json`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			listener := webhooks.NewListener(tt.handler, tt.reader)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(tt.body))
			listener.HandleNotification(w, r)

			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
			responseBody := strings.TrimSpace(w.Body.String())
			if strings.Contains(responseBody, "webhooks:") {
				t.Errorf("response body leaks internal error: %q", responseBody)
			}
		})
	}
}

func TestListener_ZeroValue(t *testing.T) {
	t.Parallel()

	ls := &webhooks.Listener{}
	w := httptest.NewRecorder()
	ls.HandleNotification(w, postRequest(validWebhookBody()))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 from zero-value Listener, got %d", w.Code)
	}
}
