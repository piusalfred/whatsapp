// Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software
// and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial
// portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
// LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package webhooks_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/piusalfred/whatsapp/webhooks"
)

// messageTextPayload returns a minimal messages notification containing a
// text message.
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

// messageUnknownTypePayload returns a messages notification with a type
// that has no dedicated handler and is not in the parseMessageTypeMap.
// This hits the MessagesHandler.Fallback (layer 3).
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
	// No messages handlers → handler.messages is nil.
	// Unlike Groups/Business/History, the Messages path in the dispatch
	// goes through handleNotificationMessageItem, which also processes
	// statuses and errors. Nil handler.messages only skips the messages
	// loop — status/error handlers still fire if registered.

	resp := h.HandleNotification(context.Background(), messageTextPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (messages silently skipped when sub-handler nil), got %d", resp.StatusCode)
	}
}

func TestFallback_Messages_StatusChangeFiresEvenWhenMessagesNil(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()
	// handler.messages is nil, but messageStatusChange is set directly on Handler.
	// A payload with statuses (not messages) should still be processed.

	var statusFired bool
	h.OnMessageStatusChange(webhooks.ChangeValueHandlerFunc[webhooks.Status](
		func(_ context.Context, req *webhooks.ChangeValueRequest[webhooks.Status]) error {
			statusFired = true
			return nil
		},
	))

	// Send a payload with statuses only, no messages.
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
		t.Fatal(
			"status handler was not invoked — handleNotificationMessageItem should process statuses even when handler.messages is nil",
		)
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

	// Register text handler to init handler.messages.
	h.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
		func(_ context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
			t.Error("text handler should not fire for unrecognized type")
			return nil
		},
	))

	// MessagesHandler.Fallback catches unrecognized message types that don't
	// match any case in the dispatch switch and don't have a probed payload
	// field (Contacts, Location, Identity).
	var subFired bool
	h.Messages().Fallback = webhooks.FallbackHandlerFunc(
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
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
	// MessagesHandler.Fallback is nil → unrecognized type silently 200.

	resp := h.HandleNotification(context.Background(), messageUnknownTypePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

func TestFallback_Messages_OnFallbackPropagates(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	// Register text handler to init handler.messages.
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
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
			fired = true
			return nil
		},
	))

	// Propagation should have set Messages.Fallback directly (same type now).
	if h.Messages().Fallback == nil {
		t.Fatal("OnFallback did not propagate to Messages.Fallback")
	}

	// Send unrecognized type — MessagesHandler.Fallback (adapter) fires.
	resp := h.HandleNotification(context.Background(), messageUnknownTypePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("propagated fallback was not invoked for unrecognized message type")
	}
}
