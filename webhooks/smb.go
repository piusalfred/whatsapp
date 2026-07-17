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

// SMBMessageEchoesHandler for the "smb_message_echoes" webhook field.
// Handles messages sent by a business customer via the WhatsApp Business
// app or a companion ("linked") device.

package webhooks

import (
	"context"
	"fmt"
)

// SMBMessageEchoHandler is the interface for handling a single SMB message
// echo. Unlike [MessagesHandler], this does not perform per-type dispatch
// (text, image, revoke, edit, etc.). All echo messages are delivered to a
// single handler.
type SMBMessageEchoHandler interface {
	Handle(ctx context.Context, nctx *MessageNotificationContext, msg *Message) error
}

// SMBMessageEchoHandlerFunc adapts a bare function to the SMBMessageEchoHandler
// interface.
type SMBMessageEchoHandlerFunc func(ctx context.Context, nctx *MessageNotificationContext, msg *Message) error

// Handle implements [SMBMessageEchoHandler] by calling the underlying function.
func (f SMBMessageEchoHandlerFunc) Handle(ctx context.Context, nctx *MessageNotificationContext, msg *Message) error {
	return f(ctx, nctx, msg)
}

// SMBMessageEchoesHandler handles the "smb_message_echoes" webhook field.
//
// Messages sent via the WhatsApp Business app or companion devices by a
// business customer onboarded to Cloud API arrive through this field. The
// payload shape is identical to incoming messages — the same [Message]
// type is used — but the field is "smb_message_echoes" instead of "messages".
//
// Dispatch is intentionally simple: one handler for all echo message types
// (text, image, revoke, edit, etc.). There is no per-type sub-dispatch.
//
// # Concurrency
//
// SMBMessageEchoesHandler is safe for concurrent calls to
// [SMBMessageEchoesHandler.Handle] (read-only access to registered callbacks).
// It is not safe for concurrent modification — register all handlers before
// the handler starts serving requests. See [Handler] for the top-level
// concurrency contract.
//
// Usage:
//
//	sh := &SMBMessageEchoesHandler{}
//	sh.OnEcho(webhooks.SMBMessageEchoHandlerFunc(
//	    func(ctx context.Context, nctx *webhooks.MessageNotificationContext, msg *webhooks.Message) error {
//	        log.Printf("echo from %s: type=%s", msg.From, msg.Type)
//	        return nil
//	    },
//	))
type SMBMessageEchoesHandler struct {
	// Handler receives every SMB message echo regardless of type. When nil,
	// echoes fall through to [FallbackHandler].
	Handler SMBMessageEchoHandler

	// Fallback is called when Handler is nil or when the Handler returns an
	// error that the ErrorHandler considers non-fatal. When nil, the event
	// is silently acknowledged (HTTP 200).
	Fallback FallbackHandler

	// ErrorHandler is called when Handler returns an error. When nil, the
	// error is returned as-is (passthrough).
	ErrorHandler ErrorHandler
}

// OnEcho sets the handler for all SMB message echo messages.
func (sh *SMBMessageEchoesHandler) OnEcho(h SMBMessageEchoHandler) {
	sh.Handler = h
}

// OnFallback sets the catch-all handler for echoes when [Handler] is nil.
func (sh *SMBMessageEchoesHandler) OnFallback(h FallbackHandler) {
	sh.Fallback = h
}

// handleError routes an error through the SMBMessageEchoesHandler's
// ErrorHandler. When ErrorHandler is nil, the error is returned as-is.
func (sh *SMBMessageEchoesHandler) handleError(ctx context.Context, err error) error {
	return handleSubHandlerError(ctx, sh.ErrorHandler, err)
}

// executeFallback routes an unhandled echo through the Fallback catch-all.
// Returns nil when Fallback is nil (silent skip).
func (sh *SMBMessageEchoesHandler) executeFallback(ctx context.Context, event NotificationEvent) error {
	if sh.Fallback == nil {
		return nil
	}
	if err := sh.Fallback.Handle(ctx, event); err != nil {
		return fmt.Errorf("smb echo fallback: %w", err)
	}
	return nil
}

// Handle dispatches the smb_message_echoes value to the echo handler.
//
//  1. If [Handler] is set, each message in value.MessageEchoes is passed
//     to it. Errors are routed through [ErrorHandler].
//  2. If [Handler] is nil, falls back to [Fallback].
//  3. If [Fallback] is also nil, the event is silently skipped (HTTP 200).
func (sh *SMBMessageEchoesHandler) Handle(
	ctx context.Context,
	event NotificationEvent,
) error {
	if event.Value == nil {
		return nil
	}

	// No dedicated handler → fallback or silent skip.
	if sh.Handler == nil {
		return sh.executeFallback(ctx, event)
	}

	nctx := &MessageNotificationContext{
		EntryID:            event.EntryID,
		EntryTime:          event.Time,
		NotificationObject: event.Object,
		MessagingProduct:   event.Value.MessagingProduct,
		Contacts:           event.Value.Contacts,
		Metadata:           event.Value.Metadata,
	}

	for _, msg := range event.Value.MessageEchoes {
		if msg == nil {
			continue
		}
		if err := sh.Handler.Handle(ctx, nctx, msg); err != nil {
			return sh.handleError(ctx, fmt.Errorf("smb message echo: %w", err))
		}
	}

	return nil
}

// OnSMBMessageEcho sets the handler for SMB message echo messages. Each
// echo message is delivered to the handler regardless of its type (text,
// image, revoke, edit, etc.).
func (handler *Handler) OnSMBMessageEcho(h SMBMessageEchoHandler) {
	handler.ensureSMBEchoes().OnEcho(h)
}

// SMBAppStateSyncHandler is the interface for handling a single SMB app
// state sync entry (a contact add, edit, or remove from the WhatsApp
// Business app address book).
type SMBAppStateSyncHandler interface {
	Handle(ctx context.Context, nctx *MessageNotificationContext, sync *SMBAppStateSync) error
}

// SMBAppStateSyncHandlerFunc adapts a bare function to the
// SMBAppStateSyncHandler interface.
type SMBAppStateSyncHandlerFunc func(ctx context.Context, nctx *MessageNotificationContext, sync *SMBAppStateSync) error

// Handle implements [SMBAppStateSyncHandler] by calling the underlying function.
func (f SMBAppStateSyncHandlerFunc) Handle(
	ctx context.Context, nctx *MessageNotificationContext, sync *SMBAppStateSync,
) error {
	return f(ctx, nctx, sync)
}

// SMBAppStateSyncsHandler handles the "smb_app_state_sync" webhook field.
//
// This webhook synchronizes contacts from a WhatsApp Business app user who
// has been onboarded to Cloud API via a solution provider. Each state_sync
// entry describes a contact add, edit, or remove action.
//
// Dispatch is intentionally simple: one handler for all sync entries.
// There is no per-action sub-dispatch (add vs. remove).
//
// Usage:
//
//	sh := &SMBAppStateSyncsHandler{}
//	sh.OnSync(webhooks.SMBAppStateSyncHandlerFunc(
//	    func(ctx context.Context, nctx *webhooks.MessageNotificationContext, s *webhooks.SMBAppStateSync) error {
//	        log.Printf("contact %s: action=%s", s.Contact.FullName, s.Action)
//	        return nil
//	    },
//	))
type SMBAppStateSyncsHandler struct {
	// Handler receives every state sync entry regardless of action. When nil,
	// entries fall through to [FallbackHandler].
	Handler SMBAppStateSyncHandler

	// Fallback is called when Handler is nil. When nil, the event is silently
	// acknowledged (HTTP 200).
	Fallback FallbackHandler

	// ErrorHandler is called when Handler returns an error. When nil, the
	// error is returned as-is (passthrough).
	ErrorHandler ErrorHandler
}

// OnSync sets the handler for all SMB app state sync entries.
func (sh *SMBAppStateSyncsHandler) OnSync(h SMBAppStateSyncHandler) {
	sh.Handler = h
}

// OnFallback sets the catch-all handler for sync entries when [Handler] is nil.
func (sh *SMBAppStateSyncsHandler) OnFallback(h FallbackHandler) {
	sh.Fallback = h
}

func (sh *SMBAppStateSyncsHandler) handleError(ctx context.Context, err error) error {
	return handleSubHandlerError(ctx, sh.ErrorHandler, err)
}

func (sh *SMBAppStateSyncsHandler) executeFallback(ctx context.Context, event NotificationEvent) error {
	if sh.Fallback == nil {
		return nil
	}
	if err := sh.Fallback.Handle(ctx, event); err != nil {
		return fmt.Errorf("smb app state sync fallback: %w", err)
	}
	return nil
}

// Handle dispatches the smb_app_state_sync value to the sync handler.
//
//  1. If [Handler] is set, each entry in value.StateSync is passed to it.
//     Errors are routed through [ErrorHandler].
//  2. If [Handler] is nil, falls back to [Fallback].
//  3. If [Fallback] is also nil, the event is silently skipped (HTTP 200).
func (sh *SMBAppStateSyncsHandler) Handle(
	ctx context.Context,
	event NotificationEvent,
) error {
	if event.Value == nil {
		return nil
	}

	if sh.Handler == nil {
		return sh.executeFallback(ctx, event)
	}

	nctx := &MessageNotificationContext{
		EntryID:            event.EntryID,
		EntryTime:          event.Time,
		NotificationObject: event.Object,
		MessagingProduct:   event.Value.MessagingProduct,
		Contacts:           event.Value.Contacts,
		Metadata:           event.Value.Metadata,
	}

	for i := range event.Value.StateSync {
		if err := sh.Handler.Handle(ctx, nctx, &event.Value.StateSync[i]); err != nil {
			return sh.handleError(ctx, fmt.Errorf("smb app state sync: %w", err))
		}
	}

	return nil
}

// OnSMBAppStateSync sets the handler for SMB app state sync entries.
// Each entry (contact add, edit, or remove) is delivered to the handler.
func (handler *Handler) OnSMBAppStateSync(h SMBAppStateSyncHandler) {
	handler.ensureSMBAppSync().OnSync(h)
}
