//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Call types and CallsHandler for WhatsApp Calling API webhooks. Handles call
// connect (WebRTC SDP), call created (SIP), call terminate, and call status
// (ringing/accepted/rejected) events delivered through the "calls" webhook field.

package webhooks

import (
	"context"
	"fmt"
)

// CallNotificationContext carries the metadata for a calls webhook notification.
type CallNotificationContext struct {
	NotificationObject string     // Corresponds to the 'object' field
	EntryID            string     // Corresponds to the 'id' field in Entry
	EntryTime          int64      // Corresponds to the 'time' field in Entry
	ChangeField        string     // Corresponds to the 'field' in Changes
	MessagingProduct   string     // Corresponds to 'messaging_product' field in Value
	Contacts           []*Contact // Callers involved in the call
	Metadata           *Metadata  // Phone number metadata
}

// CallRequest carries context and payload for a calls webhook event.
type CallRequest[T any] struct {
	// Context identifies the source WABA, the entry/change metadata, and the
	// call participants.
	Context *CallNotificationContext
	// Payload is the typed event details: *Call for connect/created/terminate
	// events, *Status for call status changes (ringing/accepted/rejected).
	Payload *T
}

// CallsEventHandler is the interface for handling typed call webhook events.
// It receives a [CallRequest] carrying the notification context and the typed
// event payload.
type CallsEventHandler[T any] interface {
	Handle(ctx context.Context, req *CallRequest[T]) error
}

// CallsEventHandlerFunc is an adapter that allows a plain function with the
// (ctx, *CallRequest[T]) signature to be used as a [CallsEventHandler].
type CallsEventHandlerFunc[T any] func(ctx context.Context, req *CallRequest[T]) error

// Handle implements [CallsEventHandler] by calling the underlying function.
func (f CallsEventHandlerFunc[T]) Handle(ctx context.Context, req *CallRequest[T]) error {
	return f(ctx, req)
}

// NewNoOpCallsEventHandler returns a [CallsEventHandler] that silently
// discards all events. Useful as a placeholder.
func NewNoOpCallsEventHandler[T any]() CallsEventHandler[T] {
	return CallsEventHandlerFunc[T](func(_ context.Context, _ *CallRequest[T]) error {
		return nil
	})
}

// Type aliases for each call event type. Each accepts a [CallsEventHandler]
// parameterized on the appropriate payload type.
type (
	// CallConnectHandler handles call connect events (event: "connect").
	// Payload: [Call] — contains session SDP for WebRTC connection.
	CallConnectHandler = CallsEventHandler[Call]

	// CallCreatedHandler handles call created events (event: "call_created").
	// Payload: [Call] — SIP calls only, no session object.
	CallCreatedHandler = CallsEventHandler[Call]

	// CallTerminateHandler handles call terminate events (event: "terminate").
	// Payload: [Call] — includes status (COMPLETED/FAILED), start_time,
	// end_time, and duration.
	CallTerminateHandler = CallsEventHandler[Call]

	// CallStatusHandler handles call status events (statuses with type "call").
	// Payload: [Status] — status is RINGING, ACCEPTED, or REJECTED.
	CallStatusHandler = CallsEventHandler[Status]
)

// CallsHandler groups all per-event-type handlers for the calls webhook field
// and a fallback for unhandled events.
//
// Each exported field accepts a [CallsEventHandler[T]] for one WhatsApp call
// event type. Leave a field nil to let it fall through to [FallbackHandler].
//
// # Concurrency
//
// CallsHandler is safe for concurrent calls to [CallsHandler.Handle]
// (read-only access to registered callbacks). It is not safe for concurrent
// modification — register all handlers before the handler starts serving
// requests. See [Handler] for the top-level concurrency contract.
//
// Usage:
//
//	ch := &CallsHandler{}
//	ch.OnCallConnect(myConnectHandler)
//	ch.OnCallTerminate(myTerminateHandler)
//	ch.OnFallback(myFallback) // catches events without a dedicated handler
type CallsHandler struct {
	// Connect handles call connect events (event: "connect"). Payload: [Call].
	Connect CallsEventHandler[Call]
	// Created handles call created events (event: "call_created"). Payload: [Call].
	Created CallsEventHandler[Call]
	// Terminate handles call terminate events (event: "terminate"). Payload: [Call].
	Terminate CallsEventHandler[Call]
	// Status handles call status events (type: "call" in statuses array).
	// Payload: [Status].
	Status CallsEventHandler[Status]

	// Fallback is called for any call event that does not have a dedicated
	// handler set — both unknown event types and known types left nil.
	// When nil, those events are silently acknowledged (HTTP 200) to prevent
	// WhatsApp from retrying.
	Fallback FallbackHandler

	// ErrorHandler is called when a handler returns an error. When nil, the
	// error is returned as-is (passthrough).
	ErrorHandler ErrorHandler
}

// OnCallConnect sets the handler for call connect events (event: "connect").
func (ch *CallsHandler) OnCallConnect(h CallConnectHandler) {
	ch.Connect = h
}

// OnCallCreated sets the handler for call created events (event: "call_created").
func (ch *CallsHandler) OnCallCreated(h CallCreatedHandler) {
	ch.Created = h
}

// OnCallTerminate sets the handler for call terminate events (event: "terminate").
func (ch *CallsHandler) OnCallTerminate(h CallTerminateHandler) {
	ch.Terminate = h
}

// OnCallStatus sets the handler for call status events (type: "call").
func (ch *CallsHandler) OnCallStatus(h CallStatusHandler) {
	ch.Status = h
}

// OnFallback sets the catch-all handler for call events without a dedicated
// handler — covers unknown event types and known types left nil.
// Equivalent to assigning [FallbackHandler] directly.
func (ch *CallsHandler) OnFallback(h FallbackHandler) {
	ch.Fallback = h
}

// handleError routes an error through the CallsHandler's ErrorHandler.
// When ErrorHandler is nil, the error is returned as-is (passthrough).
func (ch *CallsHandler) handleError(ctx context.Context, err error) error {
	return handleSubHandlerError(ctx, ch.ErrorHandler, err)
}

// executeFallback routes an unhandled call event through the Fallback
// catch-all. Returns nil when Fallback is nil (silent skip).
func (ch *CallsHandler) executeFallback(ctx context.Context, ne NotificationEntry, change Change) error {
	if ch.Fallback == nil {
		return nil
	}
	if err := ch.Fallback.Handle(ctx, ne, change); err != nil {
		return fmt.Errorf("calls fallback: %w", err)
	}
	return nil
}

// Handle dispatches the calls webhook value to the correct event handler.
//
// Dispatch order:
//  1. If value.Statuses contains items with type "call", dispatch to Status handler.
//  2. If value.Calls contains items, dispatch each by event type:
//     "connect" → Connect, "call_created" → Created, "terminate" → Terminate.
//  3. Unhandled events or nil handlers fall through to [FallbackHandler].
//  4. If [FallbackHandler] is also nil, the event is silently skipped (HTTP 200).
//
//nolint:gocognit // dispatch switch
func (ch *CallsHandler) Handle(
	ctx context.Context,
	ne NotificationEntry,
	change Change,
) error {
	if change.Value == nil {
		return nil
	}

	nctx := &CallNotificationContext{
		NotificationObject: ne.Object,
		EntryID:            ne.ID,
		EntryTime:          ne.Time,
		ChangeField:        change.Field,
		MessagingProduct:   change.Value.MessagingProduct,
		Contacts:           change.Value.Contacts,
		Metadata:           change.Value.Metadata,
	}

	// Phase 1: Dispatch call statuses (type "call"). These arrive as a
	// statuses array in the value, distinct from the calls array.
	for _, status := range change.Value.Statuses {
		if status == nil || status.Type != "call" {
			continue
		}
		if ch.Status != nil {
			req := &CallRequest[Status]{
				Context: nctx,
				Payload: status,
			}
			if err := ch.Status.Handle(ctx, req); err != nil {
				return ch.handleError(ctx, fmt.Errorf("calls status: %w", err))
			}
			return nil
		}
		// No dedicated status handler → fallback.
		return ch.executeFallback(ctx, ne, change)
	}

	// Phase 2: Dispatch call events from the calls array.
	for _, call := range change.Value.Calls {
		if call == nil {
			continue
		}
		switch call.Event {
		case "connect":
			if ch.Connect != nil {
				req := &CallRequest[Call]{
					Context: nctx,
					Payload: call,
				}
				if err := ch.Connect.Handle(ctx, req); err != nil {
					return ch.handleError(ctx, fmt.Errorf("calls connect: %w", err))
				}
				continue
			}
		case "call_created":
			if ch.Created != nil {
				req := &CallRequest[Call]{
					Context: nctx,
					Payload: call,
				}
				if err := ch.Created.Handle(ctx, req); err != nil {
					return ch.handleError(ctx, fmt.Errorf("calls created: %w", err))
				}
				continue
			}
		case "terminate":
			if ch.Terminate != nil {
				req := &CallRequest[Call]{
					Context: nctx,
					Payload: call,
				}
				if err := ch.Terminate.Handle(ctx, req); err != nil {
					return ch.handleError(ctx, fmt.Errorf("calls terminate: %w", err))
				}
				continue
			}
		}
		// Unknown event type or nil handler → fallback for this call.
		return ch.executeFallback(ctx, ne, change)
	}

	return nil
}

type (
	Call struct {
		ID                    string       `json:"id"` // The WhatsApp call ID
		To                    string       `json:"to"` // The WhatsApp user's phone number (callee)
		ToUserID              string       `json:"to_user_id,omitempty"`
		ToParentUserID        string       `json:"to_parent_user_id,omitempty"`
		From                  string       `json:"from"`
		Event                 string       `json:"event"`
		Timestamp             string       `json:"timestamp"`
		Direction             string       `json:"direction"`
		DeepLinkPayload       string       `json:"deeplink_payload,omitempty"`
		CTAPayload            string       `json:"cta_payload,omitempty"`
		Status                string       `json:"status"`
		StartTime             string       `json:"start_time"`
		EndTime               string       `json:"end_time"`
		Duration              int          `json:"duration"`
		BizOpaqueCallbackData string       `json:"biz_opaque_callback_data,omitempty"`
		Session               *CallSession `json:"session,omitempty"`
		Connection            *Connection  `json:"connection,omitempty"`
	}

	CallSession struct {
		SDPType string `json:"sdp_type"`
		SDP     string `json:"sdp"`
	}

	WebRTC struct {
		SDP string `json:"sdp"`
	}

	Connection struct {
		WebRTC *WebRTC `json:"webrtc,omitempty"`
	}

	CallStatusUpdate struct {
		MessagingProduct string
		Contacts         []*Contact
		Metadata         *Metadata `json:"metadata,omitempty"`
		Calls            []*Call
		Errors           []ErrorInfo
	}
)

func (value *Value) CallStatusUpdate() *CallStatusUpdate {
	return &CallStatusUpdate{
		MessagingProduct: value.MessagingProduct,
		Metadata:         value.Metadata,
		Contacts:         value.Contacts,
		Calls:            value.Calls,
		Errors:           value.Errors,
	}
}

// OnCallConnect registers a handler for call connect events.
func (handler *Handler) OnCallConnect(h CallConnectHandler) {
	handler.ensureCalls().OnCallConnect(h)
}

// OnCallCreated registers a handler for call created events.
func (handler *Handler) OnCallCreated(h CallCreatedHandler) {
	handler.ensureCalls().OnCallCreated(h)
}

// OnCallTerminate registers a handler for call terminate events.
func (handler *Handler) OnCallTerminate(h CallTerminateHandler) {
	handler.ensureCalls().OnCallTerminate(h)
}

// OnCallStatus registers a handler for call status events (type "call").
func (handler *Handler) OnCallStatus(h CallStatusHandler) {
	handler.ensureCalls().OnCallStatus(h)
}

// CallPermissionReply represents a WhatsApp user's response to a call
// permission request. Delivered through the "messages" webhook field as
// an interactive message with type "call_permission_reply".
//
// Response can be "accept" or "reject". IsPermanent indicates whether
// the permission is permanent. ResponseSource is "user_action" for
// explicit user approval/rejection or "automatic" for permissions
// granted by initiating a call.
type CallPermissionReply struct {
	Response            string `json:"response"`
	IsPermanent         bool   `json:"is_permanent"`
	ExpirationTimestamp string `json:"expiration_timestamp"`
	ResponseSource      string `json:"response_source"`
}

// OnCallStatusUpdate registers a handler for call status updates via the
// business notification path. This is the legacy registration point;
// prefer [Handler.OnCallStatus] for the dedicated calls handler.
func (handler *Handler) OnCallStatusUpdate(h BusinessCallsHandler) {
	handler.ensureBusiness().Calls = h
}
