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

// flowsStatusPayload returns a flows notification for FLOW_STATUS_CHANGE.
func flowsStatusPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "flows",
						Value: &webhooks.Value{
							Event:     "FLOW_STATUS_CHANGE",
							FlowID:    "123456789",
							Message:   "Flow status changed to PUBLISHED",
							OldStatus: "DRAFT",
							NewStatus: "PUBLISHED",
						},
					},
				},
			},
		},
	}
}

// flowsEndpointLatencyPayload returns a flows notification for
// ENDPOINT_LATENCY — a different flow event for sub-fallback testing.
func flowsEndpointLatencyPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "flows",
						Value: &webhooks.Value{
							Event:         "ENDPOINT_LATENCY",
							FlowID:        "123456789",
							Message:       "P90 latency exceeded threshold",
							P50Latency:    1200,
							P90Latency:    5200,
							RequestsCount: 100,
							Threshold:     5000,
							AlertState:    "ACTIVATED",
						},
					},
				},
			},
		},
	}
}

func TestFallback_Flows_NoSubHandler_Silent200(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()
	// No flows handlers → handler.flows is nil.

	resp := h.HandleNotification(context.Background(), flowsStatusPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack for nil sub-handler), got %d", resp.StatusCode)
	}
}

func TestFallback_Flows_NoSubHandler_GeneralFallbackFires(t *testing.T) {
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

	resp := h.HandleNotification(context.Background(), flowsStatusPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("general fallback was not invoked for nil flows sub-handler")
	}
	if gotField != "flows" {
		t.Errorf("expected field 'flows', got %q", gotField)
	}
}

func TestFallback_Flows_DedicatedHandlerFires(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	var fired bool
	h.OnFlowStatusChange(webhooks.FlowEventHandlerFunc[webhooks.StatusChangeDetails](
		func(_ context.Context, req *webhooks.FlowRequest[webhooks.StatusChangeDetails]) error {
			fired = true
			return nil
		},
	))

	resp := h.HandleNotification(context.Background(), flowsStatusPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("dedicated handler was not invoked")
	}
}

func TestFallback_Flows_SubFallbackFires(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	// Register status handler to init handler.flows.
	h.OnFlowStatusChange(webhooks.FlowEventHandlerFunc[webhooks.StatusChangeDetails](
		func(_ context.Context, req *webhooks.FlowRequest[webhooks.StatusChangeDetails]) error {
			t.Error("status handler should not fire for latency event")
			return nil
		},
	))

	// Set sub-fallback. EndpointLatency is nil → should hit this.
	var subFired bool
	h.Flows().OnFallback(webhooks.FlowFallbackHandlerFunc(
		func(_ context.Context, nctx *webhooks.FlowNotificationContext, value *webhooks.Value) error {
			subFired = true
			return nil
		},
	))

	resp := h.HandleNotification(context.Background(), flowsEndpointLatencyPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !subFired {
		t.Fatal("sub-handler fallback was not invoked for unhandled flow event")
	}
}

func TestFallback_Flows_NoSubFallback_Silent200(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	h.OnFlowStatusChange(webhooks.FlowEventHandlerFunc[webhooks.StatusChangeDetails](
		func(_ context.Context, req *webhooks.FlowRequest[webhooks.StatusChangeDetails]) error {
			return nil
		},
	))
	// Flows.Fallback is nil, EndpointLatency is nil → silently 200.

	resp := h.HandleNotification(context.Background(), flowsEndpointLatencyPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

func TestFallback_Flows_OnFallbackPropagates(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	// Register status handler to init handler.flows.
	h.OnFlowStatusChange(webhooks.FlowEventHandlerFunc[webhooks.StatusChangeDetails](
		func(_ context.Context, req *webhooks.FlowRequest[webhooks.StatusChangeDetails]) error {
			return nil
		},
	))

	if h.Flows().Fallback != nil {
		t.Fatal("Flows.Fallback should be nil before OnFallback")
	}

	var fired bool
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
			fired = true
			return nil
		},
	))

	// Propagation should have set Flows.Fallback via the adapter.
	if h.Flows().Fallback == nil {
		t.Fatal("OnFallback did not propagate to Flows.Fallback")
	}

	// Send an unhandled flow event — Flows.Fallback (adapter) fires.
	resp := h.HandleNotification(context.Background(), flowsEndpointLatencyPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("propagated fallback was not invoked for unhandled flow event")
	}
}
