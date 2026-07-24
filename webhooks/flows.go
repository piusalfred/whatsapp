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
	"fmt"
)

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
// Each field accepts a [FlowEventHandler[T]] for one WhatsApp flow
// event type. Leave a handler nil to skip that event during dispatch or to let
// it fall through to the [Fallback] method.
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
	status FlowEventHandler[StatusChangeDetails]
	// ClientErrorRate handles CLIENT_ERROR_RATE events. Payload: [ClientErrorRateDetails].
	clientErrorRate FlowEventHandler[ClientErrorRateDetails]
	// EndpointErrorRate handles ENDPOINT_ERROR_RATE events. Payload: [EndpointErrorRateDetails].
	endpointErrorRate FlowEventHandler[EndpointErrorRateDetails]
	// EndpointLatency handles ENDPOINT_LATENCY events. Payload: [EndpointLatencyDetails].
	endpointLatency FlowEventHandler[EndpointLatencyDetails]
	// EndpointAvailability handles ENDPOINT_AVAILABILITY events. Payload: [EndpointAvailabilityDetails].
	endpointAvailability FlowEventHandler[EndpointAvailabilityDetails]

	// Fallback is called for any flow event that does not have a dedicated
	// handler set — both unknown event types and known types left nil.
	// When nil, those events are silently acknowledged (HTTP 200) to
	// prevent WhatsApp from retrying.
	fallback FallbackHandler

	errorHandler ErrorHandler
}

// OnFlowStatusChange sets the handler for FLOW_STATUS_CHANGE events.
func (fh *FlowNotificationHandler) OnFlowStatusChange(h FlowStatusHandler) {
	fh.status = h
}

// OnFlowClientErrorRate sets the handler for CLIENT_ERROR_RATE events.
func (fh *FlowNotificationHandler) OnFlowClientErrorRate(h FlowClientErrorRateHandler) {
	fh.clientErrorRate = h
}

// OnFlowEndpointErrorRate sets the handler for ENDPOINT_ERROR_RATE events.
func (fh *FlowNotificationHandler) OnFlowEndpointErrorRate(h FlowEndpointErrorRateHandler) {
	fh.endpointErrorRate = h
}

// OnFlowEndpointLatency sets the handler for ENDPOINT_LATENCY events.
func (fh *FlowNotificationHandler) OnFlowEndpointLatency(h FlowEndpointLatencyHandler) {
	fh.endpointLatency = h
}

// OnFallback sets the catch-all handler for flow events without a dedicated
// handler — covers unknown event types and known types left nil.
func (fh *FlowNotificationHandler) OnFallback(h FallbackHandler) {
	fh.fallback = h
}

// OnError sets the error handler for flow event handlers. When nil, errors are returned as-is (passthrough).
func (fh *FlowNotificationHandler) OnError(h ErrorHandler) {
	fh.errorHandler = h
}

// OnFlowEndpointAvailability sets the handler for ENDPOINT_AVAILABILITY events.
func (fh *FlowNotificationHandler) OnFlowEndpointAvailability(h FlowEndpointAvailabilityHandler) {
	fh.endpointAvailability = h
}

// HandleError routes an error through the FlowNotificationHandler's ErrorHandler.
// When the dedicated error handler is nil, the error is returned as-is.
func (fh *FlowNotificationHandler) HandleError(ctx context.Context, err error) error {
	return execErrorHandler(ctx, fh.errorHandler, err)
}

// Fallback routes an unhandled flow event through the Fallback
// catch-all. Returns nil when no fallback handler is set (silent skip).
func (fh *FlowNotificationHandler) Fallback(
	ctx context.Context,
	event NotificationEvent,
) error {
	if fh.fallback == nil {
		return nil
	}
	if err := fh.fallback.Handle(ctx, event); err != nil {
		return fmt.Errorf("flow fallback: %w", err)
	}
	return nil
}

// Handle dispatches the flow value to the correct event handler based on
// event.Value.Event.
//
//  1. If a dedicated handler is registered and not nil, it is called with
//     the extracted details (e.g., [Value.FlowStatusChange]).
//  2. Otherwise, falls back to the [Fallback] method — this covers both
//     [RunChangeHandler]) will invoke the [Fallback] method.
//  3. Errors from dedicated handlers are routed through [HandleError].

func (fh *FlowNotificationHandler) Handle(
	ctx context.Context,
	event NotificationEvent,
) error {
	if event.Value == nil {
		return nil
	}

	nctx := &FlowNotificationContext{
		NotificationObject: event.Object,
		EntryID:            event.EntryID,
		EntryTime:          event.Time,
		ChangeField:        event.Field,
		EventName:          event.Value.Event,
		EventMessage:       event.Value.Message,
		FlowID:             event.Value.FlowID,
	}

	value := event.Value

	switch value.Event {
	case EventFlowStatusChange:
		if fh.status != nil {
			req := &FlowRequest[StatusChangeDetails]{
				Context: nctx,
				Payload: value.FlowStatusChange(),
			}
			return fh.HandleError(ctx, fh.status.Handle(ctx, req))
		}
	case EventClientErrorRate:
		if fh.clientErrorRate != nil {
			req := &FlowRequest[ClientErrorRateDetails]{
				Context: nctx,
				Payload: value.FlowClientErrorRate(),
			}
			return fh.HandleError(ctx, fh.clientErrorRate.Handle(ctx, req))
		}
	case EventEndpointErrorRate:
		if fh.endpointErrorRate != nil {
			req := &FlowRequest[EndpointErrorRateDetails]{
				Context: nctx,
				Payload: value.FlowEndpointErrorRate(),
			}
			return fh.HandleError(ctx, fh.endpointErrorRate.Handle(ctx, req))
		}
	case EventEndpointLatency:
		if fh.endpointLatency != nil {
			req := &FlowRequest[EndpointLatencyDetails]{
				Context: nctx,
				Payload: value.FlowEndpointLatency(),
			}
			return fh.HandleError(ctx, fh.endpointLatency.Handle(ctx, req))
		}
	case EventEndpointAvailability:
		if fh.endpointAvailability != nil {
			req := &FlowRequest[EndpointAvailabilityDetails]{
				Context: nctx,
				Payload: value.FlowEndpointAvailability(),
			}
			return fh.HandleError(ctx, fh.endpointAvailability.Handle(ctx, req))
		}
	}

	return ErrEventNotHandled
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

// IsEventHandlerImplemented reports whether a handler is registered for the
// flow event carried by this NotificationEvent. It checks the event field
// against known flow event types and returns true when the matching
// sub-handler is non-nil.
func (fh *FlowNotificationHandler) IsEventHandlerImplemented(event NotificationEvent) bool {
	if event.Value == nil {
		return false
	}
	switch event.Value.Event {
	case EventFlowStatusChange:
		return fh.status != nil
	case EventClientErrorRate:
		return fh.clientErrorRate != nil
	case EventEndpointErrorRate:
		return fh.endpointErrorRate != nil
	case EventEndpointLatency:
		return fh.endpointLatency != nil
	case EventEndpointAvailability:
		return fh.endpointAvailability != nil
	}
	return false
}

var _ EventHandler = (*FlowNotificationHandler)(nil)

func (handler *Handler) OnFlowStatusChange(h FlowStatusHandler) {
	handler.flows.OnFlowStatusChange(h)
}

// OnFlowClientErrorRate registers a handler for flow client error rate events in the flows webhook.
func (handler *Handler) OnFlowClientErrorRate(h FlowClientErrorRateHandler) {
	handler.flows.OnFlowClientErrorRate(h)
}

// OnFlowEndpointErrorRate registers a handler for flow endpoint error rate events in the flows webhook.
func (handler *Handler) OnFlowEndpointErrorRate(h FlowEndpointErrorRateHandler) {
	handler.flows.OnFlowEndpointErrorRate(h)
}

// OnFlowEndpointLatency registers a handler for flow endpoint latency events in the flows webhook.
func (handler *Handler) OnFlowEndpointLatency(h FlowEndpointLatencyHandler) {
	handler.flows.OnFlowEndpointLatency(h)
}

// OnFlowEndpointAvailability registers a handler for flow endpoint availability events in the flows webhook.
func (handler *Handler) OnFlowEndpointAvailability(h FlowEndpointAvailabilityHandler) {
	handler.flows.OnFlowEndpointAvailability(h)
}
