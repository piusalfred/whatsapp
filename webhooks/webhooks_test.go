package webhooks_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/piusalfred/whatsapp/webhooks"
	"github.com/piusalfred/whatsapp/webhooks/message"
)

type TestServerConfig[T any] struct {
	Handler              webhooks.NotificationHandler[T]
	VerifyTokenReader    webhooks.VerifyTokenReader
	ValidateOptions      *webhooks.ValidateOptions
	Middlewares          []webhooks.HandleMiddleware[T]
	VerifyEndpoint       string
	NotificationEndpoint string
}

func NewTestWebhookServer[T any](t *testing.T, cfg TestServerConfig[T]) *httptest.Server {
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

func TestListener_HandleNotification_Message(t *testing.T) {
	t.Parallel()
	var (
		reader = webhooks.VerifyTokenReader(func(_ context.Context) (string, error) {
			return "APP-SECRET", nil
		})
		verifyEndpoint       = "/webhooks/verify"
		notificationEndpoint = "/webhooks/notify"
	)

	type testCase[T any] struct {
		name            string
		validateOptions *webhooks.ValidateOptions
		handler         func() webhooks.NotificationHandler[T]
		middlewares     []webhooks.HandleMiddleware[T]
		payload         []byte
		challenge       string
	}

	printNotificationMiddleware := webhooks.HandleMiddleware[message.Notification](
		func(handlerFunc webhooks.NotificationHandlerFunc[message.Notification]) webhooks.NotificationHandlerFunc[message.Notification] {
			return func(ctx context.Context, notification *message.Notification) *webhooks.Response {
				fmt.Println("Before middleware")
				resp := handlerFunc.HandleNotification(ctx, notification)
				fmt.Println("After middleware")

				return resp
			}
		},
	)

	tests := []testCase[message.Notification]{
		{
			name: "normal text message webhook",
			validateOptions: &webhooks.ValidateOptions{
				Validate:  false,
				AppSecret: "NADA",
			},
			handler: func() webhooks.NotificationHandler[message.Notification] {
				handler := &message.Handlers{
					TextMessage: message.HandlerFunc[message.Text](
						func(_ context.Context, _ *message.NotificationContext,
							_ *message.Info, message *message.Text,
						) error {
							fmt.Printf("Message: %s\n", message.Body)

							return nil
						}),
				}

				return webhooks.NotificationHandlerFunc[message.Notification](handler.HandleNotification)
			},
			middlewares: []webhooks.HandleMiddleware[message.Notification]{printNotificationMiddleware},
			payload: []byte(
				`{"object": "whatsapp_business_account", "entry": [{"id": "<WHATSAPP_BUSINESS_ACCOUNT_ID>", "changes": [{"value": {"messaging_product": "whatsapp", "contacts": [{"profile": {"name": "<WHATSAPP_USER_NAME>"}, "wa_id": "<WHATSAPP_USER_ID>"}], "messages": [{"from": "<WHATSAPP_USER_PHONE_NUMBER>", "id": "<WHATSAPP_MESSAGE_ID>", "timestamp": "<WEBHOOK_SENT_TIMESTAMP>", "text": {"body": "dudeeeee"}, "type": "text"}]}, "field": "messages"}]}]}`,
			),
			challenge: "SAYCHEESEIFYOUKNOWWHATIMEAN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := NewTestWebhookServer[message.Notification](t, TestServerConfig[message.Notification]{
				Handler:              tt.handler(),
				VerifyEndpoint:       verifyEndpoint,
				VerifyTokenReader:    reader,
				NotificationEndpoint: notificationEndpoint,
				Middlewares:          tt.middlewares,
				ValidateOptions:      tt.validateOptions,
			})

			defer server.Close()

			params := url.Values{}
			params.Add("hub.mode", "subscribe")
			params.Add("hub.challenge", tt.challenge)
			params.Add("hub.verify_token", "APP-SECRET")

			ctx := t.Context()
			req, err := http.NewRequestWithContext(
				ctx, http.MethodGet, fmt.Sprintf("%s?%s", server.URL+verifyEndpoint, params.Encode()), nil)
			if err != nil {
				t.Fatalf("error creating http request for subscription verification")
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status OK for verification, got %v", resp.StatusCode)
			}

			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if string(body) != tt.challenge {
				t.Errorf("Expected challenge %s, got %s", tt.challenge, string(body))
			}

			// Test notification
			req, err = http.NewRequestWithContext(ctx,
				http.MethodPost, server.URL+notificationEndpoint, bytes.NewBuffer(tt.payload))
			if err != nil {
				t.Fatalf("error creating http request for notification")
			}

			req.Header.Set("Content-Type", "application/json")

			resp, err = http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status OK, got %v", resp.StatusCode)
			}
		})
	}
}
