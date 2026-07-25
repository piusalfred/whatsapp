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

// History types and HistoryHandler for WhatsApp history sync webhooks.
// Models history entries with phase/chunk/progress metadata and threaded
// messages. Includes async processing warnings per the official API docs.
//
// History webhooks arrive in two forms (both with "field": "history"):
//
//  1. Chat history entries — contains a "history" array with metadata
//     (phase, chunk_order, progress), threads of messages, and optional
//     errors when sharing is declined.
//
//  2. Media content — contains a "messages" array with full message
//     payloads (including media asset IDs) for messages that originally
//     appeared as "media_placeholder" type in the history threads.

package webhooks

import (
	"context"
	"fmt"
)

// HistoryEntry represents a single chat history webhook payload. Each entry
// contains metadata about the sync phase and progress, plus a set of message
// threads. Threads are delivered in chunks identified by chunk_order;
// use phase and progress to track overall sync completion (progress == 100
// means all history has been delivered).
//
// When sharing is declined, the Errors field is non-nil and Threads is empty.
type HistoryEntry struct {
	Metadata HistoryMetadata `json:"metadata"`
	Threads  []HistoryThread `json:"threads,omitempty"`
	Errors   []ErrorInfo     `json:"errors,omitempty"`
}

// HistoryMetadata describes the current sync phase and progress.
//
// Phase values:
//
//	0 — messages from day 0 (onboarding) through day 1
//	1 — messages from day 1 through day 90
//	2 — messages from day 90 through day 180
//
// Progress ranges 0–100. A value of 100 indicates sync is complete.
// ChunkOrder is a sequential index; chunks may arrive out of order.
type HistoryMetadata struct {
	Phase      int `json:"phase"`
	ChunkOrder int `json:"chunk_order"`
	Progress   int `json:"progress"`
}

// HistoryThread groups messages exchanged with a single WhatsApp user.
// The ID is the user's phone number. Messages use the standard [Message]
// type — history messages carry a [HistoryContext] with delivery status.
type HistoryThread struct {
	ID       string     `json:"id"`
	Messages []*Message `json:"messages"`
}

// HistoryContext provides the delivery status of a message in history.
// Status values: DELIVERED, ERROR, PENDING, PLAYED, READ, SENT.
type HistoryContext struct {
	Status string `json:"status"`
}

// HistoryHandler handles both forms of the history sync webhook.
//
//   - Chat history entries (threads of messages with delivery status) are
//     dispatched via [OnMessages]. A single webhook can carry thousands of
//     messages. Do NOT process synchronously — persist to a queue immediately.
//
//   - Media content webhooks containing full message payloads (including
//     media asset IDs) for messages that originally appeared as
//     "media_placeholder" in history threads are dispatched via
//     [OnMediaMessages]. These arrive under the same "history" field but
//     contain a "messages" array instead of a "history" array.
//
//   - Events that don't match either form are routed to the [Fallback] method.
//
// Usage:
//
//	hh := &HistoryHandler{}
//	hh.OnMessages(myHistoryHandler)
//	hh.OnMediaMessages(myMediaHandler)
type HistoryHandler struct {
	// message handles chat history entries (threads with delivery status).
	message ChangeValueHandler[HistoryEntry]

	// media handles media content webhooks that arrive under
	// the "history" field with a "messages" array. These carry full
	// message payloads (including media asset IDs) for messages that
	// appeared as "media_placeholder" in the history threads.
	media MessageHandler[Message]

	// fallback catches history webhooks that don't match either form.
	// When nil, unrecognized payloads are silently skipped.
	fallback FallbackHandler

	error ErrorHandler
}

var _ EventHandler = (*HistoryHandler)(nil)

// OnMessages sets the handler for chat history entries.
//
// The callback receives the full []*HistoryEntry array for the current chunk.
// Do NOT process synchronously — a single webhook can contain thousands of
// messages and will cause a timeout. Persist to a queue and process
// asynchronously.
func (hh *HistoryHandler) OnMessages(h ChangeValueHandler[HistoryEntry]) {
	hh.message = h
}

// OnMediaMessages sets the handler for media content webhooks that arrive
// under the "history" field. These carry full message payloads (including
// media asset IDs) for messages that appeared as "media_placeholder" in
// the history threads.
func (hh *HistoryHandler) OnMediaMessages(h MessageHandler[Message]) {
	hh.media = h
}

// OnFallback sets the catch-all handler for history webhooks without a
// dedicated handler.
func (hh *HistoryHandler) OnFallback(h FallbackHandler) {
	hh.fallback = h
}

// OnError sets the error handler for this domain handler. When nil, errors
// bubble up to the general error handler configured on [Handler].
func (hh *HistoryHandler) OnError(h ErrorHandler) {
	hh.error = h
}

// HandleError routes an error through the HistoryHandler's ErrorHandler.
// When the dedicated error handler is nil, the error is returned as-is.
func (hh *HistoryHandler) HandleError(ctx context.Context, err error) error {
	return execErrorHandler(ctx, hh.error, err)
}

// Fallback routes an unhandled history event through the fallback
// catch-all. Returns nil when fallback is nil (silent skip).
func (hh *HistoryHandler) Fallback(ctx context.Context, event NotificationEvent) error {
	if hh.fallback == nil {
		return nil
	}
	if err := hh.fallback.Handle(ctx, event); err != nil {
		return fmt.Errorf("history fallback: %w", err)
	}
	return nil
}

// CanHandleEvent reports whether a handler is registered for the
// specific event form carried by this NotificationEvent. It returns true when
// event.Value.History is non-empty and a messages handler is set, or when
// event.Value.Messages is non-empty and a media handler is set.
func (hh *HistoryHandler) CanHandleEvent(event NotificationEvent) bool {
	if len(event.Value.History) > 0 {
		return hh.message != nil
	}
	if len(event.Value.Messages) > 0 {
		return hh.media != nil
	}
	return false
}

// Handle dispatches the history webhook to the correct handler.
//
// Two forms are recognized:
//
//  1. event.Value.History contains entries → dispatched via [OnMessages].
//  2. event.Value.Messages is populated → dispatched via [OnMediaMessages].
//
// Returns [ErrEventNotHandled] when no handler is registered for the matching
// form, signalling the caller (typically [HandleEvent]) to invoke the
// [Fallback] method.
func (hh *HistoryHandler) Handle(
	ctx context.Context,
	event NotificationEvent,
) error {
	nctx := &MessageNotificationContext{
		EntryID:            event.EntryID,
		EntryTime:          event.Time,
		NotificationObject: event.Object,
		MessagingProduct:   event.Value.MessagingProduct,
		Metadata:           event.Value.Metadata,
	}

	// Form 1: chat history entries (threads of messages).
	if len(event.Value.History) > 0 {
		if hh.message == nil {
			return ErrEventNotHandled
		}
		entries := make([]*HistoryEntry, len(event.Value.History))
		for i := range event.Value.History {
			entries[i] = &event.Value.History[i]
		}
		req := &ChangeValueRequest[HistoryEntry]{
			Notification: nctx,
			Payload:      entries,
		}
		if err := hh.message.Handle(ctx, req); err != nil {
			return hh.HandleError(ctx, err)
		}
		return nil
	}

	// Form 2: media content for history messages (contains "messages" array
	// with full payload including media asset IDs).
	if len(event.Value.Messages) > 0 {
		if hh.media == nil {
			return ErrEventNotHandled
		}
		for _, msg := range event.Value.Messages {
			if msg == nil {
				continue
			}
			if err := hh.media.Handle(ctx, &MessageRequest[Message]{
				Notification: nctx,
				Info:         newMessageInfo(msg),
				Payload:      msg,
			}); err != nil {
				return hh.HandleError(ctx, err)
			}
		}
		return nil
	}

	return ErrEventNotHandled
}

// OnHistorySync registers a handler for chat history entries (threads of
// messages with delivery status).
func (handler *Handler) OnHistorySync(h ChangeValueHandler[HistoryEntry]) {
	handler.history.OnMessages(h)
}

// OnHistoryMediaMessages registers a handler for media content webhooks
// that arrive under the "history" field. These carry full message payloads
// (including media asset IDs) for messages that appeared as
// "media_placeholder" in the history threads.
func (handler *Handler) OnHistoryMediaMessages(h MessageHandler[Message]) {
	handler.history.OnMediaMessages(h)
}

// OnHistoryFallback registers a catch-all handler for history webhooks
// that don't match either chat history entries or media content.
func (handler *Handler) OnHistoryFallback(h FallbackHandler) {
	handler.history.OnFallback(h)
}
