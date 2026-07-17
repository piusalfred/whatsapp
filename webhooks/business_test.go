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
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piusalfred/whatsapp/webhooks"
)

func TestBusinessNotificationHandler_DedicatedCall(t *testing.T) {
	t.Parallel()

	t.Run("alerts handler is called", func(t *testing.T) {
		t.Parallel()

		bh := &webhooks.BusinessNotificationHandler{}
		var called bool
		bh.OnAlerts(webhooks.BusinessEventHandlerFunc[webhooks.AlertNotification](
			func(_ context.Context, req *webhooks.BusinessRequest[webhooks.AlertNotification]) error {
				called = true
				if req.Payload.AlertSeverity != "HIGH" {
					t.Errorf("AlertSeverity = %q, want %q", req.Payload.AlertSeverity, "HIGH")
				}
				return nil
			},
		))

		event := webhooks.NotificationEvent{
			Object:  "whatsapp_business_account",
			EntryID: "entry-1",
			Time:    123456,
			Field:   webhooks.ChangeFieldAccountAlerts.String(),
			Value:   &webhooks.Value{AlertSeverity: "HIGH"},
		}

		err := bh.Handle(context.Background(), event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("alerts handler was not called")
		}
	})

	t.Run("nil handler is silently skipped", func(t *testing.T) {
		t.Parallel()

		bh := &webhooks.BusinessNotificationHandler{} // no handlers set

		event := webhooks.NotificationEvent{
			Object:  "whatsapp_business_account",
			EntryID: "entry-1",
			Time:    123456,
			Field:   webhooks.ChangeFieldSecurity.String(),
			Value:   &webhooks.Value{Event: "PIN_CHANGED"},
		}

		err := bh.Handle(context.Background(), event)
		if err != nil {
			t.Errorf("expected nil for unhandled change, got: %v", err)
		}
	})

	t.Run("handler error propagates", func(t *testing.T) {
		t.Parallel()

		bh := &webhooks.BusinessNotificationHandler{}
		bh.OnSecurity(webhooks.BusinessEventHandlerFunc[webhooks.SecurityNotification](
			func(_ context.Context, _ *webhooks.BusinessRequest[webhooks.SecurityNotification]) error {
				return errDummy
			},
		))

		event := webhooks.NotificationEvent{
			Object:  "whatsapp_business_account",
			EntryID: "entry-1",
			Time:    123456,
			Field:   webhooks.ChangeFieldSecurity.String(),
			Value:   &webhooks.Value{Event: "PIN_CHANGED"},
		}

		err := bh.Handle(context.Background(), event)
		if err == nil {
			t.Error("expected error from handler, got nil")
		}
	})
}

func TestBusinessNotificationHandler_Setters(t *testing.T) {
	t.Parallel()

	t.Run("OnAlerts sets Alerts", func(t *testing.T) {
		t.Parallel()
		bh := &webhooks.BusinessNotificationHandler{}
		bh.OnAlerts(webhooks.BusinessEventHandlerFunc[webhooks.AlertNotification](
			func(_ context.Context, _ *webhooks.BusinessRequest[webhooks.AlertNotification]) error {
				return nil
			},
		))
		if bh.Alerts == nil {
			t.Error("Alerts should not be nil after OnAlerts")
		}
	})

	t.Run("OnAccount sets Account", func(t *testing.T) {
		t.Parallel()
		bh := &webhooks.BusinessNotificationHandler{}
		bh.OnAccount(webhooks.BusinessEventHandlerFunc[webhooks.AccountUpdate](
			func(_ context.Context, _ *webhooks.BusinessRequest[webhooks.AccountUpdate]) error {
				return nil
			},
		))
		if bh.Account == nil {
			t.Error("Account should not be nil after OnAccount")
		}
	})

	t.Run("OnSecurity sets Security", func(t *testing.T) {
		t.Parallel()
		bh := &webhooks.BusinessNotificationHandler{}
		bh.OnSecurity(webhooks.BusinessEventHandlerFunc[webhooks.SecurityNotification](
			func(_ context.Context, _ *webhooks.BusinessRequest[webhooks.SecurityNotification]) error {
				return nil
			},
		))
		if bh.Security == nil {
			t.Error("Security should not be nil after OnSecurity")
		}
	})
}

// TestDefect001_TemplateStatusUpdate_NilReason verifies that a template status
// update webhook without a "reason" field does not cause a nil pointer dereference.
func TestDefect001_TemplateStatusUpdate_NilReason(t *testing.T) {
	t.Parallel()

	payload := []byte(`{
			"object": "whatsapp_business_account",
			"entry": [{
				"id": "123456789",
				"time": 1719000000,
				"changes": [{
					"field": "message_template_status_update",
					"value": {
						"event": "APPROVED",
						"message_template_id": 12345,
						"message_template_name": "hello_world",
						"message_template_language": "en"
					}
				}]
			}]
		}`)

	r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	r.Header.Set("Content-Type", "application/json")

	notification, err := webhooks.ParseNotification(r, &webhooks.ParseNotificationOptions{
		VerifyPayloadSignature: false,
	})
	if err != nil {
		t.Fatalf("ExtractAndValidatePayload failed: %v", err)
	}
	if notification == nil {
		t.Fatal("notification is nil")
	}

	handler := webhooks.NewHandler()
	handler.OnBusinessTemplateStatusUpdate(
		webhooks.BusinessEventHandlerFunc[webhooks.TemplateStatusUpdateNotification](
			func(ctx context.Context, req *webhooks.BusinessRequest[webhooks.TemplateStatusUpdateNotification]) error {
				return nil
			},
		),
	)

	resp := handler.HandleNotification(context.Background(), notification)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("DEFECT CONFIRMED: got status %d, want 200. "+
			"TemplateStatusUpdate() at business.go:542 dereferences *value.Reason "+
			"without nil check, causing panic->500 for valid template status updates "+
			"without a reason field.", resp.StatusCode)
	}
}

// businessAlertsPayload returns a minimal account_alerts notification.
func businessAlertsPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "account_alerts",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							AlertSeverity:    "WARNING",
							AlertStatus:      "ACTIVE",
							AlertType:        "MESSAGING_LIMIT",
							AlertDescription: "80% of messaging limit reached",
							EntityType:       "PHONE_NUMBER",
							EntityID:         "15550783881",
						},
					},
				},
			},
		},
	}
}

func businessAccountReviewPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "account_review_update",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							Decision: "APPROVED",
						},
					},
				},
			},
		},
	}
}

func TestFallback_Business_NoSubHandler_Silent200(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()
	resp := h.HandleNotification(context.Background(), businessAlertsPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack for nil sub-handler), got %d", resp.StatusCode)
	}
}

func TestFallback_Business_NoSubHandler_GeneralFallbackFires(t *testing.T) {
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

	resp := h.HandleNotification(context.Background(), businessAlertsPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("general fallback was not invoked for nil business sub-handler")
	}
	if gotField != "account_alerts" {
		t.Errorf("expected field 'account_alerts', got %q", gotField)
	}
}

func TestFallback_Business_DedicatedHandlerFires(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	var fired bool
	h.OnBusinessAlertNotification(webhooks.BusinessEventHandlerFunc[webhooks.AlertNotification](
		func(_ context.Context, req *webhooks.BusinessRequest[webhooks.AlertNotification]) error {
			fired = true
			return nil
		},
	))

	resp := h.HandleNotification(context.Background(), businessAlertsPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("dedicated handler was not invoked")
	}
}

func TestFallback_Business_SubFallbackFires(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	h.OnBusinessAlertNotification(webhooks.BusinessEventHandlerFunc[webhooks.AlertNotification](
		func(_ context.Context, req *webhooks.BusinessRequest[webhooks.AlertNotification]) error {
			t.Error("alerts handler should not fire for account review")
			return nil
		},
	))
	h.Business().Alerts = nil

	var subFired bool
	var gotField string
	h.Business().OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			subFired = true
			gotField = ev.Field
			return nil
		},
	))

	resp := h.HandleNotification(context.Background(), businessAccountReviewPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !subFired {
		t.Fatal("sub-handler fallback was not invoked for unhandled business field")
	}
	if gotField != "account_review_update" {
		t.Errorf("expected field 'account_review_update', got %q", gotField)
	}
}

func TestFallback_Business_NoSubFallback_Silent200(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	h.OnBusinessAlertNotification(webhooks.BusinessEventHandlerFunc[webhooks.AlertNotification](
		func(_ context.Context, req *webhooks.BusinessRequest[webhooks.AlertNotification]) error {
			return nil
		},
	))

	resp := h.HandleNotification(context.Background(), businessAccountReviewPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

func TestFallback_Business_OnFallbackPropagates(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	h.OnBusinessAlertNotification(webhooks.BusinessEventHandlerFunc[webhooks.AlertNotification](
		func(_ context.Context, req *webhooks.BusinessRequest[webhooks.AlertNotification]) error {
			return nil
		},
	))

	if h.Business().Fallback != nil {
		t.Fatal("Business.Fallback should be nil before OnFallback")
	}

	var fired bool
	var gotField string
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			fired = true
			gotField = ev.Field
			return nil
		},
	))

	if h.Business().Fallback == nil {
		t.Fatal("OnFallback did not propagate to Business.Fallback")
	}

	resp := h.HandleNotification(context.Background(), businessAccountReviewPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("fallback was not invoked for unhandled business field")
	}
	if gotField != "account_review_update" {
		t.Errorf("expected field 'account_review_update', got %q", gotField)
	}
}
