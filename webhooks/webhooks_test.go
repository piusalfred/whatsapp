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

package webhooks_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/piusalfred/whatsapp/message/media"
	werrors "github.com/piusalfred/whatsapp/pkg/errors"
	"github.com/piusalfred/whatsapp/webhooks"
)

var (
	errClientIDRequired      = errors.New("client ID is required")
	errClientIDNotRegistered = errors.New("client ID is not registered")
	errClientEnvRequired     = errors.New("client environment is required")
	errClientNotInEnv        = errors.New("client not registered in environment")
)

type (
	TestConfigMap map[string]*webhooks.Config
	TestEnvConfig struct {
		Dev  TestConfigMap
		Stg  TestConfigMap
		Prod TestConfigMap
	}
)

type TestMultiClientConfigReader struct {
	envConfig TestEnvConfig
	ids       map[string]string
}

func (r *TestMultiClientConfigReader) ReadConfig(request *http.Request) (*webhooks.Config, error) {
	// from the URL path we expect something like /webhooks/clients/256535634/whatsapp
	// so we can extract the client ID from the path
	clientID := strings.TrimPrefix(request.URL.Path, "/webhooks/clients/")
	clientID = strings.TrimSuffix(clientID, "/whatsapp")

	if clientID == "" {
		return nil, errClientIDRequired
	}

	clientName, ok := r.ids[clientID]
	if !ok {
		return nil, errClientIDNotRegistered
	}

	env := request.Header.Get("X-Client-Env")
	if env == "" {
		return nil, errClientEnvRequired
	}

	switch env {
	case "dev":
		cfg, ok := r.envConfig.Dev[clientName]
		if !ok {
			return nil, fmt.Errorf("client %s in dev: %w", clientName, errClientNotInEnv)
		}
		return cfg, nil
	case "stg":
		cfg, ok := r.envConfig.Stg[clientName]
		if !ok {
			return nil, fmt.Errorf("client %s in stg: %w", clientName, errClientNotInEnv)
		}
		return cfg, nil
	default:
		cfg, ok := r.envConfig.Prod[clientName]
		if !ok {
			return nil, fmt.Errorf("client %s in prod: %w", clientName, errClientNotInEnv)
		}
		return cfg, nil
	}
}

func ExampleConfigReader_ReadConfig() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	envConfig := TestEnvConfig{
		Dev: TestConfigMap{
			"acme": {
				Token:           "acme-dev-token",
				ValidatePayload: false,
				AppSecret:       "acme-dev-app-secret",
			},
			"shield": {
				Token:           "shield-dev-token",
				ValidatePayload: false,
				AppSecret:       "shield-dev-app-secret",
			},
		},
		Stg: TestConfigMap{
			"acme": {
				Token:           "acme-stg-token",
				ValidatePayload: false,
				AppSecret:       "acme-stg-app-secret",
			},
			"shield": {
				Token:           "shield-stg-token",
				ValidatePayload: false,
				AppSecret:       "shield-stg-app-secret",
			},
		},
		Prod: TestConfigMap{
			"acme": {
				Token:           "acme-prod-token",
				ValidatePayload: true,
				AppSecret:       "acme-prod-app-secret",
			},
			"shield": {
				Token:           "shield-prod-token",
				ValidatePayload: true,
				AppSecret:       "shield-prod-app-secret",
			},
		},
	}

	reader := &TestMultiClientConfigReader{
		envConfig: envConfig,
		ids: map[string]string{
			"256535634": "acme",
			"030877308": "shield",
		},
	}

	// First request - Acme in dev
	request := httptest.NewRequest(
		http.MethodPost,
		"https://api.localhost.com/webhooks/clients/256535634/whatsapp",
		nil,
	)
	request.Header.Set("X-Client-Env", "dev")

	cfg, err := reader.ReadConfig(request)
	if err != nil {
		logger.Error("Error reading config", "error", err)
		return
	}

	printFn := func(cfg *webhooks.Config) {
		fmt.Printf("Token: %s\n", cfg.Token)
		fmt.Printf("ValidatePayload: %t\n", cfg.ValidatePayload)
		fmt.Printf("AppSecret: %s\n", cfg.AppSecret)
	}

	printFn(cfg)

	// Second request - Shield in prod
	request = httptest.NewRequest(http.MethodPost, "https://api.localhost.com/webhooks/clients/030877308/whatsapp", nil)
	request.Header.Set("X-Client-Env", "prod")

	cfg, err = reader.ReadConfig(request)
	if err != nil {
		logger.Error("Error reading config", "error", err)
		return
	}
	printFn(cfg)

	// Output:
	// Token: acme-dev-token
	// ValidatePayload: false
	// AppSecret: acme-dev-app-secret
	// Token: shield-prod-token
	// ValidatePayload: true
	// AppSecret: shield-prod-app-secret
}

type TestServerConfig struct {
	Handler      webhooks.NotificationHandler
	Middlewares  []webhooks.Middleware
	ConfigReader webhooks.ConfigReader
}

func NewTestWebhookServer(t *testing.T, cfg TestServerConfig) *httptest.Server {
	t.Helper()
	listener := webhooks.NewListener(
		cfg.Handler,
		cfg.ConfigReader,
		cfg.Middlewares...,
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listener.HandleSubscriptionVerification(w, r)
		case http.MethodPost:
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
	handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
		func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
			txt := req.Payload
			textHandled = true

			if txt.Body != "Hello from a text message" {
				t.Errorf("Text message body mismatch: got %s, want %s", txt.Body, "Hello from a text message")
			}

			return nil
		},
	))

	handler.OnReactionMessage(webhooks.MessageHandlerFunc[media.Reaction](
		func(ctx context.Context, req *webhooks.MessageRequest[media.Reaction]) error {
			reaction := req.Payload
			reactionHandled = true

			if reaction.Emoji != "👍" {
				t.Errorf("Reaction emoji mismatch: got %s, want %s", reaction.Emoji, "👍")
			}

			return nil
		},
	))

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{
				Token:           "dummy-verify-token",
				ValidatePayload: false,
			}, nil
		}),
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

	handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
		func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
			txt := req.Payload
			textHandled = true

			if txt.Body != "Hello from a text message" {
				t.Errorf("Text body mismatch, got=%s, want=%s", txt.Body, "Hello from a text message")
			}
			return nil
		},
	))

	// When we get user preferences update
	handler.OnUserPreferencesUpdate(
		webhooks.UserPreferenceHandlerFunc(
			func(ctx context.Context, nctx *webhooks.MessageNotificationContext, p *webhooks.UserPreference) error {
				userPreferencesSeen = true
				if p.Value != "stop" {
					t.Errorf("Preference mismatch, got=%s, want=stop", p.Value)
				}
				return nil
			},
		),
	)

	// Build the test server
	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{
				Token:           "dummy-verify-token",
				ValidatePayload: false,
			}, nil
		}),
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	// send the POST request
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

func TestListener_HandleNotification_MultipleChangeValues1(t *testing.T) {
	payload := `{
  "object": "whatsapp_business_account",
  "entry": [
    {
      "id": "8888888888",
      "changes": [
        {
          "field": "messages",
          "value": {
            "messaging_product": "whatsapp",
            "metadata": {
              "display_phone_number": "14155238886",
              "phone_number_id": "1093849293"
            },
            "contacts": [
              {
                "profile": { "name": "John Doe" },
                "wa_id": "16505551234"
              }
            ],
            "messages": [
              {
                "from": "16505551234",
                "id": "wamid.TEXT123",
                "timestamp": "1680123456",
                "type": "text",
                "text": {
                  "body": "Hello from a text message!"
                }
              },
              {
                "from": "16505551234",
                "id": "wamid.LOC789",
                "timestamp": "1680123458",
                "location": {
                  "latitude": 37.7749,
                  "longitude": -122.4194,
                  "name": "San Francisco",
                  "address": "Market Street"
                }
              }
            ]
          }
        },
        {
          "field": "messages",
          "value": {
            "messaging_product": "whatsapp",
            "metadata": {
              "display_phone_number": "14155238886",
              "phone_number_id": "1093849293"
            },
            "contacts": [
              {
                "profile": { "name": "John Doe" },
                "wa_id": "16505551234"
              }
            ],
            "messages": [
              {
                "from": "16505551234",
                "id": "wamid.RXN001",
                "timestamp": "1680123460",
                "reaction": {
                  "message_id": "wamid.TEXT123",
                  "emoji": "👍"
                },
                "type": "reaction"
              },
              {
                "from": "16505551234",
                "id": "wamid.STK002",
                "timestamp": "1680123461",
                "type": "sticker",
                "sticker": {
                  "mime_type": "image/webp",
                  "sha256": "stickerHashABC",
                  "id": "STICKER1234"
                }
              }
            ]
          }
        },
        {
          "field": "user_preferences",
          "value": {
            "messaging_product": "whatsapp",
            "metadata": {
              "display_phone_number": "14155238886",
              "phone_number_id": "1093849293"
            },
            "contacts": [
              {
                "profile": { "name": "John Doe" },
                "wa_id": "16505551234"
              }
            ],
            "user_preferences": [
              {
                "wa_id": "16505551234",
                "detail": "User requested to stop marketing messages",
                "category": "marketing_messages",
                "value": "stop",
                "timestamp": "1680123462"
              },
              {
                "wa_id": "16505551234",
                "detail": "User also blocked location-sharing promotions",
                "category": "marketing_messages",
                "value": "stop",
                "timestamp": "1680123463"
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
		locationHandled     bool
		reactionHandled     bool
		stickerHandled      bool
		userPreferencesSeen bool
	)

	handler := webhooks.NewHandler()
	handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
		func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
			txt := req.Payload
			textHandled = true
			if txt.Body != "Hello from a text message!" {
				t.Errorf("Text body mismatch, got=%s, want=%s", txt.Body, "Hello from a text message!")
			}
			return nil
		},
	))
	handler.OnLocationMessage(webhooks.MessageHandlerFunc[media.Location](
		func(ctx context.Context, req *webhooks.MessageRequest[media.Location]) error {
			loc := req.Payload
			locationHandled = true
			if loc.Name != "San Francisco" {
				t.Errorf("Location name mismatch, got=%s, want=%s", loc.Name, "San Francisco")
			}
			return nil
		},
	))

	handler.OnReactionMessage(webhooks.MessageHandlerFunc[media.Reaction](
		func(ctx context.Context, req *webhooks.MessageRequest[media.Reaction]) error {
			reaction := req.Payload
			reactionHandled = true
			if reaction.Emoji != "👍" {
				t.Errorf("Reaction emoji mismatch, got=%s, want=%s", reaction.Emoji, "👍")
			}
			return nil
		},
	))

	handler.OnUserPreferencesUpdate(
		webhooks.UserPreferenceHandlerFunc(
			func(ctx context.Context, nctx *webhooks.MessageNotificationContext, p *webhooks.UserPreference) error {
				userPreferencesSeen = true
				if p.Value != "stop" {
					t.Errorf("Preference mismatch, got=%s, want=stop", p.Value)
				}
				return nil
			},
		),
	)

	handler.OnStickerMessage(webhooks.MessageHandlerFunc[media.Info](
		func(ctx context.Context, req *webhooks.MessageRequest[media.Info]) error {
			sticker := req.Payload
			stickerHandled = true
			if sticker.MimeType != "image/webp" {
				t.Errorf("Sticker mime type mismatch, got=%s, want=%s", sticker.MimeType, "image/webp")
			}

			return nil
		},
	))

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{
				Token:           "dummy-verify-token",
				ValidatePayload: false,
			}, nil
		}),
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

	if !textHandled || !locationHandled || !reactionHandled || !stickerHandled || !userPreferencesSeen {
		t.Error("Not all messages were handled")
	}
}

func TestListener_HandleNotification_ButtonMessage(t *testing.T) {
	payload := `{
      "object": "whatsapp_business_account",
      "entry": [{
          "id": "123456789",
          "changes": [{
              "value": {
                  "messaging_product": "whatsapp",
                  "metadata": {
                      "display_phone_number": "15551234567",
                      "phone_number_id": "987654321"
                  },
                  "contacts": [{
                      "profile": { "name": "Test Button" },
                      "wa_id": "100200300"
                  }],
                  "messages": [{
                      "context": {
                        "from": "15551234567",
                        "id": "wamid.IDCTX"
                      },
                      "from": "15551234567",
                      "id": "wamid.BUTTONMSG",
                      "timestamp": "1681000000",
                      "type": "button",
                      "button": {
                        "text": "No",
                        "payload": "No-Button-Payload"
                      }
                  }]
              },
              "field": "messages"
          }]
      }]
    }`

	var buttonHandled bool

	handler := webhooks.NewHandler()
	handler.OnButtonMessage(webhooks.MessageHandlerFunc[webhooks.Button](func(ctx context.Context,
		req *webhooks.MessageRequest[webhooks.Button],
	) error {
		btn := req.Payload
		buttonHandled = true

		if btn.Text != "No" {
			t.Errorf("Expected button text='No', got=%s", btn.Text)
		}
		if btn.Payload != "No-Button-Payload" {
			t.Errorf("Expected payload='No-Button-Payload', got=%s", btn.Payload)
		}
		return nil
	}))

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{
				Token: "test-token",
			}, nil
		}),
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

	if !buttonHandled {
		t.Errorf("Button message was not handled")
	}
}

func TestListener_HandleNotification_ListReply(t *testing.T) {
	payload := `{
	  "object": "whatsapp_business_account",
	  "entry": [
	    {
	      "id": "444444",
	      "changes": [
	        {
	          "field": "messages",
	          "value": {
	            "messaging_product": "whatsapp",
	            "metadata": {
	              "display_phone_number": "15551230000",
	              "phone_number_id": "123987456"
	            },
	            "contacts": [
	              {
	                "profile": { "name": "List Tester" },
	                "wa_id": "1987654321"
	              }
	            ],
	            "messages": [
	              {
	                "from": "1987654321",
	                "id": "wamid.IDLIST",
	                "timestamp": "1682000010",
	                "type": "interactive",
	                "interactive": {
	                  "list_reply": {
	                    "id": "list_reply_id",
	                    "title": "list_reply_title",
	                    "description": "list_reply_description"
	                  },
	                  "type": "list_reply"
	                }
	              }
	            ]
	          }
	        }
	      ]
	    }
	  ]
	}`

	var listReplyHandled bool

	handler := webhooks.NewHandler()
	handler.OnListReplyMessage(webhooks.MessageHandlerFunc[webhooks.ListReply](func(ctx context.Context,
		req *webhooks.MessageRequest[webhooks.ListReply],
	) error {
		lr := req.Payload
		listReplyHandled = true
		if lr.ID != "list_reply_id" {
			t.Errorf("ListReply ID mismatch, got=%s", lr.ID)
		}
		if lr.Title != "list_reply_title" {
			t.Errorf("ListReply Title mismatch, got=%s", lr.Title)
		}
		return nil
	}))

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{}, nil
		}),
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/webhook", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	if !listReplyHandled {
		t.Error("List reply was not handled")
	}
}

func TestListener_HandleNotification_ButtonReply(t *testing.T) {
	payload := `{
	  "object": "whatsapp_business_account",
	  "entry": [
	    {
	      "id": "555555",
	      "changes": [
	        {
	          "field": "messages",
	          "value": {
	            "messaging_product": "whatsapp",
	            "metadata": {
	              "display_phone_number": "15550001111",
	              "phone_number_id": "22222222"
	            },
	            "contacts": [
	              {
	                "profile": { "name": "ButtonReplyUser" },
	                "wa_id": "17771234567"
	              }
	            ],
	            "messages": [
	              {
	                "from": "17771234567",
	                "id": "wamid.BTNREPLY1",
	                "timestamp": "1683000022",
	                "type": "interactive",
	                "interactive": {
	                  "button_reply": {
	                    "id": "unique-button-identifier-here",
	                    "title": "button-text"
	                  },
	                  "type": "button_reply"
	                }
	              }
	            ]
	          }
	        }
	      ]
	    }
	  ]
	}`

	var buttonReplyHandled bool

	handler := webhooks.NewHandler()
	handler.OnButtonReplyMessage(webhooks.MessageHandlerFunc[webhooks.ButtonReply](func(ctx context.Context,
		req *webhooks.MessageRequest[webhooks.ButtonReply],
	) error {
		btn := req.Payload
		buttonReplyHandled = true
		if btn.ID != "unique-button-identifier-here" {
			t.Errorf("ButtonReply ID mismatch, got=%s", btn.ID)
		}
		if btn.Title != "button-text" {
			t.Errorf("ButtonReply Title mismatch, got=%s", btn.Title)
		}
		return nil
	}))

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{
				Token: "dummy",
			}, nil
		}),
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/webhook", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	if !buttonReplyHandled {
		t.Error("ButtonReply was not handled")
	}
}

func TestListener_HandleNotification_ReferralMessage(t *testing.T) {
	payload := `{
	  "object": "whatsapp_business_account",
	  "entry": [
	    {
	      "id": "99999",
	      "changes": [
	        {
	          "field": "messages",
	          "value": {
	            "messaging_product": "whatsapp",
	            "metadata": {
	              "display_phone_number": "14155238886",
	              "phone_number_id": "123123123"
	            },
	            "contacts": [
	              {
	                "profile": { "name": "AdUser" },
	                "wa_id": "100100200"
	              }
	            ],
	            "messages": [
	              {
	                "referral": {
	                  "source_url": "https://facebook.com/ad/123",
	                  "source_id": "ADID123",
	                  "source_type": "ad",
	                  "headline": "Ad Title",
	                  "body": "Ad Description",
	                  "media_type": "image",
	                  "image_url": "https://example.com/ad.jpg",
	                  "ctwa_clid": "CTWA_ABC"
	                },
	                "from": "100100200",
	                "id": "wamid.REF001",
	                "timestamp": "1684000033",
	                "type": "text",
	                "text": { "body": "Hi from an ad click!" }
	              }
	            ]
	          }
	        }
	      ]
	    }
	  ]
	}`

	var referralHandled bool

	handler := webhooks.NewHandler()
	handler.OnReferralMessage(webhooks.MessageHandlerFunc[webhooks.ReferralNotification](func(ctx context.Context,
		req *webhooks.MessageRequest[webhooks.ReferralNotification],
	) error {
		ref := req.Payload
		referralHandled = true

		if ref.Text.Body != "Hi from an ad click!" {
			t.Errorf("Referral text mismatch, got=%s", ref.Text.Body)
		}
		if ref.Referral.SourceID != "ADID123" {
			t.Errorf("Referral sourceID mismatch, got=%s", ref.Referral.SourceID)
		}
		return nil
	}))

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{}, nil
		}),
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/webhook", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if !referralHandled {
		t.Error("Referral message was not handled")
	}
}

func TestListener_HandleNotification_ProductInquiry(t *testing.T) {
	payload := `{
	  "object": "whatsapp_business_account",
	  "entry": [
	    {
	      "id": "PID123",
	      "changes": [
	        {
	          "field": "messages",
	          "value": {
	            "messaging_product": "whatsapp",
	            "metadata": {
	              "display_phone_number": "15550002222",
	              "phone_number_id": "678678678"
	            },
	            "contacts": [{
	              "profile": { "name": "ProductUser" },
	              "wa_id": "16789991111"
	            }],
	            "messages": [
	              {
	                "from": "16789991111",
	                "id": "wamid.PRODINQ001",
	                "text": { "body": "Interested in your product!" },
	                "context": {
	                  "from": "16789991111",
	                  "id": "wamid.IDCONTEXT1",
	                  "referred_product": {
	                    "catalog_id": "CATALOG_9999",
	                    "product_retailer_id": "SKU-1234"
	                  }
	                },
	                "timestamp": "1685000044",
	                "type": "text"
	              }
	            ]
	          }
	        }
	      ]
	    }
	  ]
	}`

	var productInquiryHandled bool

	handler := webhooks.NewHandler()
	handler.OnProductEnquiryMessage(webhooks.MessageHandlerFunc[webhooks.Text](func(ctx context.Context,
		req *webhooks.MessageRequest[webhooks.Text],
	) error {
		txt := req.Payload
		productInquiryHandled = true
		if txt.Body != "Interested in your product!" {
			t.Errorf("Product inquiry text mismatch, got=%s", txt.Body)
		}
		return nil
	}))

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{}, nil
		}),
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/webhook", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	if !productInquiryHandled {
		t.Error("Product inquiry was not handled")
	}
}

func TestListener_HandleNotification_UserChangedNumber(t *testing.T) {
	payload := `{
	  "object": "whatsapp_business_account",
	  "entry": [{
	    "id": "WHATSAPP_BUSINESS_ACCOUNT_ID",
	    "changes": [{
	      "field": "messages",
	      "value": {
	        "messaging_product": "whatsapp",
	        "metadata": {
	          "display_phone_number": "15550001111",
	          "phone_number_id": "1093849293"
	        },
	        "messages": [{
	          "from": "15550002222",
	          "id": "wamid.USERCHANGE",
	          "timestamp": "1689999999",
	          "system": {
	            "body": "NAME changed from 15550002222 to 15550003333",
	            "new_wa_id": "15550003333",
	            "type": "user_changed_number"
	          },
	          "type": "system"
	        }]
	      }
	    }]
	  }]
	}`

	var systemMessageHandled bool

	handler := webhooks.NewHandler()
	handler.OnSystemMessage(webhooks.MessageHandlerFunc[webhooks.System](func(ctx context.Context,
		req *webhooks.MessageRequest[webhooks.System],
	) error {
		sys := req.Payload
		systemMessageHandled = true
		if sys.Type != "user_changed_number" {
			t.Errorf("System message type mismatch, got=%s", sys.Type)
		}
		return nil
	}))

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{}, nil
		}),
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/webhook", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if !systemMessageHandled {
		t.Error("System message (user changed number) was not handled")
	}
}

func TestListener_HandleNotification_StatusSent(t *testing.T) {
	payload := `{
	  "object": "whatsapp_business_account",
	  "entry": [
	    {
	      "id": "WHATSAPP_BUSINESS_ACCOUNT_ID",
	      "changes": [
	        {
	          "field": "messages",
	          "value": {
	            "messaging_product": "whatsapp",
	            "metadata": {
	              "display_phone_number": "15555551234",
	              "phone_number_id": "999888777"
	            },
	            "statuses": [
	              {
	                "id": "wamid.STATUSMSG",
	                "status": "sent",
	                "timestamp": "1690000000",
	                "recipient_id": "123456789",
	                "conversation": {
	                  "id": "CONVO123",
	                  "expiration_timestamp": "1690009999",
	                  "origin": {
	                    "type": "user_initiated"
	                  }
	                },
	                "pricing": {
	                  "billable": true,
	                  "pricing_model": "CBP",
	                  "category": "user_initiated"
	                }
	              }
	            ]
	          }
	        }
	      ]
	    }
	  ]
	}`

	var statusChangeHandled bool

	handler := webhooks.NewHandler()
	handler.OnMessageStatusChange(
		webhooks.ChangeValueHandlerFunc[webhooks.Status](
			func(ctx context.Context, req *webhooks.ChangeValueRequest[webhooks.Status]) error {
				statuses := req.Payload
				statusChangeHandled = true

				if len(statuses) != 1 {
					t.Errorf("Expected 1 status, got %d", len(statuses))
					return nil
				}
				st := statuses[0]
				if st.StatusValue != "sent" {
					t.Errorf("StatusValue mismatch, got=%s, want=sent", st.StatusValue)
				}
				return nil
			},
		),
	)

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{}, nil
		}),
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/webhook", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	if !statusChangeHandled {
		t.Error("Message status (sent) was not handled")
	}
}

func TestListener_HandleNotification_DeletedMessageUnsupported(t *testing.T) {
	payload := `{
	  "object": "whatsapp_business_account",
	  "entry": [{
	    "id": "WHATSAPP_BUSINESS_ACCOUNT_ID",
	    "changes": [{
	      "value": {
	        "messaging_product": "whatsapp",
	        "metadata": {
	          "display_phone_number": "15550001111",
	          "phone_number_id": "100200300"
	        },
	        "contacts": [{
	          "profile": { "name": "NAME" },
	          "wa_id": "15550001111"
	        }],
	        "messages": [{
	          "from": "15550001111",
	          "id": "wamid.ID",
	          "timestamp": "1690000100",
	          "errors": [{
	            "code": 131051,
	            "details": "Message type is not currently supported",
	            "title": "Unsupported message type"
	          }],
	          "type": "unsupported"
	        }]
	      },
	      "field": "messages"
	    }]
	  }]
	}`

	var messageDeletionHandled bool

	handler := webhooks.NewHandler()
	handler.OnUnsupportedMessage(webhooks.MessageErrorsHandlerFunc(func(ctx context.Context,
		req *webhooks.MessageRequest[webhooks.Message],
		errors []*werrors.Error,
	) error {
		errs := errors
		messageDeletionHandled = true
		if len(errs) != 1 {
			t.Errorf("Expected 1 error, got=%d", len(errs))
			return nil
		}
		if errs[0].Code != 131051 {
			t.Errorf("Error code mismatch, got=%d, want=131051", errs[0].Code)
		}
		if errs[0].Details != "Message type is not currently supported" {
			t.Errorf("Error message mismatch, got=%s", errs[0].Message)
		}
		return nil
	}))

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{}, nil
		}),
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/webhook", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	if !messageDeletionHandled {
		t.Error("Deleted (unsupported) message was not handled by OnMessageErrors")
	}
}

func TestListener_HandleNotification_PhoneNumberSettingsUpdate(t *testing.T) {
	payload := `{
    "object": "whatsapp_business_account",
    "entry": [
        {
            "id": "whatsapp-business-account-id",
            "changes": [
                {
                    "value": {
                        "messaging_product": "whatsapp",
                        "timestamp": "1671644824",
                        "type": "[phone_number_settings]",
                        "phone_number_settings": {
                            "phone_number_id": "TEST987654321",
                            "calling": {
                                "status": "ENABLED",
                                "call_icon_visibility": "DEFAULT",
                                "callback_permission_status": "ENABLED",
                                "call_hours": {
                                    "status": "ENABLED",
                                    "timezone_id": "[REDACTED]",
                                    "weekly_operating_hours": [
                                        {
                                            "day_of_week": "MONDAY",
                                            "open_time": "0400",
                                            "close_time": "1020"
                                        },
                                        {
                                            "day_of_week": "TUESDAY",
                                            "open_time": "0108",
                                            "close_time": "1020"
                                        }
                                    ],
                                    "holiday_schedule": [
                                        {
                                            "date": "2026-01-01",
                                            "start_time": "0000",
                                            "end_time": "2359"
                                        }
                                    ]
                                },
                                "sip": {
                                    "status": "ENABLED",
                                    "servers": [
                                        {
                                            "hostname": "example.com",
                                            "port": 9000
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "field": "account_settings_update"
                }
            ]
        }
    ]
}`

	var eventHandled bool
	handler := webhooks.NewHandler()

	handler.OnPhoneSettingsUpdate(
		webhooks.BusinessEventHandlerFunc[webhooks.PhoneNumberSettings](
			func(ctx context.Context, req *webhooks.BusinessRequest[webhooks.PhoneNumberSettings]) error {
				details := req.Payload
				eventHandled = true

				if details.PhoneNumberID != "TEST987654321" {
					t.Errorf("PhoneNumberID mismatch, got=%s", details.PhoneNumberID)
				}

				return nil
			},
		),
	)

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{}, nil
		}),
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/webhook", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	if !eventHandled {
		t.Error("phone settings update was not handled by OnPhoneSettingsUpdate")
	}
}

func TestListener_HandleNotification_GroupLifecycleUpdate(t *testing.T) {
	payload := `{
  "object": "whatsapp_business_account",
  "entry": [
    {
      "id": "102290129340398",
      "changes": [
        {
          "value": {
            "messaging_product": "whatsapp",
            "metadata": {
              "display_phone_number": "15550783881",
              "phone_number_id": "106540352242922"
            },
            "groups": [
              {
                "timestamp": "1739321024",
                "group_id": "GROUP_ID_123",
                "type": "group_create",
                "request_id": "REQ_001",
                "subject": "SDK Test Group",
                "invite_link": "https://chat.whatsapp.com/ABC123",
                "join_approval_mode": "auto_approve"
              }
            ]
          },
          "field": "group_lifecycle_update"
        }
      ]
    }
  ]
}`

	var handled bool
	handler := webhooks.NewHandler()
	handler.OnGroupLifecycleUpdate(
		webhooks.ChangeValueHandlerFunc[webhooks.Group](
			func(ctx context.Context, req *webhooks.ChangeValueRequest[webhooks.Group]) error {
				groups := req.Payload
				handled = true
				if len(groups) != 1 {
					t.Errorf("expected 1 group, got %d", len(groups))
					return nil
				}
				g := groups[0]
				if g.GroupID != "GROUP_ID_123" {
					t.Errorf("GroupID = %q, want %q", g.GroupID, "GROUP_ID_123")
				}
				if g.Type != "group_create" {
					t.Errorf("Type = %q, want %q", g.Type, "group_create")
				}
				if g.Subject != "SDK Test Group" {
					t.Errorf("Subject = %q, want %q", g.Subject, "SDK Test Group")
				}
				if g.InviteLink != "https://chat.whatsapp.com/ABC123" {
					t.Errorf("InviteLink = %q", g.InviteLink)
				}
				return nil
			},
		),
	)

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{}, nil
		}),
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/webhook", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if !handled {
		t.Error("group_lifecycle_update was not handled by OnGroupLifecycleUpdate")
	}
}

func TestListener_HandleNotification_GroupParticipantsUpdate(t *testing.T) {
	payload := `{
  "object": "whatsapp_business_account",
  "entry": [
    {
      "id": "102290129340398",
      "changes": [
        {
          "value": {
            "messaging_product": "whatsapp",
            "metadata": {
              "display_phone_number": "15550783881",
              "phone_number_id": "106540352242922"
            },
            "groups": [
              {
                "timestamp": "1739321024",
                "group_id": "GROUP_ID_456",
                "type": "group_participants_add",
                "reason": "invite_link",
                "added_participants": [
                  {
                    "wa_id": "16505551234"
                  }
                ]
              }
            ]
          },
          "field": "group_participants_update"
        }
      ]
    }
  ]
}`

	var handled bool
	handler := webhooks.NewHandler()
	handler.OnGroupParticipantsUpdate(
		webhooks.ChangeValueHandlerFunc[webhooks.Group](
			func(ctx context.Context, req *webhooks.ChangeValueRequest[webhooks.Group]) error {
				groups := req.Payload
				handled = true
				if len(groups) != 1 {
					t.Errorf("expected 1 group, got %d", len(groups))
					return nil
				}
				g := groups[0]
				if g.Type != "group_participants_add" {
					t.Errorf("Type = %q, want group_participants_add", g.Type)
				}
				if g.Reason != "invite_link" {
					t.Errorf("Reason = %q, want invite_link", g.Reason)
				}
				if len(g.AddedParticipants) != 1 || g.AddedParticipants[0].WaID != "16505551234" {
					t.Errorf("AddedParticipants mismatch: %+v", g.AddedParticipants)
				}
				return nil
			},
		),
	)

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{}, nil
		}),
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/webhook", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if !handled {
		t.Error("group_participants_update was not handled by OnGroupParticipantsUpdate")
	}
}

func TestListener_HandleNotification_GroupSettingsUpdate(t *testing.T) {
	payload := `{
  "object": "whatsapp_business_account",
  "entry": [
    {
      "id": "102290129340398",
      "changes": [
        {
          "value": {
            "messaging_product": "whatsapp",
            "metadata": {
              "display_phone_number": "15550783881",
              "phone_number_id": "106540352242922"
            },
            "groups": [
              {
                "timestamp": "1739321024",
                "group_id": "GROUP_ID_789",
                "type": "group_settings_update",
                "request_id": "REQ_003",
                "group_subject": {
                  "text": "New Subject",
                  "update_successful": true
                },
                "group_description": {
                  "text": "New Description",
                  "update_successful": false,
                  "errors": [
                    {
                      "code": 100,
                      "message": "Invalid description",
                      "title": "Update Failed",
                      "error_data": {"details": "Description too long"}
                    }
                  ]
                }
              }
            ]
          },
          "field": "group_settings_update"
        }
      ]
    }
  ]
}`

	var handled bool
	handler := webhooks.NewHandler()
	handler.OnGroupSettingsUpdate(
		webhooks.ChangeValueHandlerFunc[webhooks.Group](
			func(ctx context.Context, req *webhooks.ChangeValueRequest[webhooks.Group]) error {
				groups := req.Payload
				handled = true
				if len(groups) != 1 {
					t.Errorf("expected 1 group, got %d", len(groups))
					return nil
				}
				g := groups[0]
				if g.Type != "group_settings_update" {
					t.Errorf("Type = %q, want group_settings_update", g.Type)
				}
				if g.GroupSubject == nil || g.GroupSubject.Text != "New Subject" || !g.GroupSubject.UpdateSuccessful {
					t.Errorf("GroupSubject mismatch: %+v", g.GroupSubject)
				}
				if g.GroupDescription == nil || g.GroupDescription.UpdateSuccessful {
					t.Errorf("GroupDescription should have failed: %+v", g.GroupDescription)
				}
				if len(g.GroupDescription.Errors) != 1 {
					t.Errorf("expected 1 error on GroupDescription, got %d", len(g.GroupDescription.Errors))
				}
				return nil
			},
		),
	)

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{}, nil
		}),
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/webhook", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if !handled {
		t.Error("group_settings_update was not handled by OnGroupSettingsUpdate")
	}
}

func TestListener_HandleNotification_GroupStatusUpdate(t *testing.T) {
	payload := `{
  "object": "whatsapp_business_account",
  "entry": [
    {
      "id": "102290129340398",
      "changes": [
        {
          "value": {
            "messaging_product": "whatsapp",
            "metadata": {
              "display_phone_number": "15550783881",
              "phone_number_id": "106540352242922"
            },
            "groups": [
              {
                "timestamp": "1739321024",
                "type": "group_suspend",
                "group_id": "GROUP_ID_999"
              }
            ]
          },
          "field": "group_status_update"
        }
      ]
    }
  ]
}`

	var handled bool
	handler := webhooks.NewHandler()
	handler.OnGroupStatusUpdate(
		webhooks.ChangeValueHandlerFunc[webhooks.Group](
			func(ctx context.Context, req *webhooks.ChangeValueRequest[webhooks.Group]) error {
				groups := req.Payload
				handled = true
				if len(groups) != 1 {
					t.Errorf("expected 1 group, got %d", len(groups))
					return nil
				}
				g := groups[0]
				if g.Type != "group_suspend" {
					t.Errorf("Type = %q, want group_suspend", g.Type)
				}
				if g.GroupID != "GROUP_ID_999" {
					t.Errorf("GroupID = %q, want GROUP_ID_999", g.GroupID)
				}
				return nil
			},
		),
	)

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{}, nil
		}),
	}
	ts := NewTestWebhookServer(t, cfg)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/webhook", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if !handled {
		t.Error("group_status_update was not handled by OnGroupStatusUpdate")
	}
}

func TestListener_HandleNotification_UnrecognizedField(t *testing.T) {
	t.Parallel()

	// A field not in the KnownChangeFields list — simulates a future WhatsApp
	// notification type the library doesn't explicitly handle.
	payload := `{
  "object": "whatsapp_business_account",
  "entry": [{
    "id": "WHATSAPP_BUSINESS_ACCOUNT_ID",
    "changes": [{
      "field": "automatic_events",
      "value": {
        "event": "some_future_event",
        "details": {"key": "value"}
      }
    }]
  }]
}`

	var handled bool
	var gotField string

	handler := webhooks.NewHandler()
	handler.OnFallback(webhooks.FallbackHandlerFunc(func(ctx context.Context,
		ev webhooks.NotificationEvent,
	) error {
		handled = true
		gotField = ev.Field
		return nil
	}))

	reader := webhooks.ConfigReaderFunc(func(r *http.Request) (*webhooks.Config, error) {
		return &webhooks.Config{Token: "test", ValidatePayload: false}, nil
	})

	server := NewTestWebhookServer(t, TestServerConfig{
		Handler:      handler,
		ConfigReader: reader,
	})
	defer server.Close()

	resp, err := http.Post(server.URL, "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if !handled {
		t.Error("unrecognized field handler was not invoked for 'automatic_events'")
	}
	if gotField != "automatic_events" {
		t.Errorf("expected field 'automatic_events', got %q", gotField)
	}
}

func TestHandler_UnrecognizedField_DefaultSilent(t *testing.T) {
	t.Parallel()

	// Without OnUnrecognizedField, the handler should return 200 silently.
	payload := `{
  "object": "whatsapp_business_account",
  "entry": [{
    "id": "WHATSAPP_BUSINESS_ACCOUNT_ID",
    "changes": [{
      "field": "unknown_field_xyz",
      "value": {}
    }]
  }]
}`

	handler := webhooks.NewHandler()
	// No OnUnrecognizedField set — default nil behaviour.

	reader := webhooks.ConfigReaderFunc(func(r *http.Request) (*webhooks.Config, error) {
		return &webhooks.Config{Token: "test", ValidatePayload: false}, nil
	})

	server := NewTestWebhookServer(t, TestServerConfig{
		Handler:      handler,
		ConfigReader: reader,
	})
	defer server.Close()

	resp, err := http.Post(server.URL, "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 (silent ack for unknown field), got %d", resp.StatusCode)
	}
}

// Example demonstrates creating a Handler and registering handlers for
// WhatsApp webhook notifications.
func Example() {
	h := webhooks.NewHandler()
	_ = h
	fmt.Println("handler created")
	// Output: handler created
}

// FuzzNotificationUnmarshal verifies that Notification unmarshaling never
// panics on arbitrary input.
func FuzzNotificationUnmarshal(f *testing.F) {
	f.Add([]byte(`{"object":"whatsapp_business_account","entry":[]}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Add([]byte(`{"object": "not_whatsapp"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var n webhooks.Notification
		_ = json.Unmarshal(data, &n)
	})
}

func FuzzValueUnmarshal(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"messaging_product":"whatsapp"}`))
	f.Add([]byte(`{"statuses":[{"status":"sent"}]}`))
	f.Add([]byte(`{"errors":[{"code":400}]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var v webhooks.Value
		_ = json.Unmarshal(data, &v)
	})
}

func FuzzMessageUnmarshal(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"type":"text","text":{"body":"hello"}}`))
	f.Add([]byte(`{"type":"unknown_xyz","from":"123"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var m webhooks.Message
		_ = json.Unmarshal(data, &m)
	})
}

func FuzzExtractAndValidatePayload(f *testing.F) {
	f.Add([]byte(`{"object":"whatsapp_business_account","entry":[]}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`not json`))

	f.Fuzz(func(t *testing.T, body []byte) {
		r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		_, _ = webhooks.ParseNotification(r, &webhooks.ParseNotificationOptions{
			VerifyPayloadSignature: false,
		})
	})
}

func FuzzHandleNotification(f *testing.F) {
	f.Add([]byte(`{"object":"whatsapp_business_account","entry":[]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var n webhooks.Notification
		if err := json.Unmarshal(data, &n); err != nil {
			return
		}

		h := webhooks.NewHandler()
		resp := h.HandleNotification(context.Background(), &n)
		if resp.StatusCode != http.StatusOK &&
			resp.StatusCode != http.StatusInternalServerError &&
			resp.StatusCode != http.StatusGatewayTimeout {
			t.Errorf("unexpected status %d", resp.StatusCode)
		}
	})
}

func TestDefect002_MiddlewarePanic_NotRecovered(t *testing.T) {
	t.Parallel()

	handler := webhooks.NewHandler()
	cfgReader := webhooks.ConfigReaderFunc(func(r *http.Request) (*webhooks.Config, error) {
		return &webhooks.Config{ValidatePayload: false}, nil
	})

	panickingMiddleware := func(next webhooks.NotificationHandler) webhooks.NotificationHandler {
		return webhooks.NotificationHandlerFunc(
			func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
				panic("middleware bug")
			},
		)
	}

	listener := webhooks.NewListener(handler, cfgReader, panickingMiddleware)

	payload := `{"object":"whatsapp_business_account","entry":[]}`
	r := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(payload))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		listener.HandleNotification(w, r)
	}()

	if didPanic {
		t.Error("FIXME: Listener.HandleNotification panicked on middleware panic. " +
			"Add recover() to Listener.HandleNotification to catch this.")
	}
}

func TestDefect003_NilListenerFields_Panics(t *testing.T) {
	t.Parallel()

	listener := &webhooks.Listener{}
	r := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		listener.HandleNotification(w, r)
	}()

	if didPanic {
		t.Error("FIXME: Zero-value Listener panics on HandleNotification. " +
			"Add nil checks for listener.handler and listener.configReader.")
	}
}

func TestDefect005_PanicRecovery_WrapsPanicAsPanicError(t *testing.T) {
	t.Parallel()

	handler := webhooks.NewHandler()
	handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
		func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
			panic("intentional test panic in text handler")
		},
	))

	notification := &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{{
			ID:   "1",
			Time: 1719000000,
			Changes: []webhooks.Change{{
				Field: "messages",
				Value: &webhooks.Value{
					MessagingProduct: "whatsapp",
					Contacts:         []*webhooks.Contact{{WaID: "12345", Profile: &webhooks.Profile{Name: "Test"}}},
					Metadata:         &webhooks.Metadata{PhoneNumberID: "123", DisplayPhoneNumber: "123456789"},
					Messages: []*webhooks.Message{{
						Type: "text",
						From: "12345",
						ID:   "msg1",
						Text: &webhooks.Text{Body: "hello"},
					}},
				},
			}},
		}},
	}

	resp := handler.HandleNotification(context.Background(), notification)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 for handler panic, got %d", resp.StatusCode)
	}
}

func callConnectPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "calls",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							Contacts: []*webhooks.Contact{
								{WaID: "15550001111", Profile: &webhooks.Profile{Name: "Test User"}},
							},
							Calls: []*webhooks.Call{
								{
									ID:        "wacid.test123",
									To:        "15550001111",
									From:      "15550783881",
									Event:     "connect",
									Timestamp: "1739321024",
									Direction: "USER_INITIATED",
									Session: &webhooks.CallSession{
										SDPType: "answer",
										SDP:     "v=0...",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func callTerminatePayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "calls",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							Calls: []*webhooks.Call{
								{
									ID:        "wacid.test456",
									To:        "15550001111",
									From:      "15550783881",
									Event:     "terminate",
									Timestamp: "1739321024",
									Direction: "USER_INITIATED",
									Status:    "COMPLETED",
									StartTime: "1739321000",
									EndTime:   "1739321120",
									Duration:  120,
								},
							},
						},
					},
				},
			},
		},
	}
}

func callStatusPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "calls",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							Statuses: []*webhooks.Status{
								{
									ID:          "wacid.status123",
									Timestamp:   "1739321024",
									Type:        "call",
									StatusValue: "RINGING",
									RecipientID: "15550001111",
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestFallback_Calls_NoSubHandler_Silent200(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	resp := h.HandleNotification(context.Background(), callConnectPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (calls silently skipped when sub-handler nil), got %d", resp.StatusCode)
	}
}

func TestFallback_Calls_NoSubHandler_GeneralFallbackFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	var fired bool
	var gotField string
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			fired = true
			gotField = ev.Field
			return nil
		},
	))
	resp := h.HandleNotification(context.Background(), callConnectPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("general fallback was not invoked for nil calls sub-handler")
	}
	if gotField != "calls" {
		t.Errorf("expected field 'calls', got %q", gotField)
	}
}

func TestFallback_Calls_DedicatedHandlerFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	var fired bool
	h.OnCallConnect(webhooks.CallsEventHandlerFunc[webhooks.Call](
		func(_ context.Context, req *webhooks.CallRequest[webhooks.Call]) error {
			fired = true
			return nil
		},
	))
	resp := h.HandleNotification(context.Background(), callConnectPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("dedicated handler was not invoked")
	}
}

func TestFallback_Calls_SubFallbackFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	h.OnCallConnect(webhooks.CallsEventHandlerFunc[webhooks.Call](
		func(_ context.Context, req *webhooks.CallRequest[webhooks.Call]) error {
			t.Error("connect handler should not fire for terminate event")
			return nil
		},
	))
	var subFired bool
	h.Calls().Fallback = webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			subFired = true
			return nil
		},
	)
	resp := h.HandleNotification(context.Background(), callTerminatePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !subFired {
		t.Fatal("sub-handler fallback was not invoked for unhandled call event")
	}
}

func TestFallback_Calls_NoSubFallback_Silent200(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	h.OnCallConnect(webhooks.CallsEventHandlerFunc[webhooks.Call](
		func(_ context.Context, req *webhooks.CallRequest[webhooks.Call]) error {
			return nil
		},
	))
	resp := h.HandleNotification(context.Background(), callTerminatePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

func TestFallback_Calls_OnFallbackPropagates(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	h.OnCallConnect(webhooks.CallsEventHandlerFunc[webhooks.Call](
		func(_ context.Context, req *webhooks.CallRequest[webhooks.Call]) error {
			return nil
		},
	))
	if h.Calls().Fallback != nil {
		t.Fatal("Calls.Fallback should be nil before OnFallback")
	}
	var fired bool
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			fired = true
			return nil
		},
	))
	if h.Calls().Fallback == nil {
		t.Fatal("OnFallback did not propagate to CallsHandler.Fallback")
	}
	resp := h.HandleNotification(context.Background(), callTerminatePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("propagated fallback was not invoked for unhandled call event")
	}
}

func TestFallback_Calls_DedicatedStatusHandlerFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	var fired bool
	h.OnCallStatus(webhooks.CallsEventHandlerFunc[webhooks.Status](
		func(_ context.Context, req *webhooks.CallRequest[webhooks.Status]) error {
			fired = true
			if req.Payload.Type != "call" {
				t.Errorf("expected status type 'call', got %q", req.Payload.Type)
			}
			return nil
		},
	))
	resp := h.HandleNotification(context.Background(), callStatusPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("dedicated call status handler was not invoked")
	}
}

func TestFallback_Calls_ErrorPropagationToErrorHandler(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	h.OnCallConnect(webhooks.CallsEventHandlerFunc[webhooks.Call](
		func(_ context.Context, req *webhooks.CallRequest[webhooks.Call]) error {
			return context.Canceled
		},
	))
	var gotErr error
	h.Calls().ErrorHandler = webhooks.ErrorHandlerFunc(
		func(_ context.Context, err error) error {
			gotErr = err
			return nil
		},
	)
	resp := h.HandleNotification(context.Background(), callConnectPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (non-fatal error), got %d", resp.StatusCode)
	}
	if gotErr == nil {
		t.Fatal("error handler was not invoked for handler error")
	}
}

func TestPanicError_Recovery(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	h.OnCallConnect(webhooks.CallsEventHandlerFunc[webhooks.Call](
		func(_ context.Context, req *webhooks.CallRequest[webhooks.Call]) error {
			panic("unexpected nil map in handler")
		},
	))
	resp := h.HandleNotification(context.Background(), callConnectPayload())
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 for panicking handler, got %d", resp.StatusCode)
	}
}

func TestPanicError_IsPanicError(t *testing.T) {
	t.Parallel()
	pe := &webhooks.PanicError{Value: "test panic"}
	got, ok := webhooks.IsPanicError(pe)
	if !ok {
		t.Fatal("IsPanicError returned false for a PanicError")
	}
	if got.Value != "test panic" {
		t.Errorf("Value = %v, want 'test panic'", got.Value)
	}
	_, ok = webhooks.IsPanicError(context.Canceled)
	if ok {
		t.Error("IsPanicError returned true for context.Canceled")
	}
}

func groupStatusUpdatePayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "group_status_update",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							Groups: []*webhooks.Group{
								{GroupID: "GROUP_ID_123", Type: "group_suspend"},
							},
						},
					},
				},
			},
		},
	}
}

func TestFallback_NoSubHandler_Silent200(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	resp := h.HandleNotification(context.Background(), groupStatusUpdatePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack for nil sub-handler), got %d", resp.StatusCode)
	}
}

func TestFallback_NoSubHandler_GeneralFallbackFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	var fired bool
	var gotField string
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			fired = true
			gotField = ev.Field
			return nil
		},
	))
	resp := h.HandleNotification(context.Background(), groupStatusUpdatePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("general fallback was not invoked for nil sub-handler")
	}
	if gotField != "group_status_update" {
		t.Errorf("expected field 'group_status_update', got %q", gotField)
	}
}

func TestFallback_DedicatedHandlerFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	var fired bool
	h.OnGroupStatusUpdate(webhooks.ChangeValueHandlerFunc[webhooks.Group](
		func(_ context.Context, req *webhooks.ChangeValueRequest[webhooks.Group]) error {
			fired = true
			return nil
		},
	))
	resp := h.HandleNotification(context.Background(), groupStatusUpdatePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("dedicated handler was not invoked")
	}
}

func TestFallback_SubHandlerFallbackFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	h.OnGroupLifecycleUpdate(webhooks.ChangeValueHandlerFunc[webhooks.Group](
		func(_ context.Context, req *webhooks.ChangeValueRequest[webhooks.Group]) error {
			t.Error("lifecycle handler should not fire for status update")
			return nil
		},
	))
	var subFallbackFired bool
	h.Groups().OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			subFallbackFired = true
			return nil
		},
	))
	resp := h.HandleNotification(context.Background(), groupStatusUpdatePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !subFallbackFired {
		t.Fatal("sub-handler fallback was not invoked for unhandled group field")
	}
}

func TestFallback_NoSubFallback_Silent200(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	h.OnGroupLifecycleUpdate(webhooks.ChangeValueHandlerFunc[webhooks.Group](
		func(_ context.Context, req *webhooks.ChangeValueRequest[webhooks.Group]) error {
			t.Error("lifecycle handler should not fire")
			return nil
		},
	))
	resp := h.HandleNotification(context.Background(), groupStatusUpdatePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

func TestFallback_OnFallbackPropagates(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	h.OnGroupLifecycleUpdate(webhooks.ChangeValueHandlerFunc[webhooks.Group](
		func(_ context.Context, req *webhooks.ChangeValueRequest[webhooks.Group]) error {
			return nil
		},
	))
	if h.Groups().Fallback != nil {
		t.Fatal("groups.Fallback should be nil before OnFallback")
	}
	var fallbackFired bool
	var fallbackField string
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			fallbackFired = true
			fallbackField = ev.Field
			return nil
		},
	))
	if h.Groups().Fallback == nil {
		t.Fatal("OnFallback did not propagate to groups.Fallback")
	}
	resp := h.HandleNotification(context.Background(), groupStatusUpdatePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fallbackFired {
		t.Fatal("fallback was not invoked for unhandled group field")
	}
	if fallbackField != "group_status_update" {
		t.Errorf("expected field 'group_status_update', got %q", fallbackField)
	}
}

func smbAppStateSyncPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "102290129340398",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "smb_app_state_sync",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							StateSync: []webhooks.SMBAppStateSync{
								{
									Type:   "contact",
									Action: "add",
									Contact: &webhooks.SMBContactSync{
										FullName:    "Pablo Morales",
										FirstName:   "Pablo",
										PhoneNumber: "16505551234",
									},
									Metadata: &webhooks.SMBMetadata{Timestamp: 1739321024},
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestFallback_SMBAppState_NoSubHandler_Silent200(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	resp := h.HandleNotification(context.Background(), smbAppStateSyncPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack for nil sub-handler), got %d", resp.StatusCode)
	}
}

func TestFallback_SMBAppState_NoSubHandler_GeneralFallbackFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	var fired bool
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			fired = true
			return nil
		},
	))
	resp := h.HandleNotification(context.Background(), smbAppStateSyncPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("general fallback was not invoked for nil sub-handler")
	}
}

func TestFallback_SMBAppState_DedicatedHandlerFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	var gotAction string
	h.OnSMBAppStateSync(webhooks.SMBAppStateSyncHandlerFunc(
		func(_ context.Context, nctx *webhooks.MessageNotificationContext, s *webhooks.SMBAppStateSync) error {
			gotAction = s.Action
			return nil
		},
	))
	resp := h.HandleNotification(context.Background(), smbAppStateSyncPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if gotAction != "add" {
		t.Errorf("expected action 'add', got %q", gotAction)
	}
}

func TestFallback_SMBAppState_SubFallbackFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	var subFired bool
	h.SMBAppSync().Fallback = webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			subFired = true
			return nil
		},
	)
	resp := h.HandleNotification(context.Background(), smbAppStateSyncPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !subFired {
		t.Fatal("sub-handler fallback was not invoked when Handler is nil")
	}
}

func TestFallback_SMBAppState_NoSubFallback_Silent200(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	resp := h.HandleNotification(context.Background(), smbAppStateSyncPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

func TestFallback_SMBAppState_OnFallbackPropagates(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	h.OnSMBAppStateSync(webhooks.SMBAppStateSyncHandlerFunc(
		func(_ context.Context, nctx *webhooks.MessageNotificationContext, s *webhooks.SMBAppStateSync) error {
			return nil
		},
	))
	if h.SMBAppSync().Fallback != nil {
		t.Fatal("Fallback should be nil before OnFallback")
	}
	var fired bool
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			fired = true
			return nil
		},
	))
	if h.SMBAppSync().Fallback == nil {
		t.Fatal("OnFallback did not propagate to SMBAppSync.Fallback")
	}
	h.SMBAppSync().Handler = nil
	resp := h.HandleNotification(context.Background(), smbAppStateSyncPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("propagated fallback was not invoked")
	}
}

func smbTextEchoPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "102290129340398",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "smb_message_echoes",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							MessageEchoes: []*webhooks.Message{
								{
									From:      "15550783881",
									To:        "16505551234",
									ID:        "wamid.test123",
									Timestamp: "1739321024",
									Type:      "text",
									Text:      &webhooks.Text{Body: "Hello from business app"},
								},
							},
						},
					},
				},
			},
		},
	}
}

func smbRevokeEchoPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "102290129340398",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "smb_message_echoes",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							MessageEchoes: []*webhooks.Message{
								{
									From:      "15550783881",
									To:        "16505551234",
									ID:        "wamid.test456",
									Timestamp: "1749854575",
									Type:      "revoke",
									Revoke: &webhooks.Revoke{
										OriginalMessageID: "wamid.original123",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestFallback_SMB_NoSubHandler_Silent200(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	resp := h.HandleNotification(context.Background(), smbTextEchoPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (SMB echoes silently skipped when sub-handler nil), got %d", resp.StatusCode)
	}
}

func TestFallback_SMB_NoSubHandler_GeneralFallbackFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	var fired bool
	var gotField string
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			fired = true
			gotField = ev.Field
			return nil
		},
	))
	resp := h.HandleNotification(context.Background(), smbTextEchoPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("general fallback was not invoked for nil SMB echo sub-handler")
	}
	if gotField != "smb_message_echoes" {
		t.Errorf("expected field 'smb_message_echoes', got %q", gotField)
	}
}

func TestFallback_SMB_DedicatedHandlerFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	var gotMsgType string
	h.OnSMBMessageEcho(webhooks.SMBMessageEchoHandlerFunc(
		func(_ context.Context, nctx *webhooks.MessageNotificationContext, msg *webhooks.Message) error {
			gotMsgType = msg.Type
			return nil
		},
	))
	resp := h.HandleNotification(context.Background(), smbTextEchoPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if gotMsgType != "text" {
		t.Errorf("expected msg type 'text', got %q", gotMsgType)
	}
}

func TestFallback_SMB_SubFallbackFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	var subFired bool
	h.SMBEchoes().Fallback = webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			subFired = true
			return nil
		},
	)
	resp := h.HandleNotification(context.Background(), smbTextEchoPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !subFired {
		t.Fatal("sub-handler fallback was not invoked when Handler is nil")
	}
}

func TestFallback_SMB_NoSubFallback_Silent200(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	resp := h.HandleNotification(context.Background(), smbRevokeEchoPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

func TestFallback_SMB_OnFallbackPropagates(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	h.OnSMBMessageEcho(webhooks.SMBMessageEchoHandlerFunc(
		func(_ context.Context, nctx *webhooks.MessageNotificationContext, msg *webhooks.Message) error {
			return nil
		},
	))
	if h.SMBEchoes().Fallback != nil {
		t.Fatal("SMBEchoes.Fallback should be nil before OnFallback")
	}
	var fired bool
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			fired = true
			return nil
		},
	))
	if h.SMBEchoes().Fallback == nil {
		t.Fatal("OnFallback did not propagate to SMBEchoes.Fallback")
	}
	h.SMBEchoes().Handler = nil
	resp := h.HandleNotification(context.Background(), smbTextEchoPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("propagated fallback was not invoked")
	}
}

func userPreferencesPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "102290129340398",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "user_preferences",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							Contacts: []*webhooks.Contact{
								{WaID: "16505551234"},
							},
							UserPreferences: []*webhooks.UserPreference{
								{
									WaID:      "16505551234",
									Detail:    "User requested to stop marketing messages",
									Category:  "marketing_messages",
									Value:     "stop",
									Timestamp: "1731705721",
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestFallback_UserPrefs_NoSubHandler_Silent200(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	resp := h.HandleNotification(context.Background(), userPreferencesPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack for nil sub-handler), got %d", resp.StatusCode)
	}
}

func TestFallback_UserPrefs_NoSubHandler_GeneralFallbackFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	var fired bool
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			fired = true
			return nil
		},
	))
	resp := h.HandleNotification(context.Background(), userPreferencesPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("general fallback was not invoked for nil sub-handler")
	}
}

func TestFallback_UserPrefs_DedicatedHandlerFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	var gotValue string
	h.OnUserPreferencesUpdate(webhooks.UserPreferenceHandlerFunc(
		func(_ context.Context, nctx *webhooks.MessageNotificationContext, p *webhooks.UserPreference) error {
			gotValue = p.Value
			return nil
		},
	))
	resp := h.HandleNotification(context.Background(), userPreferencesPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if gotValue != "stop" {
		t.Errorf("expected 'stop', got %q", gotValue)
	}
}

func TestFallback_UserPrefs_SubFallbackFires(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	var subFired bool
	h.UserPrefs().Fallback = webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			subFired = true
			return nil
		},
	)
	resp := h.HandleNotification(context.Background(), userPreferencesPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !subFired {
		t.Fatal("sub-handler fallback was not invoked when Handler is nil")
	}
}

func TestFallback_UserPrefs_NoSubFallback_Silent200(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	resp := h.HandleNotification(context.Background(), userPreferencesPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

func TestFallback_UserPrefs_OnFallbackPropagates(t *testing.T) {
	t.Parallel()
	h := webhooks.NewHandler()
	h.OnUserPreferencesUpdate(webhooks.UserPreferenceHandlerFunc(
		func(_ context.Context, nctx *webhooks.MessageNotificationContext, p *webhooks.UserPreference) error {
			return nil
		},
	))
	if h.UserPrefs().Fallback != nil {
		t.Fatal("Fallback should be nil before OnFallback")
	}
	var fired bool
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			fired = true
			return nil
		},
	))
	if h.UserPrefs().Fallback == nil {
		t.Fatal("OnFallback did not propagate to UserPrefs.Fallback")
	}
	h.UserPrefs().Handler = nil
	resp := h.HandleNotification(context.Background(), userPreferencesPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("propagated fallback was not invoked")
	}
}
