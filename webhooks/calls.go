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
// connect (WebRTC SDP), call created (SIP), call terminate, call recording,
// and call status (ringing/accepted/rejected) events delivered through the
// "calls" webhook field.

package webhooks

import (
	"context"
	"fmt"

	"github.com/piusalfred/whatsapp/message/media"
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

	// CallRecordingAvailableHandler handles recording-available events
	// (event: "call_recording_available"). Payload: [Call] — call_recording
	// field carries the downloadable audio asset details.
	CallRecordingAvailableHandler = CallsEventHandler[Call]

	// CallTranscriptionAvailableHandler handles transcription-available events
	// (event: "call_transcription_available"). Payload: [Call] — call_transcript
	// field carries the downloadable transcript document details.
	CallTranscriptionAvailableHandler = CallsEventHandler[Call]
)

// CallsHandler groups all per-event-type handlers for the calls webhook field
// and a fallback for unhandled events.
//
// Each field accepts a [CallsEventHandler[T]] for one WhatsApp call
// event type. Leave a handler nil to let it fall through to the [Fallback] method.
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
	connect CallsEventHandler[Call]
	// Created handles call created events (event: "call_created"). Payload: [Call].
	created CallsEventHandler[Call]
	// Terminate handles call terminate events (event: "terminate"). Payload: [Call].
	terminate CallsEventHandler[Call]
	// Status handles call status events (type: "call" in statuses array).
	// Payload: [Status].
	callsStatus CallsEventHandler[Status]
	// RecordingAvailable handles recording-available events
	// (event: "call_recording_available"). Payload: [Call].
	recordingAvailable CallsEventHandler[Call]
	// TranscriptionAvailable handles transcription-available events
	// (event: "call_transcription_available"). Payload: [Call].
	transcriptionAvailable CallsEventHandler[Call]

	// Fallback is called for any call event that does not have a dedicated
	// handler set — both unknown event types and known types left nil.
	// When nil, those events are silently acknowledged (HTTP 200) to prevent
	// WhatsApp from retrying.
	fallback FallbackHandler

	// ErrorHandler is called when a handler returns an error. When nil, the
	// error is returned as-is (passthrough).
	error ErrorHandler
}

// OnError sets the error handler for this domain handler. When nil, errors
// bubble up to the general error handler configured on [Handler].
func (ch *CallsHandler) OnError(h ErrorHandler) {
	ch.error = h
}

// OnCallConnect sets the handler for call connect events (event: "connect").
func (ch *CallsHandler) OnCallConnect(h CallConnectHandler) {
	ch.connect = h
}

// OnCallCreated sets the handler for call created events (event: "call_created").
func (ch *CallsHandler) OnCallCreated(h CallCreatedHandler) {
	ch.created = h
}

// OnCallTerminate sets the handler for call terminate events (event: "terminate").
func (ch *CallsHandler) OnCallTerminate(h CallTerminateHandler) {
	ch.terminate = h
}

// OnCallStatus sets the handler for call status events (type: "call").
func (ch *CallsHandler) OnCallStatus(h CallStatusHandler) {
	ch.callsStatus = h
}

// OnCallRecordingAvailable sets the handler for recording-available events
// (event: "call_recording_available"). Payload: [Call] with CallRecording populated.
func (ch *CallsHandler) OnCallRecordingAvailable(h CallRecordingAvailableHandler) {
	ch.recordingAvailable = h
}

// OnCallTranscriptionAvailable sets the handler for transcription-available events
// (event: "call_transcription_available"). Payload: [Call] with CallTranscript populated.
func (ch *CallsHandler) OnCallTranscriptionAvailable(h CallTranscriptionAvailableHandler) {
	ch.transcriptionAvailable = h
}

// OnFallback sets the catch-all handler for call events without a dedicated
// handler — covers unknown event types and known types left nil.
func (ch *CallsHandler) OnFallback(h FallbackHandler) {
	ch.fallback = h
}

// HandleError routes an error through the CallsHandler's ErrorHandler.
// When the dedicated error handler is nil, the error is returned as-is.
func (ch *CallsHandler) HandleError(ctx context.Context, err error) error {
	return execErrorHandler(ctx, ch.error, err)
}

// Fallback routes an unhandled call event through the Fallback
// catch-all. Returns nil when no fallback handler is set (silent skip).
func (ch *CallsHandler) Fallback(ctx context.Context, event NotificationEvent) error {
	if ch.fallback == nil {
		return nil
	}
	if err := ch.fallback.Handle(ctx, event); err != nil {
		return fmt.Errorf("calls fallback: %w", err)
	}
	return nil
}

// Handle dispatches the calls webhook value to the correct event handler.
//
// Dispatch order:
//  1. If Value.Statuses contains items with type "call", dispatch to Status handler.
//  2. If Value.Calls contains items, dispatch each by event type:
//     "connect" → Connect, "call_created" → Created, "terminate" → Terminate,
//     "call_recording_available" → RecordingAvailable.
//  3. Unhandled events or nil handlers return [ErrEventNotHandled], signalling
//     the caller to invoke the [Fallback] method.
//
//nolint:gocognit // dispatch switch
func (ch *CallsHandler) Handle(
	ctx context.Context,
	event NotificationEvent,
) error {
	if event.Value == nil {
		return nil
	}

	nctx := &CallNotificationContext{
		NotificationObject: event.Object,
		EntryID:            event.EntryID,
		EntryTime:          event.Time,
		ChangeField:        event.Field,
		MessagingProduct:   event.Value.MessagingProduct,
		Contacts:           event.Value.Contacts,
		Metadata:           event.Value.Metadata,
	}

	// Phase 1: Dispatch call statuses (type "call"). These arrive as a
	// statuses array in the value, distinct from the calls array.
	for _, status := range event.Value.Statuses {
		if status == nil || status.Type != "call" {
			continue
		}
		if ch.callsStatus != nil {
			req := &CallRequest[Status]{
				Context: nctx,
				Payload: status,
			}
			if err := ch.callsStatus.Handle(ctx, req); err != nil {
				return ch.HandleError(ctx, fmt.Errorf("calls status: %w", err))
			}
			return nil
		}
		// No dedicated status handler → fallback.
		return ErrEventNotHandled
	}

	// Phase 2: Dispatch call events from the calls array.
	for _, call := range event.Value.Calls {
		if call == nil {
			continue
		}
		switch call.Event {
		case "connect":
			if ch.connect != nil {
				req := &CallRequest[Call]{
					Context: nctx,
					Payload: call,
				}
				if err := ch.connect.Handle(ctx, req); err != nil {
					return ch.HandleError(ctx, fmt.Errorf("calls connect: %w", err))
				}
				continue
			}
		case "call_created":
			if ch.created != nil {
				req := &CallRequest[Call]{
					Context: nctx,
					Payload: call,
				}
				if err := ch.created.Handle(ctx, req); err != nil {
					return ch.HandleError(ctx, fmt.Errorf("calls created: %w", err))
				}
				continue
			}
		case "terminate":
			if ch.terminate != nil {
				req := &CallRequest[Call]{
					Context: nctx,
					Payload: call,
				}
				if err := ch.terminate.Handle(ctx, req); err != nil {
					return ch.HandleError(ctx, fmt.Errorf("calls terminate: %w", err))
				}
				continue
			}
		case "call_recording_available":
			if ch.recordingAvailable != nil {
				req := &CallRequest[Call]{
					Context: nctx,
					Payload: call,
				}
				if err := ch.recordingAvailable.Handle(ctx, req); err != nil {
					return ch.HandleError(ctx, fmt.Errorf("calls recording: %w", err))
				}
				continue
			}
		case "call_transcription_available":
			if ch.transcriptionAvailable != nil {
				req := &CallRequest[Call]{
					Context: nctx,
					Payload: call,
				}
				if err := ch.transcriptionAvailable.Handle(ctx, req); err != nil {
					return ch.HandleError(ctx, fmt.Errorf("calls transcription: %w", err))
				}
				continue
			}
		}
		// Unknown event type or nil handler → fallback for this call.
		return ErrEventNotHandled
	}

	return nil
}

type (
	Call struct {
		ID                    string          `json:"id"` // The WhatsApp call ID
		To                    string          `json:"to"` // The WhatsApp user's phone number (callee)
		ToUserID              string          `json:"to_user_id,omitempty"`
		ToParentUserID        string          `json:"to_parent_user_id,omitempty"`
		From                  string          `json:"from"`
		Event                 string          `json:"event"`
		Timestamp             string          `json:"timestamp"`
		Direction             string          `json:"direction"`
		DeepLinkPayload       string          `json:"deeplink_payload,omitempty"`
		CTAPayload            string          `json:"cta_payload,omitempty"`
		Status                string          `json:"status"`
		StartTime             string          `json:"start_time"`
		EndTime               string          `json:"end_time"`
		Duration              int             `json:"duration"`
		BizOpaqueCallbackData string          `json:"biz_opaque_callback_data,omitempty"`
		Session               *CallSession    `json:"session,omitempty"`
		Connection            *Connection     `json:"connection,omitempty"`
		CallRecording         *CallRecording  `json:"call_recording,omitempty"`
		CallTranscript        *CallTranscript `json:"call_transcript,omitempty"`
	}

	CallSession struct {
		SDPType string `json:"sdp_type"`
		SDP     string `json:"sdp"`
	}

	// CallRecording wraps the recording payload delivered in a
	// call_recording_available webhook event. Audio uses [media.Info]
	// — the same type used for incoming media messages, since recordings
	// share the same download flow via the Media API.
	CallRecording struct {
		Type  string      `json:"type"`
		Audio *media.Info `json:"audio,omitempty"`
	}

	// CallTranscript wraps the transcription payload delivered in a
	// call_transcription_available webhook event. Document uses [media.Info]
	// — the same type used for incoming media messages. The downloaded
	// document is a JSON file with metadata, transcript text, language
	// detection, confidence scores, and word-level segments.
	CallTranscript struct {
		Document *media.Info `json:"document,omitempty"`
	}

	WebRTC struct {
		SDP string `json:"sdp"`
	}

	Connection struct {
		WebRTC *WebRTC `json:"webrtc,omitempty"`
	}
)

// CanHandleEvent reports whether a handler is registered for the
// call event carried by this NotificationEvent. It checks value.Statuses for
// call-type statuses and value.Calls for connect/created/terminate/recording
// events, returning true when the matching sub-handler is non-nil.
func (ch *CallsHandler) CanHandleEvent(event NotificationEvent) bool {
	if event.Value == nil {
		return false
	}
	for _, status := range event.Value.Statuses {
		if status != nil && status.Type == "call" {
			return ch.callsStatus != nil
		}
	}
	for _, call := range event.Value.Calls {
		if call == nil {
			continue
		}
		switch call.Event {
		case "connect":
			return ch.connect != nil
		case "call_created":
			return ch.created != nil
		case "terminate":
			return ch.terminate != nil
		case "call_recording_available":
			return ch.recordingAvailable != nil
		case "call_transcription_available":
			return ch.transcriptionAvailable != nil
		}
	}
	return false
}

var _ EventHandler = (*CallsHandler)(nil)

// OnCallConnect registers a handler for call connect events.
func (handler *Handler) OnCallConnect(h CallConnectHandler) {
	handler.calls.OnCallConnect(h)
}

// OnCallCreated registers a handler for call created events.
func (handler *Handler) OnCallCreated(h CallCreatedHandler) {
	handler.calls.OnCallCreated(h)
}

// OnCallTerminate registers a handler for call terminate events.
func (handler *Handler) OnCallTerminate(h CallTerminateHandler) {
	handler.calls.OnCallTerminate(h)
}

// OnCallStatus registers a handler for call status events (type "call").
func (handler *Handler) OnCallStatus(h CallStatusHandler) {
	handler.calls.OnCallStatus(h)
}

// OnCallRecordingAvailable registers a handler for call recording available events.
func (handler *Handler) OnCallRecordingAvailable(h CallRecordingAvailableHandler) {
	handler.calls.OnCallRecordingAvailable(h)
}

// OnCallTranscriptionAvailable registers a handler for call transcription
// available events. The handler receives the full Call payload with
// CallTranscript populated — use call_transcript.document.id with the Media
// API to download the transcript JSON document.
func (handler *Handler) OnCallTranscriptionAvailable(h CallTranscriptionAvailableHandler) {
	handler.calls.OnCallTranscriptionAvailable(h)
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
