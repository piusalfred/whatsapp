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

// smbAppStateSyncPayload returns an smb_app_state_sync notification with
// a contact add action.
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
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
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
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
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
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
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
