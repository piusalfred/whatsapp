package router

import (
	"net/http"

	"github.com/piusalfred/whatsapp/webhooks"
)

type MiddlewareFunc func(http.Handler) http.Handler

type SimpleMux interface {
	http.Handler
	Handle(pattern string, handler http.Handler)
}

type Endpoints struct {
	Webhook                  string
	SubscriptionVerification string
}

// SimpleRouter wires a webhooks.Listener into an HTTP mux with configurable endpoints and middleware support.
//
// The SimpleMux interface is intentionally thin, requiring only Handle() and ServeHTTP() methods.
// This means most popular HTTP routers (like chi, gorilla/mux, etc.) satisfy this interface
// without any adaptation, allowing you to use your preferred router seamlessly.
//
// Middlewares are applied in the following order:
//  1. Route-specific middlewares (webhookMiddlewares or subscriptionVerificationMiddlewares)
//  2. Global middlewares (applied to all routes including custom ones added via Handle())
//
// All defaults can be customized using the provided SimpleRouterOption functions.
type SimpleRouter struct {
	listener                            *webhooks.Listener
	mux                                 SimpleMux
	endpoints                           Endpoints
	globalMiddlewares                   []MiddlewareFunc
	webhookMiddlewares                  []MiddlewareFunc
	subscriptionVerificationMiddlewares []MiddlewareFunc
}

type SimpleRouterOption interface {
	apply(*SimpleRouter)
}

type SimpleRouterOptionFunc func(*SimpleRouter)

func (f SimpleRouterOptionFunc) apply(r *SimpleRouter) {
	f(r)
}

func WithSimpleRouterMux(m SimpleMux) SimpleRouterOption {
	return SimpleRouterOptionFunc(func(r *SimpleRouter) {
		r.mux = m
	})
}

func WithSimpleRouterGlobalMiddlewares(mw ...MiddlewareFunc) SimpleRouterOption {
	return SimpleRouterOptionFunc(func(r *SimpleRouter) {
		r.globalMiddlewares = append(r.globalMiddlewares, mw...)
	})
}

func WithSimpleRouterWebhookMiddlewares(mw ...MiddlewareFunc) SimpleRouterOption {
	return SimpleRouterOptionFunc(func(r *SimpleRouter) {
		r.webhookMiddlewares = append(r.webhookMiddlewares, mw...)
	})
}

func WithSimpleRouterSubscriptionVerificationMiddlewares(mw ...MiddlewareFunc) SimpleRouterOption {
	return SimpleRouterOptionFunc(func(r *SimpleRouter) {
		r.subscriptionVerificationMiddlewares = append(r.subscriptionVerificationMiddlewares, mw...)
	})
}

func WithSimpleRouterEndpoints(e Endpoints) SimpleRouterOption {
	return SimpleRouterOptionFunc(func(r *SimpleRouter) { r.endpoints = e })
}

// NewSimpleRouter creates a new SimpleRouter with the given listener and options.
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
//	router, err := NewSimpleRouter(listener)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// With custom endpoints
//	router, err := NewSimpleRouter(listener,
//		WithSimpleRouterEndpoints(Endpoints{
//			Webhook:                  "/whatsapp/webhook",
//			SubscriptionVerification: "/whatsapp/verify",
//		}),
//	)
//
//	// With middlewares
//	router, err := NewSimpleRouter(listener,
//		WithSimpleRouterGlobalMiddlewares(loggingMiddleware, authMiddleware),
//		WithSimpleRouterWebhookMiddlewares(rateLimitMiddleware),
//	)
//
//	// With custom mux (e.g., chi router)
//	chiRouter := chi.NewRouter()
//	router, err := NewSimpleRouter(listener,
//		WithSimpleRouterMux(chiRouter),
//		WithSimpleRouterGlobalMiddlewares(middleware.Logger),
//	)
func NewSimpleRouter(listener *webhooks.Listener, opts ...SimpleRouterOption) (*SimpleRouter, error) {
	r := &SimpleRouter{
		listener: listener,
		mux:      http.NewServeMux(),
		endpoints: Endpoints{
			Webhook:                  "/webhooks",
			SubscriptionVerification: "/webhooks",
		},
	}

	for _, opt := range opts {
		opt.apply(r)
	}

	postWebhook := r.applyMiddlewares(
		http.HandlerFunc(listener.HandleNotification),
		r.webhookMiddlewares,
	)
	r.mux.Handle("POST "+r.endpoints.Webhook, postWebhook)

	getVerification := r.applyMiddlewares(
		http.HandlerFunc(listener.HandleSubscriptionVerification),
		r.subscriptionVerificationMiddlewares,
	)
	r.mux.Handle("GET "+r.endpoints.SubscriptionVerification, getVerification)

	return r, nil
}

func (r *SimpleRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *SimpleRouter) Handle(pattern string, h http.Handler) {
	r.mux.Handle(pattern, r.applyMiddlewares(h, nil))
}

func (r *SimpleRouter) applyMiddlewares(
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
