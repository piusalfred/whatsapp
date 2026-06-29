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

// Flow types and FlowNotificationHandler for WhatsApp Flows webhooks.
// Handles flow status changes, client/endpoint error rates, endpoint
// latency, endpoint availability, and flow version expiry warnings.

package webhooks

import (
	"context"
)

type (
	// FlowEventHandler is a shorthand for EventHandler[FlowNotificationContext, T].
	FlowEventHandler[T any] EventHandler[FlowNotificationContext, T]

	// FlowEventHandlerFunc is a shorthand for EventHandlerFunc[FlowNotificationContext, T].
	FlowEventHandlerFunc[T any] EventHandlerFunc[FlowNotificationContext, T]
)

func (f FlowEventHandlerFunc[T]) HandleEvent(ctx context.Context,
	ntx *FlowNotificationContext, notification *T,
) error {
	return f(ctx, ntx, notification)
}

// FlowNotificationContext carries the metadata for a flows webhook notification.
type FlowNotificationContext struct {
	NotificationObject string // Corresponds to the 'object' field
	EntryID            string // Corresponds to the 'id' field in Entry
	EntryTime          int64  // Corresponds to the 'time' field in Entry
	ChangeField        string // Corresponds to the 'field' in Changes
	EventName          string // Corresponds to 'event' field in Value
	EventMessage       string // Corresponds to 'message' field in Value
	FlowID             string // Corresponds to 'flow_id' field in Value
}

type FlowStatusHandler = FlowEventHandler[StatusChangeDetails]

type FlowClientErrorRateHandler = FlowEventHandler[ClientErrorRateDetails]

type FlowEndpointErrorRateHandler = FlowEventHandler[EndpointErrorRateDetails]

type (
	FlowEndpointLatencyHandler      = FlowEventHandler[EndpointLatencyDetails]
	FlowEndpointAvailabilityHandler = FlowEventHandler[EndpointAvailabilityDetails]
)

type (
	StatusChangeDetails struct {
		OldStatus string `json:"old_status,omitempty"`
		NewStatus string `json:"new_status,omitempty"`
	}

	ClientErrorRateDetails struct {
		ErrorRate  float64     `json:"error_rate,omitempty"`
		Threshold  int         `json:"threshold,omitempty"`
		AlertState string      `json:"alert_state,omitempty"`
		Errors     []ErrorInfo `json:"errors,omitempty"`
	}

	EndpointErrorRateDetails struct {
		ErrorRate  float64     `json:"error_rate,omitempty"`
		Threshold  int         `json:"threshold,omitempty"`
		AlertState string      `json:"alert_state,omitempty"`
		Errors     []ErrorInfo `json:"errors,omitempty"`
	}

	EndpointLatencyDetails struct {
		P50Latency    int    `json:"p50_latency,omitempty"`
		P90Latency    int    `json:"p90_latency,omitempty"`
		RequestsCount int    `json:"requests_count,omitempty"`
		Threshold     int    `json:"threshold,omitempty"`
		AlertState    string `json:"alert_state,omitempty"`
	}

	EndpointAvailabilityDetails struct {
		Availability int    `json:"availability"`
		Threshold    int    `json:"threshold,omitempty"`
		AlertState   string `json:"alert_state,omitempty"`
	}
)

// FlowNotificationHandler groups all per-event-type handlers for the flows
// webhook and a fallback for unhandled events.
//
// Each exported field accepts a [FlowEventHandler[T]] for one WhatsApp flow
// event type. Leave a field nil to skip that event during dispatch or to let
// it fall through to [FallbackHandler].
//
// # Concurrency
//
// FlowNotificationHandler is safe for concurrent calls to
// [FlowNotificationHandler.Handle] (read-only access to registered callbacks).
// It is not safe for concurrent modification — register all handlers before
// the handler starts serving requests. See [Handler] for the top-level
// concurrency contract.
//
// Usage:
//
//	fh := &FlowNotificationHandler{}
//	fh.OnFlowStatusChange(myStatusHandler)
//	fh.OnFallback(myFallback) // catches known events without a handler
type FlowNotificationHandler struct {
	// Status handles FLOW_STATUS_CHANGE events. Payload: [StatusChangeDetails].
	Status FlowEventHandler[StatusChangeDetails]
	// ClientErrorRate handles CLIENT_ERROR_RATE events. Payload: [ClientErrorRateDetails].
	ClientErrorRate FlowEventHandler[ClientErrorRateDetails]
	// EndpointErrorRate handles ENDPOINT_ERROR_RATE events. Payload: [EndpointErrorRateDetails].
	EndpointErrorRate FlowEventHandler[EndpointErrorRateDetails]
	// EndpointLatency handles ENDPOINT_LATENCY events. Payload: [EndpointLatencyDetails].
	EndpointLatency FlowEventHandler[EndpointLatencyDetails]
	// EndpointAvailability handles ENDPOINT_AVAILABILITY events. Payload: [EndpointAvailabilityDetails].
	EndpointAvailability FlowEventHandler[EndpointAvailabilityDetails]

	// FallbackHandler is called for any flow event that does not have a
	// dedicated handler set — both unknown event types and known types
	// left nil. When nil, those events are silently acknowledged (HTTP 200)
	// to prevent WhatsApp from retrying.
	FallbackHandler func(ctx context.Context, nctx *FlowNotificationContext, value *Value) error
}

// OnFlowStatusChange sets the handler for FLOW_STATUS_CHANGE events.
func (fh *FlowNotificationHandler) OnFlowStatusChange(
	fn func(ctx context.Context, nctx *FlowNotificationContext, details *StatusChangeDetails) error,
) {
	fh.Status = FlowEventHandlerFunc[StatusChangeDetails](fn)
}

// OnFlowClientErrorRate sets the handler for CLIENT_ERROR_RATE events.
func (fh *FlowNotificationHandler) OnFlowClientErrorRate(
	fn func(ctx context.Context, nctx *FlowNotificationContext, details *ClientErrorRateDetails) error,
) {
	fh.ClientErrorRate = FlowEventHandlerFunc[ClientErrorRateDetails](fn)
}

// OnFlowEndpointErrorRate sets the handler for ENDPOINT_ERROR_RATE events.
func (fh *FlowNotificationHandler) OnFlowEndpointErrorRate(
	fn func(ctx context.Context, nctx *FlowNotificationContext, details *EndpointErrorRateDetails) error,
) {
	fh.EndpointErrorRate = FlowEventHandlerFunc[EndpointErrorRateDetails](fn)
}

// OnFlowEndpointLatency sets the handler for ENDPOINT_LATENCY events.
func (fh *FlowNotificationHandler) OnFlowEndpointLatency(
	fn func(ctx context.Context, nctx *FlowNotificationContext, details *EndpointLatencyDetails) error,
) {
	fh.EndpointLatency = FlowEventHandlerFunc[EndpointLatencyDetails](fn)
}

// OnFallback sets the catch-all handler for flow events without a dedicated
// handler — covers unknown event types and known types left nil.
// Equivalent to assigning [FallbackHandler] directly.
func (fh *FlowNotificationHandler) OnFallback(
	fn func(ctx context.Context, nctx *FlowNotificationContext, value *Value) error,
) {
	fh.FallbackHandler = fn
}

// OnFlowEndpointAvailability sets the handler for ENDPOINT_AVAILABILITY events.
func (fh *FlowNotificationHandler) OnFlowEndpointAvailability(
	fn func(ctx context.Context, nctx *FlowNotificationContext, details *EndpointAvailabilityDetails) error,
) {
	fh.EndpointAvailability = FlowEventHandlerFunc[EndpointAvailabilityDetails](fn)
}

// Handle dispatches the flow value to the correct event handler based on
// value.Event.
//
//  1. If a dedicated handler is registered and not nil, it is called with
//     the extracted details (e.g., [Value.FlowStatusChange]).
//  2. Otherwise, falls back to [FallbackHandler] — this covers both unknown
//     flow event types and known types without a dedicated handler.
//  3. If [FallbackHandler] is also nil, the event is silently skipped (HTTP 200).
//
//nolint:wrapcheck // typed dispatch; user handlers own error context
func (fh *FlowNotificationHandler) Handle(
	ctx context.Context,
	nctx *FlowNotificationContext,
	value *Value,
) error {
	switch value.Event {
	case EventFlowStatusChange:
		if fh.Status != nil {
			return fh.Status.HandleEvent(ctx, nctx, value.FlowStatusChange())
		}
	case EventClientErrorRate:
		if fh.ClientErrorRate != nil {
			return fh.ClientErrorRate.HandleEvent(ctx, nctx, value.FlowClientErrorRate())
		}
	case EventEndpointErrorRate:
		if fh.EndpointErrorRate != nil {
			return fh.EndpointErrorRate.HandleEvent(ctx, nctx, value.FlowEndpointErrorRate())
		}
	case EventEndpointLatency:
		if fh.EndpointLatency != nil {
			return fh.EndpointLatency.HandleEvent(ctx, nctx, value.FlowEndpointLatency())
		}
	case EventEndpointAvailability:
		if fh.EndpointAvailability != nil {
			return fh.EndpointAvailability.HandleEvent(ctx, nctx, value.FlowEndpointAvailability())
		}
	}

	// Known event without handler, or unknown flow event — fall back.
	if fh.FallbackHandler != nil {
		return fh.FallbackHandler(ctx, nctx, value)
	}

	return nil
}

// FlowStatusChange extracts status change details from a flows webhook value.
func (value *Value) FlowStatusChange() *StatusChangeDetails {
	return &StatusChangeDetails{
		OldStatus: value.OldStatus,
		NewStatus: value.NewStatus,
	}
}

// FlowClientErrorRate extracts client error rate details from a flows webhook value.
func (value *Value) FlowClientErrorRate() *ClientErrorRateDetails {
	return &ClientErrorRateDetails{
		ErrorRate:  value.ErrorRate,
		Threshold:  value.Threshold,
		AlertState: value.AlertState,
		Errors:     value.Errors,
	}
}

// FlowEndpointErrorRate extracts endpoint error rate details from a flows webhook value.
func (value *Value) FlowEndpointErrorRate() *EndpointErrorRateDetails {
	return &EndpointErrorRateDetails{
		ErrorRate:  value.ErrorRate,
		Threshold:  value.Threshold,
		AlertState: value.AlertState,
		Errors:     value.Errors,
	}
}

// FlowEndpointLatency extracts endpoint latency details from a flows webhook value.
func (value *Value) FlowEndpointLatency() *EndpointLatencyDetails {
	return &EndpointLatencyDetails{
		P50Latency:    value.P50Latency,
		P90Latency:    value.P90Latency,
		RequestsCount: value.RequestsCount,
		Threshold:     value.Threshold,
		AlertState:    value.AlertState,
	}
}

// FlowEndpointAvailability extracts endpoint availability details from a flows webhook value.
func (value *Value) FlowEndpointAvailability() *EndpointAvailabilityDetails {
	return &EndpointAvailabilityDetails{
		Availability: value.Availability,
		Threshold:    value.Threshold,
		AlertState:   value.AlertState,
	}
}

// OnFlowStatusChange registers a handler for flow status change events in the flows webhook.
func (handler *Handler) OnFlowStatusChange(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *StatusChangeDetails) error,
) {
	handler.flows.OnFlowStatusChange(fn)
}

// SetFlowStatusChangeHandler sets the handler for flow status change events.
func (handler *Handler) SetFlowStatusChangeHandler(fn FlowStatusHandler) {
	handler.flows.Status = fn
}

// OnFlowClientErrorRate registers a handler for flow client error rate events in the flows webhook.
func (handler *Handler) OnFlowClientErrorRate(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *ClientErrorRateDetails) error,
) {
	handler.flows.OnFlowClientErrorRate(fn)
}

// SetFlowClientErrorRateHandler sets the handler for flow client error rate events.
func (handler *Handler) SetFlowClientErrorRateHandler(
	fn FlowClientErrorRateHandler,
) {
	handler.flows.ClientErrorRate = fn
}

// OnFlowEndpointErrorRate registers a handler for flow endpoint error rate events in the flows webhook.
func (handler *Handler) OnFlowEndpointErrorRate(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *EndpointErrorRateDetails) error,
) {
	handler.flows.OnFlowEndpointErrorRate(fn)
}

// SetFlowEndpointErrorRateHandler sets the handler for flow endpoint error rate events.
func (handler *Handler) SetFlowEndpointErrorRateHandler(
	fn FlowEndpointErrorRateHandler,
) {
	handler.flows.EndpointErrorRate = fn
}

// OnFlowEndpointLatency registers a handler for flow endpoint latency events in the flows webhook.
func (handler *Handler) OnFlowEndpointLatency(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *EndpointLatencyDetails) error,
) {
	handler.flows.OnFlowEndpointLatency(fn)
}

// SetFlowEndpointLatencyHandler sets the handler for flow endpoint latency events.
func (handler *Handler) SetFlowEndpointLatencyHandler(
	fn FlowEndpointLatencyHandler,
) {
	handler.flows.EndpointLatency = fn
}

// OnFlowEndpointAvailability registers a handler for flow endpoint availability events in the flows webhook.
func (handler *Handler) OnFlowEndpointAvailability(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *EndpointAvailabilityDetails) error,
) {
	handler.flows.OnFlowEndpointAvailability(fn)
}

// SetFlowEndpointAvailabilityHandler sets the handler for flow endpoint availability events.
func (handler *Handler) SetFlowEndpointAvailabilityHandler(
	fn FlowEndpointAvailabilityHandler,
) {
	handler.flows.EndpointAvailability = fn
}
