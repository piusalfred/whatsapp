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

// groupStatusUpdatePayload returns a minimal group_status_update notification
// suitable for testing the fallback cascade.
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
								{
									GroupID: "GROUP_ID_123",
									Type:    "group_suspend",
								},
							},
						},
					},
				},
			},
		},
	}
}

// TestFallback_NoSubHandler_Silent200 verifies that when a group field arrives
// and handler.groups is nil (no group handlers registered), the dispatch nil-
// guard catches it and silently returns 200.
func TestFallback_NoSubHandler_Silent200(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()
	// No group handlers registered → handler.groups is nil.

	resp := h.HandleNotification(context.Background(), groupStatusUpdatePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack for nil sub-handler), got %d", resp.StatusCode)
	}
}

// TestFallback_NoSubHandler_GeneralFallbackFires verifies that when
// handler.groups is nil, the general (handler-level) fallback catches the
// group field.
func TestFallback_NoSubHandler_GeneralFallbackFires(t *testing.T) {
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

// TestFallback_DedicatedHandlerFires verifies the happy path: registering a
// dedicated handler for group_status_update and confirming it fires.
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

// TestFallback_SubHandlerFallbackFires verifies the two-level cascade: a group
// field arrives for which no dedicated handler exists, but the GroupsHandler
// itself has a Fallback set. The sub-handler fallback should catch it.
func TestFallback_SubHandlerFallbackFires(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	// Register a lifecycle handler to lazily init handler.groups.
	h.OnGroupLifecycleUpdate(webhooks.ChangeValueHandlerFunc[webhooks.Group](
		func(_ context.Context, req *webhooks.ChangeValueRequest[webhooks.Group]) error {
			t.Error("lifecycle handler should not fire for status update")
			return nil
		},
	))

	// Now handler.groups exists, but StatusUpdate is nil.
	// Set the sub-handler's Fallback explicitly.
	var subFallbackFired bool
	h.Groups().OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
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

// TestFallback_NoSubFallback_Silent200 verifies that when the GroupsHandler
// exists but neither a dedicated handler nor a sub-handler Fallback are set,
// the field is silently acknowledged (200).
func TestFallback_NoSubFallback_Silent200(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	// Init groups but don't set any fallback.
	h.OnGroupLifecycleUpdate(webhooks.ChangeValueHandlerFunc[webhooks.Group](
		func(_ context.Context, req *webhooks.ChangeValueRequest[webhooks.Group]) error {
			t.Error("lifecycle handler should not fire")
			return nil
		},
	))
	// groups.Fallback is still nil.

	resp := h.HandleNotification(context.Background(), groupStatusUpdatePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

// TestFallback_OnFallbackPropagates verifies that calling Handler.OnFallback
// propagates the fallback to sub-handlers that don't already have one set,
// and that unhandled fields within a domain are caught by the fallback
// (whether at the sub-handler level or the general level).
func TestFallback_OnFallbackPropagates(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	// Init groups but don't set a dedicated handler for status updates.
	h.OnGroupLifecycleUpdate(webhooks.ChangeValueHandlerFunc[webhooks.Group](
		func(_ context.Context, req *webhooks.ChangeValueRequest[webhooks.Group]) error {
			return nil
		},
	))

	// Verify groups.Fallback is nil before propagation.
	if h.Groups().Fallback != nil {
		t.Fatal("groups.Fallback should be nil before OnFallback")
	}

	var fallbackFired bool
	var fallbackField string
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
			fallbackFired = true
			fallbackField = c.Field
			return nil
		},
	))

	// Verify propagation happened — groups.Fallback should now be non-nil.
	if h.Groups().Fallback == nil {
		t.Fatal("OnFallback did not propagate to groups.Fallback")
	}

	// Send an unhandled group field. The GroupsHandler exists but
	// StatusUpdate is nil → the propagated sub-fallback catches it.
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
