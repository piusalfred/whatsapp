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

package webhooks

import (
	"context"
	"fmt"
)

// HistorySyncContext carries the notification-level metadata for a history
// sync webhook. It is passed to the history handler callback.
type HistorySyncContext struct {
	NotificationObject string // "whatsapp_business_account"
	EntryID            string // WABA ID
	EntryTime          int64  // UNIX timestamp
}

// HistoryEntry represents a single chat history webhook payload. Each entry
// contains metadata about the sync phase and progress, plus a set of message
// threads. Threads are delivered in chunks identified by chunk_order;
// use phase and progress to track overall sync completion (progress == 100
// means all history has been delivered).
type HistoryEntry struct {
	Metadata HistoryMetadata `json:"metadata"`
	Threads  []HistoryThread `json:"threads,omitempty"`
	Errors   []ErrorInfo     `json:"errors,omitempty"` // non-nil when sharing is declined
}

// HistoryMetadata describes the current sync phase and progress.
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

// HistoryHandler handles history sync webhooks.
//
// WARNING: A single webhook can carry thousands of messages. Do NOT process
// the payload synchronously — it will cause a timeout and WhatsApp will
// retry. Capture the raw payload, persist it immediately to a queue or
// staging store, and process on a background worker.
//
// The callback receives the full []*HistoryEntry array for the current
// chunk. Each entry contains sync metadata (phase, chunk_order, progress)
// and threads of messages with delivery status. Progress reaches 100 when
// sync is complete. Media messages appear as "media_placeholder" type;
// actual media content arrives as a separate webhook routed through the
// standard message handler.
//
// Usage:
//
//	hh := &HistoryHandler{}
//	hh.OnMessages(webhooks.ChangeValueHandlerFunc[webhooks.HistoryEntry](
//		func(ctx context.Context, req *webhooks.ChangeValueRequest[webhooks.HistoryEntry]) error {
//		    // Don't process here — persist to queue immediately.
//		    for _, e := range req.Payload {
//		        queue.Push(e)
//		    }
//		    return nil
//		},
//	))
type HistoryHandler struct {
	Messages ChangeValueHandler[HistoryEntry]
}

// OnMessages sets the handler for history sync webhooks.
//
// The callback receives the full []*HistoryEntry array for the current chunk.
// Do NOT process synchronously — a single webhook can contain thousands of
// messages and will cause a timeout. Persist to a queue and process
// asynchronously.
func (hh *HistoryHandler) OnMessages(h ChangeValueHandler[HistoryEntry]) {
	hh.Messages = h
}

// HandleHistoryMessages dispatches the history entries to the configured
// handler. Returns nil when no handler is set.
//
// WARNING: Do not process the payload in this handler. A single webhook can
// carry thousands of messages and will cause a timeout. The handler should
// persist the data immediately and return — processing should happen
// asynchronously on a background worker.
//
// Errors from the user handler are routed through onError; if onError
// returns a non-nil error, processing stops.
//
// History media content webhooks (change.Value.Messages populated under the
// "history" field) are handled separately by the standard message handler —
// they are not processed here.
func (hh *HistoryHandler) HandleHistoryMessages(
	ctx context.Context,
	ne NotificationEntry,
	change Change,
	onError ErrorHandler,
) error {
	if len(change.Value.History) == 0 || hh.Messages == nil {
		return nil
	}

	nctx := &MessageNotificationContext{
		EntryID:            ne.ID,
		EntryTime:          ne.Time,
		NotificationObject: ne.Object,
		MessagingProduct:   change.Value.MessagingProduct,
		Metadata:           change.Value.Metadata,
	}

	entries := make([]*HistoryEntry, len(change.Value.History))
	for i := range change.Value.History {
		entries[i] = &change.Value.History[i]
	}

	req := &ChangeValueRequest[HistoryEntry]{
		Notification: nctx,
		Payload:      entries,
	}
	if err := hh.Messages.Handle(ctx, req); err != nil {
		if onError != nil {
			if handlerErr := onError.Handle(ctx, err); handlerErr != nil {
				return fmt.Errorf("error handler: %w", handlerErr)
			}
		}
	}

	return nil
}
