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

// businessAccountReviewPayload returns a minimal account_review_update
// notification — a different business field used for sub-fallback testing.
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
	// No business handlers → handler.business is nil.

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
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
			fired = true
			gotField = c.Field
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

	// Register alerts handler to init handler.business, then clear it.
	h.OnBusinessAlertNotification(webhooks.BusinessEventHandlerFunc[webhooks.AlertNotification](
		func(_ context.Context, req *webhooks.BusinessRequest[webhooks.AlertNotification]) error {
			t.Error("alerts handler should not fire for account review")
			return nil
		},
	))
	h.Business().Alerts = nil

	// Set sub-fallback for unhandled business fields.
	var subFired bool
	var gotField string
	h.Business().OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
			subFired = true
			gotField = c.Field
			return nil
		},
	))

	// Send a different business field (account_review_update).
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

	// Init business but don't set any fallback.
	h.OnBusinessAlertNotification(webhooks.BusinessEventHandlerFunc[webhooks.AlertNotification](
		func(_ context.Context, req *webhooks.BusinessRequest[webhooks.AlertNotification]) error {
			return nil
		},
	))
	// Business.Fallback is nil, AccountReview is nil — should silently 200
	// for account_review_update.

	resp := h.HandleNotification(context.Background(), businessAccountReviewPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

func TestFallback_Business_OnFallbackPropagates(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	// Init business.
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
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
			fired = true
			gotField = c.Field
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
