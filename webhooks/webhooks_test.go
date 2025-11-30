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
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/piusalfred/whatsapp/message"
	werrors "github.com/piusalfred/whatsapp/pkg/errors"
	"github.com/piusalfred/whatsapp/webhooks"
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
		return nil, errors.New("client ID is required")
	}

	clientName, ok := r.ids[clientID]
	if !ok {
		return nil, errors.New("client ID is not registered")
	}

	env := request.Header.Get("X-Client-Env")
	if env == "" {
		return nil, errors.New("client environment is required")
	}

	switch env {
	case "dev":
		cfg, ok := r.envConfig.Dev[clientName]
		if !ok {
			return nil, fmt.Errorf("client %s is not registered in dev environment", clientName)
		}
		return cfg, nil
	case "stg":
		cfg, ok := r.envConfig.Stg[clientName]
		if !ok {
			return nil, fmt.Errorf("client %s is not registered in stg environment", clientName)
		}
		return cfg, nil
	default:
		cfg, ok := r.envConfig.Prod[clientName]
		if !ok {
			return nil, fmt.Errorf("client %s is not registered in prod environment", clientName)
		}
		return cfg, nil
	}
}

func ExampleConfigReader_ReadConfig() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	envConfig := TestEnvConfig{
		Dev: TestConfigMap{
			"acme": {
				Token:     "acme-dev-token",
				Validate:  false,
				AppSecret: "acme-dev-app-secret",
			},
			"shield": {
				Token:     "shield-dev-token",
				Validate:  false,
				AppSecret: "shield-dev-app-secret",
			},
		},
		Stg: TestConfigMap{
			"acme": {
				Token:     "acme-stg-token",
				Validate:  false,
				AppSecret: "acme-stg-app-secret",
			},
			"shield": {
				Token:     "shield-stg-token",
				Validate:  false,
				AppSecret: "shield-stg-app-secret",
			},
		},
		Prod: TestConfigMap{
			"acme": {
				Token:     "acme-prod-token",
				Validate:  true,
				AppSecret: "acme-prod-app-secret",
			},
			"shield": {
				Token:     "shield-prod-token",
				Validate:  true,
				AppSecret: "shield-prod-app-secret",
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
		fmt.Printf("Validate: %t\n", cfg.Validate)
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
	// Validate: false
	// AppSecret: acme-dev-app-secret
	// Token: shield-prod-token
	// Validate: true
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
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{
				Token:    "dummy-verify-token",
				Validate: false,
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
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{
				Token:    "dummy-verify-token",
				Validate: false,
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
	handler.OnTextMessage(
		func(ctx context.Context, nctx *webhooks.MessageNotificationContext, mctx *webhooks.MessageInfo, txt *webhooks.Text) error {
			textHandled = true
			if txt.Body != "Hello from a text message!" {
				t.Errorf("Text body mismatch, got=%s, want=%s", txt.Body, "Hello from a text message!")
			}
			return nil
		},
	)
	handler.OnLocationMessage(
		func(ctx context.Context, nctx *webhooks.MessageNotificationContext, mctx *webhooks.MessageInfo, loc *message.Location) error {
			locationHandled = true
			if loc.Name != "San Francisco" {
				t.Errorf("Location name mismatch, got=%s, want=%s", loc.Name, "San Francisco")
			}
			return nil
		},
	)

	handler.OnReactionMessage(
		func(ctx context.Context, nctx *webhooks.MessageNotificationContext, mctx *webhooks.MessageInfo, reaction *message.Reaction) error {
			reactionHandled = true
			if reaction.Emoji != "👍" {
				t.Errorf("Reaction emoji mismatch, got=%s, want=%s", reaction.Emoji, "👍")
			}
			return nil
		},
	)

	handler.OnUserPreferencesUpdate(
		func(ctx context.Context, nctx *webhooks.MessageNotificationContext, prefs []*webhooks.UserPreference) error {
			userPreferencesSeen = true
			if len(prefs) != 2 {
				t.Errorf("Expected 2 user preferences, got %d", len(prefs))
				return nil
			}
			for _, p := range prefs {
				if p.Value != "stop" {
					t.Errorf("Preference mismatch, got=%s, want=stop", p.Value)
				}
			}
			return nil
		})

	handler.OnStickerMessage(
		func(ctx context.Context, nctx *webhooks.MessageNotificationContext, mctx *webhooks.MessageInfo, sticker *message.MediaInfo) error {
			stickerHandled = true
			if sticker.MimeType != "image/webp" {
				t.Errorf("Sticker mime type mismatch, got=%s, want=%s", sticker.MimeType, "image/webp")
			}

			return nil
		},
	)

	cfg := TestServerConfig{
		Handler: handler,
		ConfigReader: webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
			return &webhooks.Config{
				Token:    "dummy-verify-token",
				Validate: false,
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
	handler.OnButtonMessage(func(ctx context.Context,
		nctx *webhooks.MessageNotificationContext,
		mctx *webhooks.MessageInfo,
		btn *webhooks.Button,
	) error {
		buttonHandled = true

		if btn.Text != "No" {
			t.Errorf("Expected button text='No', got=%s", btn.Text)
		}
		if btn.Payload != "No-Button-Payload" {
			t.Errorf("Expected payload='No-Button-Payload', got=%s", btn.Payload)
		}
		return nil
	})

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
	handler.OnListReplyMessage(func(ctx context.Context,
		nctx *webhooks.MessageNotificationContext,
		mctx *webhooks.MessageInfo,
		lr *webhooks.ListReply,
	) error {
		listReplyHandled = true
		if lr.ID != "list_reply_id" {
			t.Errorf("ListReply ID mismatch, got=%s", lr.ID)
		}
		if lr.Title != "list_reply_title" {
			t.Errorf("ListReply Title mismatch, got=%s", lr.Title)
		}
		return nil
	})

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
	handler.OnButtonReplyMessage(func(ctx context.Context,
		nctx *webhooks.MessageNotificationContext,
		mctx *webhooks.MessageInfo,
		btn *webhooks.ButtonReply,
	) error {
		buttonReplyHandled = true
		if btn.ID != "unique-button-identifier-here" {
			t.Errorf("ButtonReply ID mismatch, got=%s", btn.ID)
		}
		if btn.Title != "button-text" {
			t.Errorf("ButtonReply Title mismatch, got=%s", btn.Title)
		}
		return nil
	})

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
	handler.OnReferralMessage(func(ctx context.Context,
		nctx *webhooks.MessageNotificationContext,
		mctx *webhooks.MessageInfo,
		ref *webhooks.ReferralNotification,
	) error {
		referralHandled = true

		if ref.Text.Body != "Hi from an ad click!" {
			t.Errorf("Referral text mismatch, got=%s", ref.Text.Body)
		}
		if ref.Referral.SourceID != "ADID123" {
			t.Errorf("Referral sourceID mismatch, got=%s", ref.Referral.SourceID)
		}
		return nil
	})

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
	handler.OnProductEnquiryMessage(func(ctx context.Context,
		nctx *webhooks.MessageNotificationContext,
		mctx *webhooks.MessageInfo,
		txt *webhooks.Text,
	) error {
		productInquiryHandled = true
		if txt.Body != "Interested in your product!" {
			t.Errorf("Product inquiry text mismatch, got=%s", txt.Body)
		}
		return nil
	})

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
	handler.OnSystemMessage(func(ctx context.Context,
		nctx *webhooks.MessageNotificationContext,
		mctx *webhooks.MessageInfo,
		sys *webhooks.System,
	) error {
		systemMessageHandled = true
		if sys.Type != "user_changed_number" {
			t.Errorf("System message type mismatch, got=%s", sys.Type)
		}
		return nil
	})

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
		func(ctx context.Context, nctx *webhooks.MessageNotificationContext, statuses []*webhooks.Status) error {
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
	handler.OnUnsupportedMessage(func(ctx context.Context,
		nctx *webhooks.MessageNotificationContext,
		mctx *webhooks.MessageInfo,
		errs []*werrors.Error,
	) error {
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
	})

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
		func(ctx context.Context, notificationContext *webhooks.BusinessNotificationContext, details *webhooks.PhoneNumberSettings) error {
			eventHandled = true

			if details.PhoneNumberID != "TEST987654321" {
				t.Errorf("PhoneNumberID mismatch, got=%s", details.PhoneNumberID)
			}

			return nil
		},
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
