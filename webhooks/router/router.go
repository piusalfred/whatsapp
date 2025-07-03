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

// SimpleRouter wires a webhooks.Listener into an HTTP mux.
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
