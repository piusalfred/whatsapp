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
	"testing"

	"github.com/piusalfred/whatsapp/webhooks"
)

var errDummy = errors.New("dummy error")

func TestFlowNotificationHandler_DedicatedHandlers(t *testing.T) {
	t.Parallel()

	nctx := &webhooks.FlowNotificationContext{FlowID: "flow-1"}

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
				fh.OnFlowStatusChange(
					func(_ context.Context, _ *webhooks.FlowNotificationContext, details *webhooks.StatusChangeDetails) error {
						called = true
						if details.OldStatus != "DRAFT" || details.NewStatus != "PUBLISHED" {
							t.Errorf("unexpected details: old=%s new=%s", details.OldStatus, details.NewStatus)
						}
						return nil
					},
				)
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
				fh.OnFlowClientErrorRate(
					func(_ context.Context, _ *webhooks.FlowNotificationContext, details *webhooks.ClientErrorRateDetails) error {
						called = true
						if details.ErrorRate != 12.5 || details.AlertState != "ACTIVATED" {
							t.Errorf("unexpected details: rate=%f state=%s", details.ErrorRate, details.AlertState)
						}
						return nil
					},
				)
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
				fh.OnFlowEndpointErrorRate(
					func(_ context.Context, _ *webhooks.FlowNotificationContext, details *webhooks.EndpointErrorRateDetails) error {
						called = true
						return nil
					},
				)
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
				fh.OnFlowEndpointLatency(
					func(_ context.Context, _ *webhooks.FlowNotificationContext, details *webhooks.EndpointLatencyDetails) error {
						called = true
						return nil
					},
				)
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
				fh.OnFlowEndpointAvailability(
					func(_ context.Context, _ *webhooks.FlowNotificationContext, details *webhooks.EndpointAvailabilityDetails) error {
						called = true
						return nil
					},
				)
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
				fh.OnFlowStatusChange(
					func(_ context.Context, _ *webhooks.FlowNotificationContext, _ *webhooks.StatusChangeDetails) error {
						called = true
						return errDummy
					},
				)
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

			err := fh.Handle(context.Background(), nctx, tt.value)

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

	nctx := &webhooks.FlowNotificationContext{FlowID: "flow-1"}

	t.Run("known event with nil dedicated handler calls FallbackHandler", func(t *testing.T) {
		t.Parallel()

		var fallbackCalled bool

		fh := &webhooks.FlowNotificationHandler{}
		fh.OnFallback(
			func(_ context.Context, _ *webhooks.FlowNotificationContext, _ *webhooks.Value) error {
				fallbackCalled = true
				return nil
			},
		)

		err := fh.Handle(context.Background(), nctx, &webhooks.Value{
			Event:     webhooks.EventFlowStatusChange,
			OldStatus: "DRAFT",
			NewStatus: "PUBLISHED",
		})
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
			func(_ context.Context, _ *webhooks.FlowNotificationContext, value *webhooks.Value) error {
				fallbackCalled = true
				if value == nil || value.Event != "NEW_FLOW_EVENT" {
					t.Errorf("fallback value.Event = %q, want %q", value.Event, "NEW_FLOW_EVENT")
				}
				return nil
			},
		)

		err := fh.Handle(context.Background(), nctx, &webhooks.Value{
			Event:   "NEW_FLOW_EVENT",
			Message: "A new flow event",
		})
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

		err := fh.Handle(context.Background(), nctx, &webhooks.Value{
			Event: webhooks.EventFlowStatusChange,
		})
		if err != nil {
			t.Errorf("expected nil for unhandled event, got: %v", err)
		}
	})

	t.Run("nil FallbackHandler for unknown event returns nil silently", func(t *testing.T) {
		t.Parallel()

		fh := &webhooks.FlowNotificationHandler{}

		err := fh.Handle(context.Background(), nctx, &webhooks.Value{
			Event: "SOME_RANDOM_EVENT",
		})
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
		fh.OnFlowStatusChange(
			func(_ context.Context, _ *webhooks.FlowNotificationContext, _ *webhooks.StatusChangeDetails) error {
				return nil
			},
		)
		if fh.Status == nil {
			t.Error("Status should not be nil after OnFlowStatusChange")
		}
	})

	t.Run("OnFlowClientErrorRate sets ClientErrorRate", func(t *testing.T) {
		t.Parallel()
		fh := &webhooks.FlowNotificationHandler{}
		fh.OnFlowClientErrorRate(
			func(_ context.Context, _ *webhooks.FlowNotificationContext, _ *webhooks.ClientErrorRateDetails) error {
				return nil
			},
		)
		if fh.ClientErrorRate == nil {
			t.Error("ClientErrorRate should not be nil after OnFlowClientErrorRate")
		}
	})

	t.Run("OnFlowEndpointErrorRate sets EndpointErrorRate", func(t *testing.T) {
		t.Parallel()
		fh := &webhooks.FlowNotificationHandler{}
		fh.OnFlowEndpointErrorRate(
			func(_ context.Context, _ *webhooks.FlowNotificationContext, _ *webhooks.EndpointErrorRateDetails) error {
				return nil
			},
		)
		if fh.EndpointErrorRate == nil {
			t.Error("EndpointErrorRate should not be nil after OnFlowEndpointErrorRate")
		}
	})

	t.Run("OnFlowEndpointLatency sets EndpointLatency", func(t *testing.T) {
		t.Parallel()
		fh := &webhooks.FlowNotificationHandler{}
		fh.OnFlowEndpointLatency(
			func(_ context.Context, _ *webhooks.FlowNotificationContext, _ *webhooks.EndpointLatencyDetails) error {
				return nil
			},
		)
		if fh.EndpointLatency == nil {
			t.Error("EndpointLatency should not be nil after OnFlowEndpointLatency")
		}
	})

	t.Run("OnFallback sets FallbackHandler", func(t *testing.T) {
		t.Parallel()
		fh := &webhooks.FlowNotificationHandler{}
		fh.OnFallback(func(_ context.Context, _ *webhooks.FlowNotificationContext, _ *webhooks.Value) error {
			return nil
		})
		if fh.FallbackHandler == nil {
			t.Error("FallbackHandler should not be nil after OnFallback")
		}
	})

	t.Run("OnFlowEndpointAvailability sets EndpointAvailability", func(t *testing.T) {
		t.Parallel()
		fh := &webhooks.FlowNotificationHandler{}
		fh.OnFlowEndpointAvailability(
			func(_ context.Context, _ *webhooks.FlowNotificationContext, _ *webhooks.EndpointAvailabilityDetails) error {
				return nil
			},
		)
		if fh.EndpointAvailability == nil {
			t.Error("EndpointAvailability should not be nil after OnFlowEndpointAvailability")
		}
	})
}

func TestFlowNotificationHandler_DedicatedOverridesFallback(t *testing.T) {
	t.Parallel()

	nctx := &webhooks.FlowNotificationContext{FlowID: "flow-1"}

	fh := &webhooks.FlowNotificationHandler{}

	var dedicatedCalled, fallbackCalled bool

	fh.OnFlowStatusChange(
		func(_ context.Context, _ *webhooks.FlowNotificationContext, _ *webhooks.StatusChangeDetails) error {
			dedicatedCalled = true
			return nil
		},
	)
	fh.OnFallback(func(_ context.Context, _ *webhooks.FlowNotificationContext, _ *webhooks.Value) error {
		fallbackCalled = true
		return nil
	})

	err := fh.Handle(context.Background(), nctx, &webhooks.Value{
		Event:     webhooks.EventFlowStatusChange,
		OldStatus: "DRAFT",
		NewStatus: "PUBLISHED",
	})
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
