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

package webhooks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"golang.org/x/sync/errgroup"
)

// NotificationEvent is a single flattened webhook event. It combines the
// Notification envelope metadata, the Entry-level identifiers, and one
// Change (field + value) into a single struct so callers can iterate over
// a flat event list without navigating the nested Notification → Entry →
// Change hierarchy.
type NotificationEvent struct {
	// Object is the top-level notification object, always "whatsapp_business_account".
	Object string
	// EntryID is the WhatsApp Business Account ID from the parent Entry.
	EntryID string
	// Time is the Unix timestamp set by the WhatsApp server on the parent Entry.
	Time int64
	// Field is the webhook change field ("messages", "flows",
	// "account_review_update", etc.) that triggered this event.
	Field string
	// Value holds the typed payload for this change. It may be nil when the
	// WhatsApp App Dashboard "Include Values" setting is disabled.
	Value *Value
}

// HandleNotificationEvent dispatches a single flattened NotificationEvent to
// the correct sub-handler based on event.Field. It builds a NotificationEntry
// and Change from the event and delegates to the shared dispatch core used by
// the native notification path.
//
// Use this when you've parsed a notification into events via
// [Notification.Events] and want to feed individual events through the handler
// pipeline (e.g., for selective async processing).
func (handler *Handler) HandleNotificationEvent(ctx context.Context, event NotificationEvent) error {
	if event.Value == nil {
		return nil
	}

	return handler.dispatchEvent(ctx, event)
}

// dispatchEvent is the unified dispatch core for all event routing. It maps
// the event field to a [EventHandler] and delegates to [HandleEvent].
// Unknown fields are routed to the general fallback.
func (handler *Handler) dispatchEvent(
	ctx context.Context,
	event NotificationEvent,
) error {
	_, isImplemented := handler.changeFieldHandlers.Check(event.Field)
	if !isImplemented {
		if handler.fallback != nil {
			if fErr := handler.fallback.Handle(ctx, event); fErr != nil {
				return fmt.Errorf("general fallback: %w", fErr)
			}
		}
		return nil
	}

	ch := handler.changeHandler(GetChangeFieldCategory(event.Field))
	if ch == nil {
		if handler.fallback != nil {
			if fErr := handler.fallback.Handle(ctx, event); fErr != nil {
				return fmt.Errorf("fallback: %w", fErr)
			}
		}
		return nil
	}

	return HandleEvent(ctx, ch, event)
}

// changeHandler returns the [EventHandler] for the given category, or nil
// if the category is unregistered.
func (handler *Handler) changeHandler(cfc ChangeFieldCategory) EventHandler {
	switch cfc {
	case ChangeFieldCategoryFlows:
		return handler.flows
	case ChangeFieldCategoryBusiness:
		return handler.business
	case ChangeFieldCategoryCalls:
		return handler.calls
	case ChangeFieldCategoryUserPreferences:
		return handler.userPrefs
	case ChangeFieldCategorySMBAppStateSync:
		return handler.smbAppSync
	case ChangeFieldCategoryMessages:
		return handler.messages
	case ChangeFieldCategorySMBMessageEchoes:
		return handler.smbEcho
	case ChangeFieldCategoryGroups:
		return handler.groups
	case ChangeFieldCategoryHistory:
		return handler.history
	default:
		return nil
	}
}

// Events flattens the Notification hierarchy into a slice of NotificationEvent.
// Each Change in each Entry produces one NotificationEvent carrying its parent
// metadata. The result is deterministic and preserves iteration order (entries
// then changes within each entry).
func (n *Notification) Events() []NotificationEvent {
	if n == nil {
		return nil
	}

	var events []NotificationEvent
	for _, entry := range n.Entry {
		for _, change := range entry.Changes {
			events = append(events, NotificationEvent{
				Object:  n.Object,
				EntryID: entry.ID,
				Time:    entry.Time,
				Field:   change.Field,
				Value:   change.Value,
			})
		}
	}
	return events
}

// HandleNotificationEvents processes an incoming WhatsApp webhook notification.
// It flattens the payload into [NotificationEvent] values and dispatches each
// concurrently to the correct sub-handler. All events are processed in parallel
// within each entry group. Returns a Response indicating success (200) or
// gateway timeout (504) if the context is cancelled.
func (handler *Handler) HandleNotificationEvents(ctx context.Context, notification *Notification) *Response {
	select {
	case <-ctx.Done():
		return &Response{StatusCode: http.StatusGatewayTimeout}
	default:
	}

	events := notification.Events()
	if len(events) == 0 {
		return &Response{StatusCode: http.StatusOK}
	}

	g, gContext := errgroup.WithContext(ctx)
	for _, event := range events {
		ev := event
		g.Go(func() error {
			return handler.HandleNotificationEvent(gContext, ev)
		})
	}
	if err := g.Wait(); err != nil {
		return &Response{StatusCode: http.StatusGatewayTimeout}
	}

	return &Response{StatusCode: http.StatusOK}
}

type EventHandler interface {
	Handle(ctx context.Context, event NotificationEvent) error
	HandleError(ctx context.Context, err error) error
	Fallback(ctx context.Context, event NotificationEvent) error
	IsEventHandlerImplemented(event NotificationEvent) bool
}

// HandleEvent dispatches a single event through the [EventHandler]
// pipeline:
//
//  1. If [EventHandler.IsEventHandlerImplemented] returns false, the event
//     is routed directly to [EventHandler.Fallback].
//  2. Otherwise, [EventHandler.Handle] is called. If Handle returns
//     [ErrEventNotHandled], the event is routed to [EventHandler.Fallback].
//  3. Any other error from Handle is routed through
//     [EventHandler.HandleError].
//  4. Panics in Handle, HandleError, or Fallback are recovered and wrapped
//     as [PanicError].
func HandleEvent(ctx context.Context, handler EventHandler, event NotificationEvent) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = &PanicError{Value: r, Stack: debug.Stack()}
		}
	}()

	if !handler.IsEventHandlerImplemented(event) {
		if fErr := handler.Fallback(ctx, event); fErr != nil {
			return fmt.Errorf("change handler fallback: %w", fErr)
		}
		return nil
	}

	if err = handler.Handle(ctx, event); err != nil {
		if errors.Is(err, ErrEventNotHandled) {
			if fErr := handler.Fallback(ctx, event); fErr != nil {
				return fmt.Errorf("change handler fallback: %w", fErr)
			}
			return nil
		}
		if hErr := handler.HandleError(ctx, err); hErr != nil {
			return fmt.Errorf("change handler error: %w", hErr)
		}
		return nil
	}

	return nil
}
