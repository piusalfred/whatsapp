package webhooks_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piusalfred/whatsapp/webhooks"
)

type TestServerConfig struct {
	Handler              webhooks.NotificationHandler
	VerifyTokenReader    webhooks.VerifyTokenReader
	ValidateOptions      *webhooks.ValidateOptions
	Middlewares          []webhooks.Middleware
	VerifyEndpoint       string
	NotificationEndpoint string
}

func NewTestWebhookServer(t *testing.T, cfg TestServerConfig) *httptest.Server {
	t.Helper()
	listener := webhooks.NewListener(
		cfg.Handler.HandleNotification,
		cfg.VerifyTokenReader,
		cfg.ValidateOptions,
		cfg.Middlewares...,
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case cfg.VerifyEndpoint:
			listener.HandleSubscriptionVerification(w, r)
		case cfg.NotificationEndpoint:
			listener.HandleNotification(w, r)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))

	return server
}
