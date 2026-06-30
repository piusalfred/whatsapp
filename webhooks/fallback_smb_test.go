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

// smbTextEchoPayload returns an smb_message_echoes notification with a
// text message sent by the business customer via the WhatsApp Business app.
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

// smbRevokeEchoPayload returns an smb_message_echoes notification with a
// revoke (delete) message — a different type for sub-fallback testing.
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
	// No SMB echo handlers → handler.smbEcho is nil.

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
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
			fired = true
			gotField = c.Field
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

	// SMBMessageEchoesHandler has a single Handler (not per-type). When it's
	// nil, all echoes — regardless of type — fall through to Fallback.
	// Set Fallback directly, leave Handler nil.
	var subFired bool
	h.SMBEchoes().Fallback = webhooks.FallbackHandlerFunc(
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
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
	// Handler is nil, Fallback is nil → silently 200.

	resp := h.HandleNotification(context.Background(), smbRevokeEchoPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

func TestFallback_SMB_OnFallbackPropagates(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	// Register handler to init handler.smbEcho.
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
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
			fired = true
			return nil
		},
	))

	// Propagation should have set SMBEchoes.Fallback.
	if h.SMBEchoes().Fallback == nil {
		t.Fatal("OnFallback did not propagate to SMBEchoes.Fallback")
	}

	// Clear Handler so Fallback is exercised.
	h.SMBEchoes().Handler = nil

	// Send an echo — Fallback fires.
	resp := h.HandleNotification(context.Background(), smbTextEchoPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("propagated fallback was not invoked")
	}
}
