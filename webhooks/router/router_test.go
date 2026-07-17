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

package router_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/piusalfred/whatsapp/message/media"
	"github.com/piusalfred/whatsapp/webhooks"
	"github.com/piusalfred/whatsapp/webhooks/router"
)

// ---------------------------------------------------------------------------
// test helpers
// ---------------------------------------------------------------------------

func newTestListener() *webhooks.Listener {
	handler := webhooks.NotificationHandlerFunc(
		func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
			return &webhooks.Response{StatusCode: http.StatusOK}
		},
	)
	return webhooks.NewListener(
		handler,
		webhooks.ConfigReaderFunc(func(r *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{Token: "test-token", ValidatePayload: false}, nil
		}),
	)
}

// middlewareRecorder appends its name to calls on every request, so tests can
// verify middleware execution order.
func middlewareRecorder(name string, calls *[]string) router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*calls = append(*calls, name)
			next.ServeHTTP(w, r)
		})
	}
}

// spyMux delegates to http.ServeMux while recording every Handle call so
// tests can assert which patterns were registered.
type spyMux struct {
	*http.ServeMux
	mu          sync.Mutex
	handleCalls []handleCall
}

type handleCall struct {
	pattern string
}

func newSpyMux() *spyMux {
	return &spyMux{ServeMux: http.NewServeMux()}
}

func (s *spyMux) Handle(pattern string, handler http.Handler) {
	s.mu.Lock()
	s.handleCalls = append(s.handleCalls, handleCall{pattern: pattern})
	s.mu.Unlock()
	s.ServeMux.Handle(pattern, handler)
}

func (s *spyMux) handlePatterns() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.handleCalls))
	for i, c := range s.handleCalls {
		out[i] = c.pattern
	}
	return out
}

// ---------------------------------------------------------------------------
// endpoint registration
// ---------------------------------------------------------------------------

func TestEndpoints_Defaults(t *testing.T) {
	spy := newSpyMux()
	l := newTestListener()

	_, err := router.NewWebhookRouter(l, router.WithWebhookRouterMux(spy))
	if err != nil {
		t.Fatalf("NewWebhookRouter() error = %v", err)
	}

	want := []string{"POST /webhooks", "GET /webhooks"}
	if diff := cmp.Diff(want, spy.handlePatterns()); diff != "" {
		t.Errorf("registered patterns (-want +got):\n%s", diff)
	}
}

func TestEndpoints_Custom(t *testing.T) {
	spy := newSpyMux()
	l := newTestListener()

	_, err := router.NewWebhookRouter(l,
		router.WithWebhookRouterMux(spy),
		router.WithWebhookRouterEndpoints(router.Endpoints{
			Webhook:                  "/whatsapp/events",
			SubscriptionVerification: "/whatsapp/verify",
		}),
	)
	if err != nil {
		t.Fatalf("NewWebhookRouter() error = %v", err)
	}

	want := []string{"POST /whatsapp/events", "GET /whatsapp/verify"}
	if diff := cmp.Diff(want, spy.handlePatterns()); diff != "" {
		t.Errorf("registered patterns (-want +got):\n%s", diff)
	}
}

// ---------------------------------------------------------------------------
// middleware ordering
// ---------------------------------------------------------------------------

func TestMiddlewareOrdering_WebhookRoute(t *testing.T) {
	var calls []string
	l := newTestListener()

	r, err := router.NewWebhookRouter(l,
		router.WithWebhookRouterGlobalMiddlewares(
			middlewareRecorder("global-1", &calls),
			middlewareRecorder("global-2", &calls),
		),
		router.WithWebhookRouterWebhookMiddlewares(
			middlewareRecorder("webhook-1", &calls),
			middlewareRecorder("webhook-2", &calls),
		),
	)
	if err != nil {
		t.Fatalf("NewWebhookRouter() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/webhooks",
		strings.NewReader(`{"object":"whatsapp_business_account","entry":[]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// global outermost → runs first; branch innermost → runs last before handler.
	want := []string{"global-1", "global-2", "webhook-1", "webhook-2"}
	if diff := cmp.Diff(want, calls); diff != "" {
		t.Errorf("execution order (-want +got):\n%s", diff)
	}
}

func TestMiddlewareOrdering_VerificationRoute(t *testing.T) {
	var calls []string
	l := newTestListener()

	r, err := router.NewWebhookRouter(l,
		router.WithWebhookRouterGlobalMiddlewares(
			middlewareRecorder("global", &calls),
		),
		router.WithWebhookRouterSubscriptionVerificationMiddlewares(
			middlewareRecorder("verify", &calls),
		),
	)
	if err != nil {
		t.Fatalf("NewWebhookRouter() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet,
		"/webhooks?hub.mode=subscribe&hub.challenge=chal&hub.verify_token=test-token", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	want := []string{"global", "verify"}
	if diff := cmp.Diff(want, calls); diff != "" {
		t.Errorf("execution order (-want +got):\n%s", diff)
	}
}

// ---------------------------------------------------------------------------
// middleware isolation (no leak between routes)
// ---------------------------------------------------------------------------

func TestMiddlewareIsolation_WebhookDoesNotLeakToVerification(t *testing.T) {
	var calls []string
	l := newTestListener()

	r, err := router.NewWebhookRouter(l,
		router.WithWebhookRouterWebhookMiddlewares(
			middlewareRecorder("webhook-only", &calls),
		),
	)
	if err != nil {
		t.Fatalf("NewWebhookRouter() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet,
		"/webhooks?hub.mode=subscribe&hub.challenge=ch&hub.verify_token=test-token", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if len(calls) != 0 {
		t.Errorf("webhook middleware leaked to verification: %v", calls)
	}
}

func TestMiddlewareIsolation_VerificationDoesNotLeakToWebhook(t *testing.T) {
	var calls []string
	l := newTestListener()

	r, err := router.NewWebhookRouter(l,
		router.WithWebhookRouterSubscriptionVerificationMiddlewares(
			middlewareRecorder("verify-only", &calls),
		),
	)
	if err != nil {
		t.Fatalf("NewWebhookRouter() error = %v", err)
	}

	body := `{"object":"whatsapp_business_account","entry":[{"id":"1","changes":[{"field":"messages","value":{"messaging_product":"whatsapp","metadata":{"display_phone_number":"1","phone_number_id":"1"},"contacts":[{"profile":{"name":"test"},"wa_id":"1"}],"messages":[{"from":"1","id":"1","timestamp":"1","type":"text","text":{"body":"hi"}}]}}]}]}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if len(calls) != 0 {
		t.Errorf("verification middleware leaked to webhook: %v", calls)
	}
}

// ---------------------------------------------------------------------------
// Handle — custom routes get global middleware
// ---------------------------------------------------------------------------

func TestHandle_GlobalMiddlewareApplied(t *testing.T) {
	var calls []string
	l := newTestListener()

	r, err := router.NewWebhookRouter(l,
		router.WithWebhookRouterGlobalMiddlewares(
			middlewareRecorder("global", &calls),
		),
	)
	if err != nil {
		t.Fatalf("NewWebhookRouter() error = %v", err)
	}

	r.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	want := []string{"global"}
	if diff := cmp.Diff(want, calls); diff != "" {
		t.Errorf("custom route middleware (-want +got):\n%s", diff)
	}
}

func TestHandle_BranchMiddlewareNotApplied(t *testing.T) {
	// Custom routes added via Handle() should NOT receive branch middlewares
	// (webhook or verification) — only global.
	var calls []string
	l := newTestListener()

	r, err := router.NewWebhookRouter(l,
		router.WithWebhookRouterWebhookMiddlewares(
			middlewareRecorder("webhook", &calls),
		),
		router.WithWebhookRouterSubscriptionVerificationMiddlewares(
			middlewareRecorder("verify", &calls),
		),
	)
	if err != nil {
		t.Fatalf("NewWebhookRouter() error = %v", err)
	}

	r.Handle("GET /ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if len(calls) != 0 {
		t.Errorf("branch middleware leaked to custom route: %v", calls)
	}
}

// ---------------------------------------------------------------------------
// middleware accumulation (multiple With... calls)
// ---------------------------------------------------------------------------

func TestMiddlewareAccumulation_Global(t *testing.T) {
	var calls []string
	l := newTestListener()

	r, err := router.NewWebhookRouter(l,
		router.WithWebhookRouterGlobalMiddlewares(
			middlewareRecorder("first", &calls),
		),
		router.WithWebhookRouterGlobalMiddlewares(
			middlewareRecorder("second", &calls),
		),
	)
	if err != nil {
		t.Fatalf("NewWebhookRouter() error = %v", err)
	}

	r.Handle("GET /test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// first With... call's middlewares end up outermost
	want := []string{"first", "second"}
	if diff := cmp.Diff(want, calls); diff != "" {
		t.Errorf("accumulation order (-want +got):\n%s", diff)
	}
}

func TestMiddlewareAccumulation_Branch(t *testing.T) {
	var calls []string
	l := newTestListener()

	r, err := router.NewWebhookRouter(l,
		router.WithWebhookRouterWebhookMiddlewares(
			middlewareRecorder("first", &calls),
		),
		router.WithWebhookRouterWebhookMiddlewares(
			middlewareRecorder("second", &calls),
		),
	)
	if err != nil {
		t.Fatalf("NewWebhookRouter() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/webhooks",
		strings.NewReader(`{"object":"whatsapp_business_account","entry":[]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	want := []string{"first", "second"}
	if diff := cmp.Diff(want, calls); diff != "" {
		t.Errorf("accumulation order (-want +got):\n%s", diff)
	}
}

// ---------------------------------------------------------------------------
// example: wiring into an HTTP server
// ---------------------------------------------------------------------------

// This example demonstrates mounting a WebhookRouter as a sub-handler under
// a /whatsapp path prefix on a parent mux that also serves other routes like
// /login and /logout.
func TestWebhookRouter_Example(t *testing.T) {
	t.Parallel()
	// 1. Create a typed webhooks sub-handler for WhatsApp events.
	handler := webhooks.NewHandler()
	handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
		func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
			// handle incoming text message via req.Payload.Body
			return nil
		},
	))
	handler.OnReactionMessage(webhooks.MessageHandlerFunc[media.Reaction](
		func(ctx context.Context, req *webhooks.MessageRequest[media.Reaction]) error {
			// handle incoming reaction via req.Payload.Emoji
			return nil
		},
	))

	reader := webhooks.ConfigReaderFunc(func(r *http.Request) (*webhooks.Config, error) {
		cfg := &webhooks.Config{Token: "my-verify-token", ValidatePayload: false}

		return cfg, nil
	})

	listener := webhooks.NewListener(
		handler,
		reader,
	)

	// 2. Build the WebhookRouter with endpoint paths relative to the mount
	//    point. When mounted at /whatsapp, the webhook URL is
	//    POST /whatsapp/webhooks and verification is GET /whatsapp/webhooks.
	r, err := router.NewWebhookRouter(listener,
		router.WithWebhookRouterEndpoints(router.Endpoints{
			Webhook:                  "/webhooks",
			SubscriptionVerification: "/webhooks",
		}),
	)
	if err != nil {
		panic(err)
	}

	// 3. Mount the WebhookRouter as a sub-handler on the main mux alongside
	//    other application routes. Use StripPrefix so the router sees its
	//    registered paths (/webhooks) instead of the full URL path
	//    (/whatsapp/webhooks).
	mux := http.NewServeMux()

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("/whatsapp/", http.StripPrefix("/whatsapp", r))
	mux.Handle("/whatsapp", r)

	// 4. Serve.
	//   srv := &http.Server{Addr: ":8080", Handler: mux}
	//   srv.ListenAndServe()
	_ = mux
}
