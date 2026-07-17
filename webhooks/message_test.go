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
	"context"
	"net/http"
	"testing"

	"github.com/piusalfred/whatsapp/webhooks"
)

func TestParseMessageType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected webhooks.MessageType
	}{
		{"text", webhooks.MessageTypeText},
		{"image", webhooks.MessageTypeImage},
		{"audio", webhooks.MessageTypeAudio},
		{"video", webhooks.MessageTypeVideo},
		{"document", webhooks.MessageTypeDocument},
		{"sticker", webhooks.MessageTypeSticker},
		{"location", webhooks.MessageTypeLocation},
		{"contacts", webhooks.MessageTypeContacts},
		{"reaction", webhooks.MessageTypeReaction},
		{"interactive", webhooks.MessageTypeInteractive},
		{"button", webhooks.MessageTypeButton},
		{"order", webhooks.MessageTypeOrder},
		{"system", webhooks.MessageTypeSystem},
		{"request_welcome", webhooks.MessageTypeRequestWelcome},
		{"unsupported", webhooks.MessageTypeUnsupported},
		{"unknown", webhooks.MessageTypeUnknown},
		{"revoke", webhooks.MessageTypeRevoke},
		{"edit", webhooks.MessageTypeEdit},
		// Case-insensitive
		{"TEXT", webhooks.MessageTypeText},
		{"Image", webhooks.MessageTypeImage},
		// Whitespace trimming
		{"  text  ", webhooks.MessageTypeText},
		// Unknown type
		{"nonexistent", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := webhooks.ParseMessageType(tt.input)
			if got != tt.expected {
				t.Errorf("ParseMessageType(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMessage_IsAReply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  *webhooks.Message
		want bool
	}{
		{
			name: "nil context",
			msg:  &webhooks.Message{Context: nil},
			want: false,
		},
		{
			name: "forwarded message",
			msg: &webhooks.Message{
				Context: &webhooks.Context{Forwarded: true},
			},
			want: false,
		},
		{
			name: "product inquiry",
			msg: &webhooks.Message{
				Context: &webhooks.Context{
					ReferredProduct: &webhooks.ReferredProduct{},
				},
			},
			want: false,
		},
		{
			name: "forwarded and product inquiry",
			msg: &webhooks.Message{
				Context: &webhooks.Context{
					Forwarded:       true,
					ReferredProduct: &webhooks.ReferredProduct{},
				},
			},
			want: false,
		},
		{
			name: "plain reply",
			msg: &webhooks.Message{
				Context: &webhooks.Context{ID: "wamid_xxx"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.msg.IsAReply(); got != tt.want {
				t.Errorf("IsAReply() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageNotificationContext_SenderInfo(t *testing.T) {
	t.Parallel()

	t.Run("empty contacts returns nil", func(t *testing.T) {
		t.Parallel()
		ctx := &webhooks.MessageNotificationContext{
			Contacts: nil,
		}
		if got := ctx.SenderInfo(); got != nil {
			t.Errorf("SenderInfo() = %v, want nil for empty contacts", got)
		}
	})

	t.Run("returns first contact", func(t *testing.T) {
		t.Parallel()
		ctx := &webhooks.MessageNotificationContext{
			Contacts: []*webhooks.Contact{
				{Profile: &webhooks.Profile{Name: "Alice"}, WaID: "111"},
				{Profile: &webhooks.Profile{Name: "Bob"}, WaID: "222"},
			},
		}
		got := ctx.SenderInfo()
		if got == nil {
			t.Fatal("SenderInfo() returned nil")
		}
		if got.Name != "Alice" || got.WaID != "111" {
			t.Errorf("SenderInfo() = {Name: %q, WaID: %q}, want {Alice, 111}", got.Name, got.WaID)
		}
	})
}

func TestInteractiveHandler_DedicatedSubTypes(t *testing.T) {
	t.Parallel()

	nctx := &webhooks.MessageNotificationContext{EntryID: "entry-1"}
	info := &webhooks.MessageInfo{Type: "interactive"}

	t.Run("list reply is handled", func(t *testing.T) {
		t.Parallel()

		ih := &webhooks.InteractiveHandler{}
		var called bool
		ih.OnListReply(
			webhooks.MessageHandlerFunc[webhooks.ListReply](
				func(_ context.Context, req *webhooks.MessageRequest[webhooks.ListReply]) error {
					called = true
					if req.Payload.ID != "option-1" {
						t.Errorf("ID = %q, want %q", req.Payload.ID, "option-1")
					}
					return nil
				},
			),
		)

		msg := &webhooks.Message{
			Type: "interactive",
			Interactive: &webhooks.Interactive{
				Type:      webhooks.InteractiveTypeListReply,
				ListReply: &webhooks.ListReply{ID: "option-1", Title: "Select Me", Description: "An option"},
			},
		}

		err := ih.Handle(context.Background(), nctx, info, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("ListReply handler was not called")
		}
	})

	t.Run("button reply is handled", func(t *testing.T) {
		t.Parallel()

		ih := &webhooks.InteractiveHandler{}
		var called bool
		ih.OnButtonReply(
			webhooks.MessageHandlerFunc[webhooks.ButtonReply](
				func(_ context.Context, req *webhooks.MessageRequest[webhooks.ButtonReply]) error {
					called = true
					return nil
				},
			),
		)

		msg := &webhooks.Message{
			Type: "interactive",
			Interactive: &webhooks.Interactive{
				Type:        webhooks.InteractiveTypeButtonReply,
				ButtonReply: &webhooks.ButtonReply{ID: "btn-1", Title: "Click"},
			},
		}

		err := ih.Handle(context.Background(), nctx, info, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("ButtonReply handler was not called")
		}
	})

	t.Run("flow completion is handled", func(t *testing.T) {
		t.Parallel()

		ih := &webhooks.InteractiveHandler{}
		var called bool
		ih.OnFlowCompletion(
			webhooks.MessageHandlerFunc[webhooks.NFMReply](
				func(_ context.Context, req *webhooks.MessageRequest[webhooks.NFMReply]) error {
					called = true
					return nil
				},
			),
		)

		msg := &webhooks.Message{
			Type: "interactive",
			Interactive: &webhooks.Interactive{
				Type:     webhooks.InteractiveTypeNFMReply,
				NFMReply: &webhooks.NFMReply{Name: "flow-1", Body: "done", ResponseJSON: []byte(`{}`)},
			},
		}

		err := ih.Handle(context.Background(), nctx, info, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("FlowCompletion handler was not called")
		}
	})

	t.Run("address submission is handled", func(t *testing.T) {
		t.Parallel()

		ih := &webhooks.InteractiveHandler{}
		var called bool
		ih.OnAddressSubmission(
			webhooks.MessageHandlerFunc[webhooks.NFMReply](
				func(_ context.Context, req *webhooks.MessageRequest[webhooks.NFMReply]) error {
					called = true
					return nil
				},
			),
		)

		msg := &webhooks.Message{
			Type: "interactive",
			Interactive: &webhooks.Interactive{
				Type:     webhooks.InteractiveAddressSubmission,
				NFMReply: &webhooks.NFMReply{Name: "addr-1", Body: "123 Main St", ResponseJSON: []byte(`{}`)},
			},
		}

		err := ih.Handle(context.Background(), nctx, info, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("AddressSubmission handler was not called")
		}
	})
}

func TestInteractiveHandler_Fallback(t *testing.T) {
	t.Parallel()

	nctx := &webhooks.MessageNotificationContext{EntryID: "entry-1"}
	info := &webhooks.MessageInfo{Type: "interactive"}

	t.Run("unrecognized sub-type falls through to Interactive handler", func(t *testing.T) {
		t.Parallel()

		ih := &webhooks.InteractiveHandler{}
		var called bool
		ih.OnFallback(
			webhooks.MessageHandlerFunc[webhooks.Interactive](
				func(_ context.Context, req *webhooks.MessageRequest[webhooks.Interactive]) error {
					called = true
					if req.Payload.Type != "some_new_type" {
						t.Errorf("Type = %q, want %q", req.Payload.Type, "some_new_type")
					}
					return nil
				},
			),
		)

		msg := &webhooks.Message{
			Type: "interactive",
			Interactive: &webhooks.Interactive{
				Type: "some_new_type",
			},
		}

		err := ih.Handle(context.Background(), nctx, info, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("Interactive fallback was not called")
		}
	})

	t.Run("nil Interactive payload is silently skipped", func(t *testing.T) {
		t.Parallel()

		ih := &webhooks.InteractiveHandler{}

		msg := &webhooks.Message{Type: "interactive"}

		err := ih.Handle(context.Background(), nctx, info, msg)
		if err != nil {
			t.Errorf("expected nil for nil Interactive, got: %v", err)
		}
	})

	t.Run("dedicated handler takes priority over fallback", func(t *testing.T) {
		t.Parallel()

		var dedicatedCalled, fallbackCalled bool

		ih := &webhooks.InteractiveHandler{}
		ih.OnButtonReply(
			webhooks.MessageHandlerFunc[webhooks.ButtonReply](
				func(_ context.Context, _ *webhooks.MessageRequest[webhooks.ButtonReply]) error {
					dedicatedCalled = true
					return nil
				},
			),
		)
		ih.OnFallback(
			webhooks.MessageHandlerFunc[webhooks.Interactive](
				func(_ context.Context, _ *webhooks.MessageRequest[webhooks.Interactive]) error {
					fallbackCalled = true
					return nil
				},
			),
		)

		msg := &webhooks.Message{
			Type: "interactive",
			Interactive: &webhooks.Interactive{
				Type:        webhooks.InteractiveTypeButtonReply,
				ButtonReply: &webhooks.ButtonReply{ID: "btn-1"},
			},
		}

		err := ih.Handle(context.Background(), nctx, info, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !dedicatedCalled {
			t.Error("dedicated handler was not called")
		}
		if fallbackCalled {
			t.Error("fallback should not be called when dedicated handler is set")
		}
	})
}

func TestInteractiveHandler_Setters(t *testing.T) {
	t.Parallel()

	t.Run("OnListReply sets ListReply", func(t *testing.T) {
		t.Parallel()
		ih := &webhooks.InteractiveHandler{}
		ih.OnListReply(
			webhooks.MessageHandlerFunc[webhooks.ListReply](
				func(_ context.Context, _ *webhooks.MessageRequest[webhooks.ListReply]) error {
					return nil
				},
			),
		)
		if ih.ListReply == nil {
			t.Error("ListReply should not be nil")
		}
	})

	t.Run("OnButtonReply sets ButtonReply", func(t *testing.T) {
		t.Parallel()
		ih := &webhooks.InteractiveHandler{}
		ih.OnButtonReply(
			webhooks.MessageHandlerFunc[webhooks.ButtonReply](
				func(_ context.Context, _ *webhooks.MessageRequest[webhooks.ButtonReply]) error {
					return nil
				},
			),
		)
		if ih.ButtonReply == nil {
			t.Error("ButtonReply should not be nil")
		}
	})

	t.Run("OnInteractive sets Interactive", func(t *testing.T) {
		t.Parallel()
		ih := &webhooks.InteractiveHandler{}
		ih.OnFallback(
			webhooks.MessageHandlerFunc[webhooks.Interactive](
				func(_ context.Context, _ *webhooks.MessageRequest[webhooks.Interactive]) error {
					return nil
				},
			),
		)
		if ih.Fallback == nil {
			t.Error("Interactive should not be nil")
		}
	})
}

// messageTextPayload returns a minimal messages notification containing a text message.
func messageTextPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "messages",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							Contacts: []*webhooks.Contact{
								{WaID: "15550001111", Profile: &webhooks.Profile{Name: "Test User"}},
							},
							Messages: []*webhooks.Message{
								{
									From:      "15550001111",
									ID:        "wamid.test123",
									Timestamp: "1739321024",
									Type:      "text",
									Text:      &webhooks.Text{Body: "Hello"},
								},
							},
						},
					},
				},
			},
		},
	}
}

func messageUnknownTypePayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "messages",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							Contacts: []*webhooks.Contact{
								{WaID: "15550001111", Profile: &webhooks.Profile{Name: "Test User"}},
							},
							Messages: []*webhooks.Message{
								{
									From:      "15550001111",
									ID:        "wamid.test456",
									Timestamp: "1739321024",
									Type:      "future_message_type_v99",
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestFallback_Messages_NoSubHandler_Silent200(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()
	resp := h.HandleNotification(context.Background(), messageTextPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (messages silently skipped when sub-handler nil), got %d", resp.StatusCode)
	}
}

func TestFallback_Messages_StatusChangeFiresEvenWhenMessagesNil(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	var statusFired bool
	h.OnMessageStatusChange(webhooks.ChangeValueHandlerFunc[webhooks.Status](
		func(_ context.Context, req *webhooks.ChangeValueRequest[webhooks.Status]) error {
			statusFired = true
			return nil
		},
	))

	notification := &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "messages",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							Statuses: []*webhooks.Status{
								{
									ID:          "wamid.status123",
									StatusValue: "sent",
									Timestamp:   "1739321024",
								},
							},
						},
					},
				},
			},
		},
	}

	resp := h.HandleNotification(context.Background(), notification)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !statusFired {
		t.Fatal("status handler was not invoked")
	}
}

func TestFallback_Messages_DedicatedHandlerFires(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	var fired bool
	h.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
		func(_ context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
			fired = true
			return nil
		},
	))

	resp := h.HandleNotification(context.Background(), messageTextPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("dedicated handler was not invoked")
	}
}

func TestFallback_Messages_SubFallbackFires(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	h.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
		func(_ context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
			t.Error("text handler should not fire for unrecognized type")
			return nil
		},
	))

	var subFired bool
	h.Messages().Fallback = webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			subFired = true
			return nil
		},
	)

	resp := h.HandleNotification(context.Background(), messageUnknownTypePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !subFired {
		t.Fatal("sub-handler fallback was not invoked for unhandled message type")
	}
}

func TestFallback_Messages_NoSubFallback_Silent200(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	h.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
		func(_ context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
			return nil
		},
	))

	resp := h.HandleNotification(context.Background(), messageUnknownTypePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

func TestFallback_Messages_OnFallbackPropagates(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	h.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
		func(_ context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
			return nil
		},
	))

	if h.Messages().Fallback != nil {
		t.Fatal("Messages.Fallback should be nil before OnFallback")
	}

	var fired bool
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			fired = true
			return nil
		},
	))

	if h.Messages().Fallback == nil {
		t.Fatal("OnFallback did not propagate to Messages.Fallback")
	}

	resp := h.HandleNotification(context.Background(), messageUnknownTypePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("propagated fallback was not invoked for unrecognized message type")
	}
}
