package webhooks_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piusalfred/whatsapp/message"
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

func TestListener_HandleNotification_MultipleMessages(t *testing.T) {
	payload := `{
	  "object": "whatsapp_business_account",
	  "entry": [
	    {
	      "id": "234234234234",
	      "changes": [
	        {
	          "field": "messages",
	          "value": {
	            "messaging_product": "whatsapp",
	            "metadata": {
	              "display_phone_number": "1-555-123-4567",
	              "phone_number_id": "987654321"
	            },
	            "contacts": [
	              {
	                "profile": {
	                  "name": "Sam Sample"
	                },
	                "wa_id": "1234567890"
	              }
	            ],
	            "messages": [
	              {
	                "from": "1234567890",
	                "id": "wamid.GBGX2323902",
	                "timestamp": "1680123456",
	                "text": {
	                  "body": "Hello from a text message"
	                },
	                "type": "text"
	              },
	              {
	                "from": "1234567890",
	                "id": "wamid.OMGReaction",
	                "timestamp": "1680123457",
	                "reaction": {
	                  "message_id": "wamid.GBGX2323902",
	                  "emoji": "👍"
	                },
	                "type": "reaction"
	              }
	            ]
	          }
	        }
	      ]
	    }
	  ]
	}`

	var textHandled, reactionHandled bool

	handler := webhooks.NewHandler()
	handler.OnTextMessage(
		func(ctx context.Context, nctx *webhooks.MessageNotificationContext, mctx *webhooks.MessageInfo, txt *webhooks.Text) error {
			textHandled = true

			if txt.Body != "Hello from a text message" {
				t.Errorf("Text message body mismatch: got %s, want %s", txt.Body, "Hello from a text message")
			}

			return nil
		},
	)

	handler.OnReactionMessage(
		func(ctx context.Context, nctx *webhooks.MessageNotificationContext, mctx *webhooks.MessageInfo, reaction *message.Reaction) error {
			reactionHandled = true

			if reaction.Emoji != "👍" {
				t.Errorf("Reaction emoji mismatch: got %s, want %s", reaction.Emoji, "👍")
			}

			return nil
		},
	)

	cfg := TestServerConfig{
		Handler: handler,
		VerifyTokenReader: func(ctx context.Context) (string, error) {
			return "dummy-verify-token", nil
		},
		ValidateOptions:      &webhooks.ValidateOptions{Validate: false}, // no HMAC signature validation
		Middlewares:          nil,
		VerifyEndpoint:       "/verify",
		NotificationEndpoint: "/webhook",
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/webhook", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	if !textHandled || !reactionHandled {
		t.Error("Not all messages were handled")
	}
}

func TestListener_HandleNotification_MultipleChangeValues(t *testing.T) {
	payload := `{
	  "object": "whatsapp_business_account",
	  "entry": [
	    {
	      "id": "111111111111",
	      "changes": [
	        {
	          "field": "messages",
	          "value": {
	            "messaging_product": "whatsapp",
	            "metadata": {
	              "display_phone_number": "1-555-123-4567",
	              "phone_number_id": "987654321"
	            },
	            "contacts": [
	              {
	                "profile": {
	                  "name": "Alex Tester"
	                },
	                "wa_id": "1234567890"
	              }
	            ],
	            "messages": [
	              {
	                "from": "1234567890",
	                "id": "wamid.TEXTXYZ123",
	                "timestamp": "1680123456",
	                "text": {
	                  "body": "Hello from a text message"
	                },
	                "type": "text"
	              }
	            ]
	          }
	        },
	        {
	          "field": "user_preferences",
	          "value": {
	            "messaging_product": "whatsapp",
	            "metadata": {
	              "display_phone_number": "1-555-123-4567",
	              "phone_number_id": "987654321"
	            },
	            "contacts": [
	              {
	                "profile": {
	                  "name": "Alex Tester"
	                },
	                "wa_id": "1234567890"
	              }
	            ],
	            "user_preferences": [
	              {
	                "wa_id": "1234567890",
	                "detail": "User wants to stop marketing notifications",
	                "category": "marketing_messages",
	                "value": "stop",
	                "timestamp": "1690000000"
	              }
	            ]
	          }
	        }
	      ]
	    }
	  ]
	}`

	var (
		textHandled         bool
		userPreferencesSeen bool
	)

	// Create a new Handler
	handler := webhooks.NewHandler()

	handler.OnTextMessage(
		func(ctx context.Context, nctx *webhooks.MessageNotificationContext, mctx *webhooks.MessageInfo, txt *webhooks.Text) error {
			textHandled = true

			if txt.Body != "Hello from a text message" {
				t.Errorf("Text body mismatch, got=%s, want=%s", txt.Body, "Hello from a text message")
			}
			return nil
		},
	)

	// When we get user preferences update
	handler.OnUserPreferencesUpdate(
		func(ctx context.Context, nctx *webhooks.MessageNotificationContext, prefs []*webhooks.UserPreference) error {
			userPreferencesSeen = true

			if len(prefs) != 1 {
				t.Errorf("Expected 1 user preference, got %d", len(prefs))
				return nil
			}
			p := prefs[0]
			if p.Value != "stop" {
				t.Errorf("Preference mismatch, got=%s, want=stop", p.Value)
			}
			return nil
		},
	)

	// Build the test server
	cfg := TestServerConfig{
		Handler: handler,
		VerifyTokenReader: func(ctx context.Context) (string, error) {
			return "dummy-verify-token", nil
		},
		ValidateOptions:      &webhooks.ValidateOptions{Validate: false}, // Disabling signature validation
		Middlewares:          nil,
		VerifyEndpoint:       "/verify",
		NotificationEndpoint: "/webhook",
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	// Send the POST request
	resp, err := http.Post(ts.URL+"/webhook", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// Check that both handlers ran
	if !textHandled {
		t.Error("Text message was not handled")
	}
	if !userPreferencesSeen {
		t.Error("User preferences were not handled")
	}
}
