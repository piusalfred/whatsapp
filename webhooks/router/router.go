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

package router

import (
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp/webhooks"
)

type MiddlewareFunc func(http.Handler) http.Handler

type Mux interface {
	http.Handler
	Handle(pattern string, handler http.Handler)
}

type Endpoints struct {
	Webhook                  string
	SubscriptionVerification string
}

// WebhookRouter wires a webhooks.Listener into an HTTP mux with configurable endpoints and middleware support.
//
// The Mux interface is intentionally thin, requiring only Handle() and ServeHTTP() methods.
// This means most popular HTTP routers (like chi, gorilla/mux, etc.) satisfy this interface
// without any adaptation, allowing you to use your preferred router seamlessly.
//
// Middlewares are applied in the following order:
//  1. Route-specific middlewares (webhookMiddlewares or subscriptionVerificationMiddlewares)
//  2. Global middlewares (applied to all routes including custom ones added via Handle())
//
// All defaults can be customized using the provided WebhookRouterOption functions.
type WebhookRouter struct {
	listener                            *webhooks.Listener
	mux                                 Mux
	endpoints                           Endpoints
	globalMiddlewares                   []MiddlewareFunc
	webhookMiddlewares                  []MiddlewareFunc
	subscriptionVerificationMiddlewares []MiddlewareFunc
}

type WebhookRouterOption interface {
	apply(*WebhookRouter)
}

type WebhookRouterOptionFunc func(*WebhookRouter)

func (f WebhookRouterOptionFunc) apply(r *WebhookRouter) {
	f(r)
}

func WithWebhookRouterMux(m Mux) WebhookRouterOption {
	return WebhookRouterOptionFunc(func(r *WebhookRouter) {
		r.mux = m
	})
}

func WithWebhookRouterGlobalMiddlewares(mw ...MiddlewareFunc) WebhookRouterOption {
	return WebhookRouterOptionFunc(func(r *WebhookRouter) {
		r.globalMiddlewares = append(r.globalMiddlewares, mw...)
	})
}

func WithWebhookRouterWebhookMiddlewares(mw ...MiddlewareFunc) WebhookRouterOption {
	return WebhookRouterOptionFunc(func(r *WebhookRouter) {
		r.webhookMiddlewares = append(r.webhookMiddlewares, mw...)
	})
}

func WithWebhookRouterSubscriptionVerificationMiddlewares(mw ...MiddlewareFunc) WebhookRouterOption {
	return WebhookRouterOptionFunc(func(r *WebhookRouter) {
		r.subscriptionVerificationMiddlewares = append(r.subscriptionVerificationMiddlewares, mw...)
	})
}

func WithWebhookRouterEndpoints(e Endpoints) WebhookRouterOption {
	return WebhookRouterOptionFunc(func(r *WebhookRouter) { r.endpoints = e })
}

const defaultEndpoint = "/webhooks"

// NewWebhookRouter creates a new WebhookRouter with the given listener and options.
//
// Defaults (if not modified by options):
//   - Mux: http.NewServeMux()
//   - Webhook endpoint: "/webhooks" (POST)
//   - Subscription verification endpoint: "/webhooks" (GET)
//   - No middlewares applied
//
// Example usage:
//
//	// Basic usage with defaults
//	router, err := NewWebhookRouter(listener)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// With custom endpoints
//	router, err := NewWebhookRouter(listener,
//		WithWebhookRouterEndpoints(Endpoints{
//			Webhook:                  "/whatsapp/webhook",
//			SubscriptionVerification: "/whatsapp/verify",
//		}),
//	)
//
//	// With middlewares
//	router, err := NewWebhookRouter(listener,
//		WithWebhookRouterGlobalMiddlewares(loggingMiddleware, authMiddleware),
//		WithWebhookRouterWebhookMiddlewares(rateLimitMiddleware),
//	)
//
//	// With custom mux (e.g., chi router)
//	chiRouter := chi.NewRouter()
//	router, err := NewWebhookRouter(listener,
//		WithWebhookRouterMux(chiRouter),
//		WithWebhookRouterGlobalMiddlewares(middleware.Logger),
//	)
func NewWebhookRouter(listener *webhooks.Listener, opts ...WebhookRouterOption) (*WebhookRouter, error) {
	r := &WebhookRouter{
		listener: listener,
		mux:      http.NewServeMux(),
		endpoints: Endpoints{
			Webhook:                  defaultEndpoint,
			SubscriptionVerification: defaultEndpoint,
		},
	}

	for _, opt := range opts {
		opt.apply(r)
	}

	var (
		postEndpoint = fmt.Sprintf("POST %s", r.endpoints.Webhook)
		getEndpoint  = fmt.Sprintf("GET %s", r.endpoints.SubscriptionVerification)
		postHandler  = r.applyMiddlewares(
			http.HandlerFunc(listener.HandleNotification),
			r.webhookMiddlewares,
		)
		getHandler = r.applyMiddlewares(
			http.HandlerFunc(listener.HandleSubscriptionVerification),
			r.subscriptionVerificationMiddlewares,
		)
	)

	r.mux.Handle(postEndpoint, postHandler)
	r.mux.Handle(getEndpoint, getHandler)

	return r, nil
}

func (r *WebhookRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *WebhookRouter) Handle(pattern string, h http.Handler) {
	r.mux.Handle(pattern, r.applyMiddlewares(h, nil))
}

func (r *WebhookRouter) applyMiddlewares(
	h http.Handler,
	branch []MiddlewareFunc,
) http.Handler {
	for i := len(branch) - 1; i >= 0; i-- {
		h = branch[i](h)
	}
	for i := len(r.globalMiddlewares) - 1; i >= 0; i-- {
		h = r.globalMiddlewares[i](h)
	}
	return h
}
