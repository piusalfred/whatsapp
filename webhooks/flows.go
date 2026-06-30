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

type (
	// FlowFallbackHandlerFunc adapts a bare function to the FlowFallbackHandler interface.
	FlowFallbackHandlerFunc func(ctx context.Context, nctx *FlowNotificationContext, value *Value) error

	// FlowFallbackHandler is called for any flow event that does not have a
	// dedicated handler set — both unknown event types and known types left nil.
	// When nil, those events are silently acknowledged (HTTP 200).
	FlowFallbackHandler interface {
		Handle(ctx context.Context, nctx *FlowNotificationContext, value *Value) error
	}
)

func (fn FlowFallbackHandlerFunc) Handle(ctx context.Context, nctx *FlowNotificationContext, value *Value) error {
	return fn(ctx, nctx, value)
}

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
	Fallback FlowFallbackHandler

	ErrorHandler ErrorHandler
}

// OnFlowStatusChange sets the handler for FLOW_STATUS_CHANGE events.
func (fh *FlowNotificationHandler) OnFlowStatusChange(h FlowStatusHandler) {
	fh.Status = h
}

// OnFlowClientErrorRate sets the handler for CLIENT_ERROR_RATE events.
func (fh *FlowNotificationHandler) OnFlowClientErrorRate(h FlowClientErrorRateHandler) {
	fh.ClientErrorRate = h
}

// OnFlowEndpointErrorRate sets the handler for ENDPOINT_ERROR_RATE events.
func (fh *FlowNotificationHandler) OnFlowEndpointErrorRate(h FlowEndpointErrorRateHandler) {
	fh.EndpointErrorRate = h
}

// OnFlowEndpointLatency sets the handler for ENDPOINT_LATENCY events.
func (fh *FlowNotificationHandler) OnFlowEndpointLatency(h FlowEndpointLatencyHandler) {
	fh.EndpointLatency = h
}

// OnFallback sets the catch-all handler for flow events without a dedicated
// handler — covers unknown event types and known types left nil.
// Equivalent to assigning [FallbackHandler] directly.
func (fh *FlowNotificationHandler) OnFallback(h FlowFallbackHandler) {
	fh.Fallback = h
}

// OnFlowEndpointAvailability sets the handler for ENDPOINT_AVAILABILITY events.
func (fh *FlowNotificationHandler) OnFlowEndpointAvailability(h FlowEndpointAvailabilityHandler) {
	fh.EndpointAvailability = h
}

// handleError routes an error through the FlowNotificationHandler's ErrorHandler.
// When ErrorHandler is nil, the error is returned as-is (passthrough).
func (fh *FlowNotificationHandler) handleError(ctx context.Context, err error) error {
	if fh.ErrorHandler == nil {
		return err
	}
	if handlerErr := fh.ErrorHandler.Handle(ctx, err); handlerErr != nil {
		return fmt.Errorf("error handler: %w", handlerErr)
	}
	return nil
}

// executeFallback routes an unhandled flow event through the Fallback
// catch-all. Returns nil when Fallback is nil (silent skip).
func (fh *FlowNotificationHandler) executeFallback(
	ctx context.Context,
	nctx *FlowNotificationContext,
	value *Value,
) error {
	if fh.Fallback == nil {
		return nil
	}
	if err := fh.Fallback.Handle(ctx, nctx, value); err != nil {
		return fmt.Errorf("flow fallback: %w", err)
	}
	return nil
}

// Handle dispatches the flow value to the correct event handler based on
// change.Value.Event.
//
//  1. If a dedicated handler is registered and not nil, it is called with
//     the extracted details (e.g., [Value.FlowStatusChange]).
//  2. Otherwise, falls back to [FlowFallbackHandler] — this covers both
//     unknown flow event types and known types without a dedicated handler.
//  3. If [FlowFallbackHandler] is also nil, the event is silently skipped
//     (HTTP 200).
func (fh *FlowNotificationHandler) Handle(
	ctx context.Context,
	ne NotificationEntry,
	change Change,
) error {
	if change.Value == nil {
		return nil
	}

	nctx := &FlowNotificationContext{
		NotificationObject: ne.Object,
		EntryID:            ne.ID,
		EntryTime:          ne.Time,
		ChangeField:        change.Field,
		EventName:          change.Value.Event,
		EventMessage:       change.Value.Message,
		FlowID:             change.Value.FlowID,
	}

	value := change.Value

	switch value.Event {
	case EventFlowStatusChange:
		if fh.Status != nil {
			req := &FlowRequest[StatusChangeDetails]{
				Context: nctx,
				Payload: value.FlowStatusChange(),
			}
			return fh.handleError(ctx, fh.Status.Handle(ctx, req))
		}
	case EventClientErrorRate:
		if fh.ClientErrorRate != nil {
			req := &FlowRequest[ClientErrorRateDetails]{
				Context: nctx,
				Payload: value.FlowClientErrorRate(),
			}
			return fh.handleError(ctx, fh.ClientErrorRate.Handle(ctx, req))
		}
	case EventEndpointErrorRate:
		if fh.EndpointErrorRate != nil {
			req := &FlowRequest[EndpointErrorRateDetails]{
				Context: nctx,
				Payload: value.FlowEndpointErrorRate(),
			}
			return fh.handleError(ctx, fh.EndpointErrorRate.Handle(ctx, req))
		}
	case EventEndpointLatency:
		if fh.EndpointLatency != nil {
			req := &FlowRequest[EndpointLatencyDetails]{
				Context: nctx,
				Payload: value.FlowEndpointLatency(),
			}
			return fh.handleError(ctx, fh.EndpointLatency.Handle(ctx, req))
		}
	case EventEndpointAvailability:
		if fh.EndpointAvailability != nil {
			req := &FlowRequest[EndpointAvailabilityDetails]{
				Context: nctx,
				Payload: value.FlowEndpointAvailability(),
			}
			return fh.handleError(ctx, fh.EndpointAvailability.Handle(ctx, req))
		}
	}

	return fh.executeFallback(ctx, nctx, value)
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
