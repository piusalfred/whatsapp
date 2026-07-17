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
	"errors"
	"net/http"
	"testing"

	"github.com/piusalfred/whatsapp/webhooks"
)

var errDummy = errors.New("dummy error")

func TestFlowNotificationHandler_DedicatedHandlers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		event     string
		value     *webhooks.Value
		setupFn   func(fh *webhooks.FlowNotificationHandler) (called *bool)
		wantError bool
	}{
		{
			name:  "FLOW_STATUS_CHANGE calls Status handler",
			event: webhooks.EventFlowStatusChange,
			value: &webhooks.Value{
				Event:     webhooks.EventFlowStatusChange,
				OldStatus: "DRAFT",
				NewStatus: "PUBLISHED",
			},
			setupFn: func(fh *webhooks.FlowNotificationHandler) *bool {
				var called bool
				fh.OnFlowStatusChange(webhooks.FlowEventHandlerFunc[webhooks.StatusChangeDetails](
					func(_ context.Context, req *webhooks.FlowRequest[webhooks.StatusChangeDetails]) error {
						called = true
						if req.Context == nil || req.Context.FlowID != "" {
							t.Logf("context FlowID: %v", req.Context)
						}
						if req.Payload.OldStatus != "DRAFT" || req.Payload.NewStatus != "PUBLISHED" {
							t.Errorf("unexpected details: old=%s new=%s", req.Payload.OldStatus, req.Payload.NewStatus)
						}
						return nil
					},
				))
				return &called
			},
		},
		{
			name:  "CLIENT_ERROR_RATE calls ClientErrorRate handler",
			event: webhooks.EventClientErrorRate,
			value: &webhooks.Value{
				Event:      webhooks.EventClientErrorRate,
				ErrorRate:  12.5,
				Threshold:  10,
				AlertState: "ACTIVATED",
			},
			setupFn: func(fh *webhooks.FlowNotificationHandler) *bool {
				var called bool
				fh.OnFlowClientErrorRate(webhooks.FlowEventHandlerFunc[webhooks.ClientErrorRateDetails](
					func(_ context.Context, req *webhooks.FlowRequest[webhooks.ClientErrorRateDetails]) error {
						called = true
						if req.Payload.ErrorRate != 12.5 || req.Payload.AlertState != "ACTIVATED" {
							t.Errorf(
								"unexpected details: rate=%f state=%s",
								req.Payload.ErrorRate,
								req.Payload.AlertState,
							)
						}
						return nil
					},
				))
				return &called
			},
		},
		{
			name:  "ENDPOINT_ERROR_RATE calls EndpointErrorRate handler",
			event: webhooks.EventEndpointErrorRate,
			value: &webhooks.Value{
				Event:      webhooks.EventEndpointErrorRate,
				ErrorRate:  5.0,
				Threshold:  3,
				AlertState: "DEACTIVATED",
			},
			setupFn: func(fh *webhooks.FlowNotificationHandler) *bool {
				var called bool
				fh.OnFlowEndpointErrorRate(webhooks.FlowEventHandlerFunc[webhooks.EndpointErrorRateDetails](
					func(_ context.Context, req *webhooks.FlowRequest[webhooks.EndpointErrorRateDetails]) error {
						called = true
						return nil
					},
				))
				return &called
			},
		},
		{
			name:  "ENDPOINT_LATENCY calls EndpointLatency handler",
			event: webhooks.EventEndpointLatency,
			value: &webhooks.Value{
				Event:         webhooks.EventEndpointLatency,
				P50Latency:    200,
				P90Latency:    500,
				RequestsCount: 1000,
				Threshold:     300,
				AlertState:    "ACTIVATED",
			},
			setupFn: func(fh *webhooks.FlowNotificationHandler) *bool {
				var called bool
				fh.OnFlowEndpointLatency(webhooks.FlowEventHandlerFunc[webhooks.EndpointLatencyDetails](
					func(_ context.Context, req *webhooks.FlowRequest[webhooks.EndpointLatencyDetails]) error {
						called = true
						return nil
					},
				))
				return &called
			},
		},
		{
			name:  "ENDPOINT_AVAILABILITY calls EndpointAvailability handler",
			event: webhooks.EventEndpointAvailability,
			value: &webhooks.Value{
				Event:        webhooks.EventEndpointAvailability,
				Availability: 95,
				Threshold:    99,
				AlertState:   "ACTIVATED",
			},
			setupFn: func(fh *webhooks.FlowNotificationHandler) *bool {
				var called bool
				fh.OnFlowEndpointAvailability(webhooks.FlowEventHandlerFunc[webhooks.EndpointAvailabilityDetails](
					func(_ context.Context, req *webhooks.FlowRequest[webhooks.EndpointAvailabilityDetails]) error {
						called = true
						return nil
					},
				))
				return &called
			},
		},
		{
			name:  "dedicated handler propagates error",
			event: webhooks.EventFlowStatusChange,
			value: &webhooks.Value{
				Event:     webhooks.EventFlowStatusChange,
				OldStatus: "DRAFT",
				NewStatus: "PUBLISHED",
			},
			setupFn: func(fh *webhooks.FlowNotificationHandler) *bool {
				var called bool
				fh.OnFlowStatusChange(webhooks.FlowEventHandlerFunc[webhooks.StatusChangeDetails](
					func(_ context.Context, _ *webhooks.FlowRequest[webhooks.StatusChangeDetails]) error {
						called = true
						return errDummy
					},
				))
				return &called
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fh := &webhooks.FlowNotificationHandler{}
			called := tt.setupFn(fh)

			event := webhooks.NotificationEvent{
				Object:  "whatsapp_business_account",
				EntryID: "entry-1",
				Time:    123456,
				Field:   webhooks.ChangeFieldFlows.String(),
				Value:   tt.value,
			}
			err := fh.Handle(context.Background(), event)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !*called {
				t.Error("dedicated handler was not called")
			}
		})
	}
}

func TestFlowNotificationHandler_Fallback(t *testing.T) {
	t.Parallel()

	t.Run("known event with nil dedicated handler calls FallbackHandler", func(t *testing.T) {
		t.Parallel()

		var fallbackCalled bool

		fh := &webhooks.FlowNotificationHandler{}
		fh.OnFallback(
			webhooks.FallbackHandlerFunc(
				func(_ context.Context, _ webhooks.NotificationEvent) error {
					fallbackCalled = true
					return nil
				},
			),
		)

		event := webhooks.NotificationEvent{
			Object:  "whatsapp_business_account",
			EntryID: "entry-1",
			Time:    123456,
			Field:   webhooks.ChangeFieldFlows.String(),
			Value: &webhooks.Value{
				Event:     webhooks.EventFlowStatusChange,
				OldStatus: "DRAFT",
				NewStatus: "PUBLISHED",
			},
		}
		err := fh.Handle(context.Background(), event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !fallbackCalled {
			t.Error("FallbackHandler was not called for known event without dedicated handler")
		}
	})

	t.Run("unknown event type calls FallbackHandler", func(t *testing.T) {
		t.Parallel()

		var fallbackCalled bool

		fh := &webhooks.FlowNotificationHandler{}
		fh.OnFallback(
			webhooks.FallbackHandlerFunc(
				func(_ context.Context, ev webhooks.NotificationEvent) error {
					fallbackCalled = true
					if ev.Value == nil || ev.Value.Event != "NEW_FLOW_EVENT" {
						t.Errorf("fallback value.Event = %q, want %q", ev.Value.Event, "NEW_FLOW_EVENT")
					}
					return nil
				},
			),
		)

		event := webhooks.NotificationEvent{
			Object:  "whatsapp_business_account",
			EntryID: "entry-1",
			Time:    123456,
			Field:   webhooks.ChangeFieldFlows.String(),
			Value: &webhooks.Value{
				Event:   "NEW_FLOW_EVENT",
				Message: "A new flow event",
			},
		}
		err := fh.Handle(context.Background(), event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !fallbackCalled {
			t.Error("FallbackHandler was not called for unknown event type")
		}
	})

	t.Run("nil FallbackHandler returns nil silently", func(t *testing.T) {
		t.Parallel()

		fh := &webhooks.FlowNotificationHandler{} // no handlers, no fallback

		event := webhooks.NotificationEvent{
			Object:  "whatsapp_business_account",
			EntryID: "entry-1",
			Time:    123456,
			Field:   webhooks.ChangeFieldFlows.String(),
			Value: &webhooks.Value{
				Event: webhooks.EventFlowStatusChange,
			},
		}
		err := fh.Handle(context.Background(), event)
		if err != nil {
			t.Errorf("expected nil for unhandled event, got: %v", err)
		}
	})

	t.Run("nil FallbackHandler for unknown event returns nil silently", func(t *testing.T) {
		t.Parallel()

		fh := &webhooks.FlowNotificationHandler{}

		event := webhooks.NotificationEvent{
			Object:  "whatsapp_business_account",
			EntryID: "entry-1",
			Time:    123456,
			Field:   webhooks.ChangeFieldFlows.String(),
			Value: &webhooks.Value{
				Event: "SOME_RANDOM_EVENT",
			},
		}
		err := fh.Handle(context.Background(), event)
		if err != nil {
			t.Errorf("expected nil for unknown event, got: %v", err)
		}
	})
}

func TestFlowNotificationHandler_Setters(t *testing.T) {
	t.Parallel()

	t.Run("OnFlowStatusChange sets Status", func(t *testing.T) {
		t.Parallel()
		fh := &webhooks.FlowNotificationHandler{}
		fh.OnFlowStatusChange(webhooks.FlowEventHandlerFunc[webhooks.StatusChangeDetails](
			func(_ context.Context, _ *webhooks.FlowRequest[webhooks.StatusChangeDetails]) error {
				return nil
			},
		))
		if fh.Status == nil {
			t.Error("Status should not be nil after OnFlowStatusChange")
		}
	})

	t.Run("OnFlowClientErrorRate sets ClientErrorRate", func(t *testing.T) {
		t.Parallel()
		fh := &webhooks.FlowNotificationHandler{}
		fh.OnFlowClientErrorRate(webhooks.FlowEventHandlerFunc[webhooks.ClientErrorRateDetails](
			func(_ context.Context, _ *webhooks.FlowRequest[webhooks.ClientErrorRateDetails]) error {
				return nil
			},
		))
		if fh.ClientErrorRate == nil {
			t.Error("ClientErrorRate should not be nil after OnFlowClientErrorRate")
		}
	})

	t.Run("OnFlowEndpointErrorRate sets EndpointErrorRate", func(t *testing.T) {
		t.Parallel()
		fh := &webhooks.FlowNotificationHandler{}
		fh.OnFlowEndpointErrorRate(webhooks.FlowEventHandlerFunc[webhooks.EndpointErrorRateDetails](
			func(_ context.Context, _ *webhooks.FlowRequest[webhooks.EndpointErrorRateDetails]) error {
				return nil
			},
		))
		if fh.EndpointErrorRate == nil {
			t.Error("EndpointErrorRate should not be nil after OnFlowEndpointErrorRate")
		}
	})

	t.Run("OnFlowEndpointLatency sets EndpointLatency", func(t *testing.T) {
		t.Parallel()
		fh := &webhooks.FlowNotificationHandler{}
		fh.OnFlowEndpointLatency(webhooks.FlowEventHandlerFunc[webhooks.EndpointLatencyDetails](
			func(_ context.Context, _ *webhooks.FlowRequest[webhooks.EndpointLatencyDetails]) error {
				return nil
			},
		))
		if fh.EndpointLatency == nil {
			t.Error("EndpointLatency should not be nil after OnFlowEndpointLatency")
		}
	})

	t.Run("OnFallback sets FallbackHandler", func(t *testing.T) {
		t.Parallel()
		fh := &webhooks.FlowNotificationHandler{}
		fh.OnFallback(
			webhooks.FallbackHandlerFunc(
				func(_ context.Context, _ webhooks.NotificationEvent) error {
					return nil
				},
			),
		)
		if fh.Fallback == nil {
			t.Error("Fallback should not be nil after OnFallback")
		}
	})

	t.Run("OnFlowEndpointAvailability sets EndpointAvailability", func(t *testing.T) {
		t.Parallel()
		fh := &webhooks.FlowNotificationHandler{}
		fh.OnFlowEndpointAvailability(webhooks.FlowEventHandlerFunc[webhooks.EndpointAvailabilityDetails](
			func(_ context.Context, _ *webhooks.FlowRequest[webhooks.EndpointAvailabilityDetails]) error {
				return nil
			},
		))
		if fh.EndpointAvailability == nil {
			t.Error("EndpointAvailability should not be nil after OnFlowEndpointAvailability")
		}
	})
}

func TestFlowNotificationHandler_DedicatedOverridesFallback(t *testing.T) {
	t.Parallel()

	fh := &webhooks.FlowNotificationHandler{}

	var dedicatedCalled, fallbackCalled bool

	fh.OnFlowStatusChange(webhooks.FlowEventHandlerFunc[webhooks.StatusChangeDetails](
		func(_ context.Context, _ *webhooks.FlowRequest[webhooks.StatusChangeDetails]) error {
			dedicatedCalled = true
			return nil
		},
	))
	fh.OnFallback(
		webhooks.FallbackHandlerFunc(
			func(_ context.Context, _ webhooks.NotificationEvent) error {
				fallbackCalled = true
				return nil
			},
		),
	)

	event := webhooks.NotificationEvent{
		Object:  "whatsapp_business_account",
		EntryID: "entry-1",
		Time:    123456,
		Field:   webhooks.ChangeFieldFlows.String(),
		Value: &webhooks.Value{
			Event:     webhooks.EventFlowStatusChange,
			OldStatus: "DRAFT",
			NewStatus: "PUBLISHED",
		},
	}
	err := fh.Handle(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dedicatedCalled {
		t.Error("dedicated handler was not called")
	}
	if fallbackCalled {
		t.Error("FallbackHandler should not be called when dedicated handler is set")
	}
}

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
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			fired = true
			gotField = ev.Field
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

	h.OnFlowStatusChange(webhooks.FlowEventHandlerFunc[webhooks.StatusChangeDetails](
		func(_ context.Context, req *webhooks.FlowRequest[webhooks.StatusChangeDetails]) error {
			t.Error("status handler should not fire for latency event")
			return nil
		},
	))

	var subFired bool
	h.Flows().OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ev webhooks.NotificationEvent) error {
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

	resp := h.HandleNotification(context.Background(), flowsEndpointLatencyPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

func TestFallback_Flows_OnFallbackPropagates(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

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
		func(_ context.Context, ev webhooks.NotificationEvent) error {
			fired = true
			return nil
		},
	))

	if h.Flows().Fallback == nil {
		t.Fatal("OnFallback did not propagate to Flows.Fallback")
	}

	resp := h.HandleNotification(context.Background(), flowsEndpointLatencyPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("propagated fallback was not invoked for unhandled flow event")
	}
}
